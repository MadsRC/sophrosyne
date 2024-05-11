FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:6dff3d81e2edaa0ef48ea87b808c34c4b24959169d9ad317333bdda4ec3c4002
ARG TARGETARCH
USER nonroot
COPY dist/sophrosyne_linux_$TARGETARCH /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
