apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Release.Name}}-controller-manager
  namespace: {{.Release.Namespace}}
