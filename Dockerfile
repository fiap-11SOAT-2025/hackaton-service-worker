# ==========================================
# STAGE 1: Build
# ==========================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instala git para baixar dependências do Go
RUN apk add --no-cache git

# Baixa as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Compila o binário do worker (sem CGO para rodar lisinho no Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o hackaton-worker cmd/main.go

# ==========================================
# STAGE 2: Produção (A Imagem Final)
# ==========================================
FROM alpine:latest

WORKDIR /app

# ATENÇÃO AQUI: Instala os certificados de rede (para acessar a AWS), tzdata e o FFMPEG (para processar os vídeos)
RUN apk --no-cache add ca-certificates tzdata ffmpeg

# Traz o binário compilado do Stage 1
COPY --from=builder /app/hackaton-worker .

# Inicia o Worker
CMD ["./hackaton-worker"]