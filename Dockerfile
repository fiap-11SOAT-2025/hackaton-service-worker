# Etapa 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o hackaton-service-worker cmd/main.go

# Etapa 2: Runner
FROM alpine:latest

WORKDIR /app

# Instala certificados E o FFMPEG (Essencial para o Worker)
RUN apk --no-cache add ca-certificates ffmpeg

# Copia o binário
COPY --from=builder /app/hackaton-service-worker .

# Cria pasta temporária para processamento (opcional, mas boa prática)
RUN mkdir -p temp

CMD ["./hackaton-service-worker"]