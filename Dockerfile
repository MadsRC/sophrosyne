FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:17f942295bb0ba9c1d27c06382d4a999bc8becc5cf6bbcbde0af0baa00b9b470
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
