---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{.Release.Name}}-leader-election-rolebinding
  namespace: {{.Release.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{.Release.Name}}-leader-election-role
subjects:
- kind: ServiceAccount
  name: {{.Release.Name}}-controller-manager
  namespace: {{.Release.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Name}}-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{.Release.Name}}-manager-role
subjects:
- kind: ServiceAccount
  name: {{.Release.Name}}-controller-manager
  namespace: {{.Release.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Release.Name}}-metrics-auth-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{.Release.Name}}-metrics-auth-role
subjects:
- kind: ServiceAccount
  name: {{.Release.Name}}-controller-manager
  namespace: {{.Release.Namespace}}
