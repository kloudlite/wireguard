---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels: {{.Values.podLabels | toJson }}
  name: {{.Release.Name}}-controller
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels: {{ .Values.podLabels | toJson }}
  template:
    metadata:
      labels: {{ .Values.podLabels | toJson }}
    spec:
      containers:
      - name: manager
        args:
        - --metrics-bind-address=:8443
        - --leader-elect
        - --health-probe-bind-address=:8081
        image: "{{.Values.image.repository}}:{{.Values.image.tag}}"
        imagePullPolicy: {{.Values.image.pullPolicy | default "IfNotPresent"}}
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        ports: []
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
        resources: {{.Values.resources | toJson }}
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
        volumeMounts: []
      securityContext:
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      serviceAccountName: {{.Release.Name}}-controller-manager
      terminationGracePeriodSeconds: 10
      volumes: []

