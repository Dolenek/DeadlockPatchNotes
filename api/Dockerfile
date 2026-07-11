FROM golang:1.25.12-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/sync ./cmd/sync

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app

COPY --from=build /out/server /usr/local/bin/server
COPY --from=build /out/sync /usr/local/bin/sync

EXPOSE 8080
USER app
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O - http://127.0.0.1:8080/api/healthz >/dev/null || exit 1
CMD ["server"]
