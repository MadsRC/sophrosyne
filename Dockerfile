FROM cgr.dev/chainguard/python:latest-dev AS builder

ARG dist_file

WORKDIR /app

COPY "dist/${dist_file}" "/home/nonroot/${dist_file}"

RUN --mount=type=secret,id=requirements,dst=/home/nonroot/requirements.txt,uid=65532,gid=65532 \
    pip install --no-cache-dir -r "/home/nonroot/requirements.txt"
RUN pip install --no-cache-dir "/home/nonroot/${dist_file}"

FROM cgr.dev/chainguard/python:latest

WORKDIR /app

COPY --from=builder /home/nonroot/.local/lib/python3.12/site-packages /home/nonroot/.local/lib/python3.12/site-packages

ENTRYPOINT [ "python", "-m", "sophrosyne.main", "run" ]
