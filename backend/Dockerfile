# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
 && addgroup -S app \
 && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/backend /out/migrate ./
COPY --from=builder /src/migrations ./migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/app/backend"]