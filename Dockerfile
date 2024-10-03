FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:7fa43737be034509a394129d6966fbb7fbb6cbc01f34ec03486f4acd5d657edc
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
