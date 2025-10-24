FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:dfe55497b1b74855e14b27f1710bf9658ebca69cbebe00e5370ab3bf6da2f9d1
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
