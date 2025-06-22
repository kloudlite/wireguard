FROM alpine:latest

RUN apk add --no-cache wireguard-tools-wg-quick iputils iptables procps libcap curl nmap-ncat dog

RUN mkdir -p /etc/wireguard
WORKDIR /etc/wireguard

RUN <<EOF
# for interoperability with docker.io/linuxserver/wireguard image
  mkdir -p /etc/wireguard /config
  ln -sf /etc/wireguard /config/wg_confs 
EOF

# Read more about [udp2raw](https://www.procustodibus.com/blog/2022/02/wireguard-over-tcp/)
# ARG TARGETARCH
# RUN <<EOF
#   mkdir -p /tmp/x
#   cd /tmp/x
#   curl -L0 https://github.com/wangyu-/udp2raw/releases/download/20230206.0/udp2raw_binaries.tar.gz > b.t.gz
#   tar xf b.t.gz
#   mv udp2raw_$TARGETARCH /usr/local/bin/udp2raw
#   rm -rf /tmp/x
# EOF
