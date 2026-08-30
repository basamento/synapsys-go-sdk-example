# syntax=docker/dockerfile:1

# Builds the example exactly as a user would: the SDK is resolved from the
# public Go module proxy, not from a local checkout.

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/example .

FROM alpine:3.20

RUN addgroup -S worker && adduser -S -G worker worker

COPY --from=build /out/example /app/example

USER worker

# Local status page. Not published by the sandbox; Core is the console.
EXPOSE 3005

ENTRYPOINT ["/app/example"]
