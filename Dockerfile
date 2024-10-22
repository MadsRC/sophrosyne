FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:90c3d6a7b820f594e7a65cc4105c5a5e2203496fbd1d768f016bbf1b7fab0be6
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
