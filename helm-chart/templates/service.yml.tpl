---
apiVersion: v1
kind: Service
metadata:
  name: {{.Release.Name}}-controller-manager-metrics-service
  namespace: {{.Release.Namespace}}
spec:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: 8443
  selector:
    app.kubernetes.io/name: wireguard
    control-plane: controller-manager
