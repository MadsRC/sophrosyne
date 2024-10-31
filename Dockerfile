FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:0c09bcfc6a1f8755b7a20bd7550e0448adc75d75d22baddd57d9b87577d3f8b4
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
