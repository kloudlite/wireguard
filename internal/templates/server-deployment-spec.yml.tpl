spec:
  replicas: 1
  selector:
    matchLabels: {{.PodLabels | toJson}}
  template:
    metadata:
      labels: {{.PodLabels | toJson}}
    spec:
      securityContext:
        sysctls:
          - name: net.ipv4.ip_forward
            value: "1"

      containers:
        - name: wireguard
          image: ghcr.io/kloudlite/wireguard/images/wireguard:latest
          imagePullPolicy: Always
          command:
            - sh
            - -c
            - |+
              cat > /etc/wireguard/wg0.conf <<'EOF'
              {{ .Wg0Conf | nindent 14 }}
              EOF

              wg-quick down wg0 || echo "starting wg0"
              wg-quick up wg0

              tail -f /dev/null &
              pid=$!
              trap 'kill -9 $pid' EXIT INT TERM
              wait $pid
          ports:
            - containerPort: 51820
              protocol: UDP
          resources:
            requests:
              cpu: 50m
              memory: 50Mi
            limits:
              cpu: 100m
              memory: 100Mi
          securityContext:
            sysctls:
              - name: net.ipv4.ip_forward
                value: "1"
            capabilities:
              add:
                - NET_ADMIN
              # drop:
              #   - all

        - name: dns
          image: ghcr.io/nxtcoder17/simple-dns:master-nightly
          imagePullPolicy: Always
          args:
            - --addr
            - ":53"
            - --debug
            - --upstream
            - svc.cluster.local={{.KubeDNSSvcIP}}
            {{- range $_, $item := .DNSLocalhosts }}
            - --wildcard-host
            - {{$item}}=$(NODE_IP)
            {{- end }}
          env:
            - name: NODE_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.hostIP
          ports:
            - containerPort: 53
              protocol: UDP
          resources:
            requests:
              cpu: 20m
              memory: 20Mi
            limits:
              cpu: 40m
              memory: 40Mi
