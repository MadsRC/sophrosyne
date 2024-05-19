FROM cgr.dev/chainguard/go:latest@sha256:de4e3ede01a508b268fa5abd35b0fd43aed8c98af92225304b34680e797d14a1 AS builder
COPY . /app
WORKDIR /app
RUN go build -o dummycheck cmd/dummycheck/main.go

FROM cgr.dev/chainguard/glibc-dynamic:latest@sha256:6dff3d81e2edaa0ef48ea87b808c34c4b24959169d9ad317333bdda4ec3c4002
COPY --from=builder /app/dummycheck /usr/bin/
USER nonroot
CMD ["dummycheck"]
