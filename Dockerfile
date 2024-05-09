FROM cgr.dev/chainguard/glibc-dynamic:latest
COPY --chown=noneroot:noneroot dist/sophrosyne /usr/bin/
ENTRYPOINT ["/usr/bin/sophrosyne"]
