FROM cgr.dev/chainguard/glibc-dynamic:latest@sha256:6dff3d81e2edaa0ef48ea87b808c34c4b24959169d9ad317333bdda4ec3c4002
COPY --chown=noneroot:noneroot dist/sophrosyne /usr/bin/
ENTRYPOINT ["/usr/bin/sophrosyne"]
