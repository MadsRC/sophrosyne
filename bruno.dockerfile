FROM cgr.dev/chainguard/node:latest-dev as builder

RUN npm install @usebruno/cli

FROM cgr.dev/chainguard/node:latest

WORKDIR /app

COPY --from=builder /app/node_modules /app/node_modules

ENTRYPOINT ["node_modules/.bin/bru"]
