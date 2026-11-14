FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/main ./cmd/autoreach

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -g 1001 app \
    && adduser -D -u 1001 -G app app

WORKDIR /app

COPY --from=builder /out/main /app/main
COPY --from=builder /src/migrations /app/migrations

USER app

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT}/health" >/dev/null || exit 1

ENTRYPOINT ["/app/main"]
