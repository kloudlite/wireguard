FROM ghcr.io/nxtcoder17/nix AS builder
WORKDIR /app

RUN --mount=type=bind,source=flake.nix,target=flake.nix \
  --mount=type=bind,source=flake.lock,target=flake.lock \
  <<EOF
nix develop --verbose --command echo "nix setup complete"
EOF

ARG GOMODCACHE=/cache/gomodcache
ARG GOCACHE=/cache/gocache

ENV GOMODCACHE=${GOMODCACHE}
ENV GOCACHE=${GOCACHE}
ENV CGO_ENABLED=0

RUN --mount=type=bind,source=flake.nix,target=flake.nix \
  --mount=type=bind,source=flake.lock,target=flake.lock \
  --mount=type=bind,source=go.mod,target=go.mod \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=cache,target=$GOMODCACHE \
  --mount=type=cache,target=$GOCACHE \
  <<EOF
time nix develop --command go mod download -x -json
echo "DOWNLOADED go modules"
EOF

RUN --mount=type=bind,source=.,target=/app \
  --mount=type=cache,target=$GOMODCACHE \
  --mount=type=cache,target=$GOCACHE \
  <<EOF

time nix develop --command go build -v -ldflags='-s -w' -o /out/wireguard-controller ./cmd/
echo "BUILT binary"
EOF

FROM gcr.io/distroless/static:nonroot
WORKDIR /home/nonroot
COPY --from=builder --chown=nonroot:nonroot /out/wireguard-controller ./
USER 65532:65532
ENTRYPOINT ["./wireguard-controller"]

