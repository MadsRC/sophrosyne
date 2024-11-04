FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/glibc-dynamic:latest@sha256:1cfd1bd060c8c46e60778f8e8aa3664e335a57e661e97a448a455454c128851e
USER nonroot
COPY sophrosyne /usr/bin/sophrosyne
ENTRYPOINT ["/usr/bin/sophrosyne"]
