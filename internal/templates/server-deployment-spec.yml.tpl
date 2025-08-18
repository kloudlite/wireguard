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
          image: {{.WgServerImage}}
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

        {{- with .WgDNSTemplateParams }}
        - name: dns
          image: {{.SimpleDNSServerImage}}
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
        {{- end }}

        {{- if .PortMappings }}
        {{- with .WgProxyTemplateParams }}
        - name: wg-proxy
          image: ghcr.io/kloudlite/hub/socat:latest
          command:
            - sh
            - -c
            - |+
              {{- range $_, $v := .PortMappings }}

              {{- if eq $v.Protocol "TCP" }}
              (socat -dd tcp4-listen:{{$v.Port}},fork,reuseaddr tcp4:{{$v.TargetHost}}:{{$v.TargetPort}} 2>&1 | grep -iE --line-buffered 'listening|exiting') &
              pid="$pid $!"
              {{- else if eq $v.Protocol "UDP" }}
              (socat -dd UDP4-LISTEN:{{$v.Port}},fork,reuseaddr UDP4:{{$v.TargetHost}}:{{$v.TargetPort}} 2>&1 | grep -iE --line-buffered 'listening|exiting') &
              pid="$pid $!"
              {{- end }}

              {{- end }}

              trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM
              eval wait $pid
          securityContext:
            capabilities:
              add:
                - NET_BIND_SERVICE
                - SETGID
              drop:
                - all
          resources:
            limits:
              memory: "50Mi"
              cpu: "50m"
            requests:
              memory: "50Mi"
              cpu: "50m"
        {{- end }}
        {{- end }}
