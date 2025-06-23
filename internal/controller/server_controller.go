/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	rApi "github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	v1 "github.com/kloudlite/wireguard/api/v1"
	"github.com/kloudlite/wireguard/internal/templates"
	"github.com/seancfoley/ipaddress-go/ipaddr"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	*Env

	YAMLClient kubectl.YAMLClient

	templateServerSetup  []byte
	templateServerConfig []byte

	templateServerDeploymentSpec []byte
	templateServerServiceSpec    []byte

	templatePeerConf []byte
}

// GetName implements reconciler.Reconciler.
func (r *ServerReconciler) GetName() string {
	return "wireguard-server"
}

// +kubebuilder:rbac:groups=wireguard.kloudlite.github.com,resources=servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.github.com,resources=servers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.github.com,resources=servers/finalizers,verbs=update
// +kubebuilder:rbac:groups=,resources=namespaces,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=services,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups=apps/v1,resources=deployments,verbs=get;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=secrets,verbs=get;create;update;patch;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *ServerReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, request.NamespacedName, &v1.Server{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if result, err := reconciler.ReconcileSteps(req, []rApi.Step[*v1.Server]{
		{
			Name:     "namespace",
			Title:    "Kubernetes Namespace for wg server",
			OnCreate: r.CreateNamespace,
			OnDelete: r.cleanupNamespace,
		},
		{
			Name:     "generate-wg-server-keys",
			Title:    "Generates Wireguard Server Keys",
			OnCreate: r.CreateWireguardServerKeys,
			OnDelete: nil,
		},
		{
			Name:     "sync-peers",
			Title:    "Sync All Listed Peers",
			OnCreate: r.syncPeers,
			OnDelete: r.cleanupPeers,
		},
		{
			Name:     "setup-wireguard-server-deployment",
			Title:    "Setup Wireguard Server Deployment",
			OnCreate: r.createDeployment,
			OnDelete: r.cleanupDeployment,
		},
		{
			Name:     "setup-wireguard-server-service",
			Title:    "Setup Wireguard Server Service",
			OnCreate: r.createService,
			OnDelete: r.cleanupService,
		},
	}); err != nil {
		return result, err
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func generateWgKeys() (privateKey, publicKey string, err error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return "", "", err
	}

	return key.String(), key.PublicKey().String(), nil
}

func GenIPAddr(cidr string, offset int) (string, error) {
	deviceRange := ipaddr.NewIPAddressString(cidr)

	address, err := deviceRange.ToAddress()
	if err != nil {
		return "", err
	}

	increment := address.Increment(int64(offset))
	if ok := deviceRange.Contains(increment.ToAddressString()); !ok {
		return "", fmt.Errorf("IP Addresses MaxedOut in this CIDR (%s)", cidr)
	}

	return ipaddr.NewIPAddressString(increment.GetNetIP().String()).String(), nil
}

func pickFirstAvailableIP(cidr string, ipMap map[string]struct{}) (string, error) {
	for i := 2; ; i++ {
		ip, err := GenIPAddr(cidr, i)
		if err != nil {
			return "", err
		}

		if _, ok := ipMap[ip]; ok {
			continue
		}

		return ip, nil
	}
}

func (r *ServerReconciler) CreateNamespace(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	if obj.Spec.TargetNamespace == "" {
		obj.Spec.TargetNamespace = "wg-" + obj.Name
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.Abort()
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, ns, func() error {
		fn.MapSet(&ns.Annotations, "kloudlite.io/description", fmt.Sprintf("Managed By Wireguard Controller. It is created to store deployments and configurations related to wireguard server (%s)", obj.Name))
		return nil
	}); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *ServerReconciler) cleanupNamespace(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *ServerReconciler) CreateWireguardServerKeys(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	if obj.Spec.PrivateKey == nil || obj.Spec.PublicKey == nil {
		privateKey, publicKey, err := generateWgKeys()
		if err != nil {
			return check.Failed(err)
		}
		obj.Spec.PrivateKey = &privateKey
		obj.Spec.PublicKey = &publicKey

		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
	}

	return check.Passed()
}

func (r *ServerReconciler) createDeployment(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	wg0Config, err := templates.ParseBytes(r.templateServerConfig, templates.ParamsWgServerConf{
		ServerIP:         *obj.Spec.IP,
		ServerPrivateKey: *obj.Spec.PrivateKey,
		PodCIDR:          r.Env.PodCIDR,
		Peers:            obj.Spec.Peers,
	})
	if err != nil {
		return check.Failed(err)
	}

	kubeDNSSvc, err := rApi.Get(check.Context(), r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	b, err := templates.ParseBytes(r.templateServerDeploymentSpec, templates.ParamsServerDeploymentSpec{
		PodLabels:     map[string]string{"app": obj.Name},
		Wg0Conf:       string(wg0Config),
		KubeDNSSvcIP:  kubeDNSSvc.Spec.ClusterIP,
		DNSLocalhosts: obj.Spec.DNS.Localhosts,
	})
	if err != nil {
		return check.Failed(err)
	}

	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Spec.TargetNamespace}}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, deployment, func() error {
		deployment.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		return yaml.Unmarshal(b, &deployment)
	}); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *ServerReconciler) createService(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	b, err := templates.ParseBytes(r.templateServerServiceSpec, templates.ParamsServerServiceSpec{
		SelectorLabels: map[string]string{"app": obj.Name},
		ServiceType:    obj.Spec.Expose.ServiceType,
		Port:           obj.Spec.Expose.Port,
	})
	if err != nil {
		return check.Failed(err)
	}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svc, func() error {
		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		return yaml.Unmarshal(b, &svc)
	}); err != nil {
		fmt.Printf("\nYAML:\n%s\n", b)
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *ServerReconciler) cleanupDeployment(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Spec.TargetNamespace}}); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *ServerReconciler) cleanupService(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Spec.TargetNamespace}}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *ServerReconciler) syncPeers(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	ipMap := map[string]struct{}{
		*obj.Spec.IP: {},
	}

	for _, peer := range obj.Spec.Peers {
		if peer.IP != nil {
			ipMap[*peer.IP] = struct{}{}
		}
	}

	for i := range obj.Spec.Peers {
		if obj.Spec.Peers[i].IP == nil {
			ip, err := pickFirstAvailableIP(*obj.Spec.CIDR, ipMap)
			if err != nil {
				return check.Failed(err)
			}
			obj.Spec.Peers[i].IP = &ip
			if err := r.Update(check.Context(), obj); err != nil {
				return check.Failed(err)
			}
			return check.Abort()
		}

		if obj.Spec.Peers[i].PrivateKey == nil || obj.Spec.Peers[i].PublicKey == nil {
			privateKey, publicKey, err := generateWgKeys()
			if err != nil {
				return check.Failed(err)
			}

			obj.Spec.Peers[i].PrivateKey = &privateKey
			obj.Spec.Peers[i].PublicKey = &publicKey
			if err := r.Update(check.Context(), obj); err != nil {
				return check.Failed(err)
			}
			return check.Abort()
		}

		if obj.Spec.Peers[i].AllowedIPs == nil {
			obj.Spec.Peers[i].AllowedIPs = []string{*obj.Spec.CIDR, r.Env.PodCIDR, r.Env.SvcCIDR}
			if err := r.Update(check.Context(), obj); err != nil {
				return check.Failed(err)
			}
			return check.Abort()
		}
	}

	for i := range obj.Spec.Peers {
		peer := obj.Spec.Peers[i]
		b, err := templates.ParseBytes(r.templatePeerConf, templates.ParamsWgPeerConf{
			Name:          peer.Name,
			IP:            *peer.IP,
			PrivateKey:    *peer.PrivateKey,
			DNS:           *obj.Spec.IP,
			DNSLocalhosts: obj.Spec.DNS.Localhosts,
			Peers: append(obj.Spec.Peers, v1.Peer{
				Name:       "server-" + obj.Name,
				IP:         obj.Spec.IP,
				PrivateKey: obj.Spec.PrivateKey,
				PublicKey:  obj.Spec.PublicKey,
				AllowedIPs: peer.AllowedIPs,
				Endpoint:   &obj.Spec.Endpoint,
			}),
		})
		if err != nil {
			return check.Failed(err)
		}

		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "wg-" + obj.Name + "-" + peer.Name, Namespace: obj.Spec.TargetNamespace}}
		if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, secret, func() error {
			secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
			if secret.Data == nil {
				secret.Data = make(map[string][]byte, 1)
			}
			secret.Data["wg.conf"] = b
			return nil
		}); err != nil {
			return check.Failed(err)
		}
	}

	return check.Passed()
}

func (r *ServerReconciler) cleanupPeers(check *reconciler.Check[*v1.Server], obj *v1.Server) reconciler.StepResult {
	for _, peer := range obj.Spec.Peers {
		if err := r.Delete(check.Context(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "wg-" + obj.Name + "-" + peer.Name, Namespace: obj.Spec.TargetNamespace}}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}
	}

	return check.Passed()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}

	if r.YAMLClient == nil {
		return fmt.Errorf("yamlclient must be set")
	}

	if r.Env == nil {
		return fmt.Errorf("env must be set")
	}

	var err error

	r.templateServerDeploymentSpec, err = templates.Read(templates.ServerDeploymentSpec)
	if err != nil {
		return err
	}

	r.templateServerServiceSpec, err = templates.Read(templates.ServerServiceSpec)
	if err != nil {
		return err
	}

	r.templateServerConfig, err = templates.Read(templates.WgServerConf)
	if err != nil {
		return err
	}

	r.templatePeerConf, err = templates.Read(templates.WgPeerConf)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Server{}).Named("wireguard:server")
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&corev1.Secret{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: 1})
	builder.WithEventFilter(rApi.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))

	return builder.Complete(r)
}
