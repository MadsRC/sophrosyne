FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:09c74721beb055d94e49515381cc5118d85bae208e7ddefe1d28f6b355a1a6f5
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
