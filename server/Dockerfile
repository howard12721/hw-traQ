# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.4-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/server .

FROM docker.io/library/alpine:3.22

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && \
    adduser -S -G app app

USER app
WORKDIR /app

COPY --from=build /out/server /app/server

EXPOSE 8080
ENV PORT=8080
ENV TRAQ_API_BASE_URL=https://q.trap.jp/api/v3

ENTRYPOINT ["/app/server"]
