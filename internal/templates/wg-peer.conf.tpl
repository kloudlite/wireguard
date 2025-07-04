[Interface]
# NAME: {{.Name}}
Address = {{.IP}}/32
PrivateKey = {{.PrivateKey}}
{{- if .DNS }}
{{- /* DNS = {{.DNS}} */}}
PostUp = [ -x /usr/bin/resolvectl ] && resolvectl dns %i {{.DNS}} && resolvectl domain %i 'svc.cluster.local' {{- range $_, $host := .DNSLocalhosts }} ~'{{$host}}' {{- end }}
PreDown = [ -x /usr/bin/resolvectl ] && resolvectl revert %i
{{- end }}

{{- range $_, $peer := .Peers }}
{{- if not (eq $peer.Name $.Name) }}
[Peer]
# NAME: {{$peer.Name}}
PublicKey = {{$peer.PublicKey}}
AllowedIPs = {{$peer.AllowedIPs | join ", " }}
{{- if $peer.Endpoint }}
Endpoint = {{$peer.Endpoint}}
{{- end }}
{{- end }}
{{ end }}
