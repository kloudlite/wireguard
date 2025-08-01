[Interface]
Address = {{.ServerIP}}/32
ListenPort = 51820
PrivateKey = {{.ServerPrivateKey}}

# Enable IP forwarding and set up NAT rules
PostUp = iptables -A FORWARD -i %i -j ACCEPT;
PostUp = iptables -A FORWARD -o %i -j ACCEPT; 
PostUp = iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE;

PostDown = iptables -D FORWARD -i %i -j ACCEPT;
PostDown = iptables -D FORWARD -o %i -j ACCEPT; 
PostDown = iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE;

{{- range $_, $peer := .Peers }}
[Peer]
# NAME: {{$peer.Name}}
PublicKey = {{ $peer.PublicKey }}
AllowedIPs = {{ $peer.IP }}/32
{{- if $peer.Endpoint }}
Endpoint = {{$peer.Endpoint}}
{{- if gt $.KeepAlive 0 }}
PersistentKeepalive = {{$.KeepAlive}}
{{- end }}
{{- end }}
{{- end}}

