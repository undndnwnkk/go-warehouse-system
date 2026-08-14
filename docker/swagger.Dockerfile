# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/swagger ./api/openapi

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/swagger ./swagger
# main.go serves this file using the relative path "./api/openapi/openapi.yaml"
COPY --from=builder /src/api/openapi/openapi.yaml ./api/openapi/openapi.yaml

USER app
EXPOSE 8085

ENTRYPOINT ["./swagger"]
