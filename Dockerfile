# Build SPA
FROM node:23-alpine AS SPA

COPY ./ui /ui
COPY Makefile /
WORKDIR /

RUN apk add --no-cache make && make ui

# Build binary
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache make git

COPY . /app
WORKDIR /app

## Embed SPA
COPY --from=SPA /ui/dist /app/ui/dist

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN make creamy-prediction-market && mv creamy-prediction-market /creamy-prediction-market

# Lightweight Runtime Env
FROM alpine:3.20
RUN apk add --no-cache tini
COPY --from=builder /creamy-prediction-market /creamy-prediction-market
ENTRYPOINT ["tini", "--"]
CMD ["/creamy-prediction-market"]
