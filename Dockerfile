FROM cgr.dev/chainguard/python:latest-dev@sha256:40b3a19b7e2a50824b1ff32d61ae5f59af1b0de67f7bb8e56f5804bace0d94b7 AS builder

ARG dist_file

WORKDIR /app

COPY "dist/${dist_file}" "/home/nonroot/${dist_file}"

RUN --mount=type=secret,id=requirements,dst=/home/nonroot/requirements.txt,uid=65532,gid=65532 \
    pip install --no-cache-dir -r "/home/nonroot/requirements.txt"
RUN pip install --no-cache-dir "/home/nonroot/${dist_file}"

FROM cgr.dev/chainguard/python:latest@sha256:5f16431f56f330925a9c8f5168b31ca65f603de15b127b376f8532bab11583c0

WORKDIR /app

COPY --from=builder /home/nonroot/.local/lib/python3.12/site-packages /home/nonroot/.local/lib/python3.12/site-packages

ENTRYPOINT [ "python", "-m", "sophrosyne.main", "run" ]
