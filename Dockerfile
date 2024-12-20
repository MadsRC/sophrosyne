FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:1e61f5abe1233356c3e63ab4e050b846cc0ffcb5e6516da8e74cd81982bd9328
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
