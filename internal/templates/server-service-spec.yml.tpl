spec:
  selector: {{.SelectorLabels | toJson }}
  type: {{.ServiceType}}
  ports:
    - name: wg-udp
      protocol: UDP
      port: {{.Port}}
      targetPort: 51820
      {{- if eq .ServiceType "NodePort" }}
      nodePort: {{.Port}}
      {{- end }}

    - name: wg-dns
      protocol: UDP
      port: 53
      targetPort: 53
