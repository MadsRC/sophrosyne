FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:7cbd1cd7a6a1ca37de8d242ade7e65253323d1991611ebcb20b15347a95b8ebf
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
