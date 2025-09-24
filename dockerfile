# Multi-stage build para otimização
FROM golang:1.24-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum .env ./

# Copiar código fonte
COPY . .

# Download e verificação das dependências
RUN go mod download && \
    go mod verify

# Build da aplicação com otimizações
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o vision-data ./cmd/api/main.go

# Stage final - usar alpine ao invés de scratch para debugging
FROM alpine:3.19

# Labels para metadata
LABEL maintainer="go-gin-api" \
      version="1.0.0" \
      description="Go Gin API com Elasticsearch e Redis"

# Instalar ca-certificates e timezone
RUN apk --no-cache add ca-certificates tzdata

# Criar diretórios necessários (como root primeiro)
RUN mkdir -p /app/logs

WORKDIR /app

# Copiar binário
COPY --from=builder /app/vision-data /app/vision-data

# Copiar .env para a imagem final
COPY --from=builder /app/.env /app/.env

# Ajustar permissões do binário
RUN chmod +x /app/vision-data

# Dar permissões amplas ao diretório de logs (será sobrescrito pelo volume)
RUN chmod 777 /app/logs

# Comando de execução (como root para resolver permissões)
USER root

ENTRYPOINT ["/app/vision-data"]

