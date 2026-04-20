# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-w -s" \
      -o /out/app ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /out/app /app

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app"]
