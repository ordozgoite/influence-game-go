# Etapa 1 — build
FROM golang:1.24.3-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc g++ make git

COPY . .

RUN go mod download

# builda o binário buffalo (cmd/app/main.go)
RUN go build -o server ./cmd/app

# Etapa 2 — runtime
FROM alpine:3.18
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /app/server

EXPOSE 3000

CMD ["/app/server"]