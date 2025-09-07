# Multi-stage build para otimização
FROM golang:1.24-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

RUN ls -la

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
    -o visiona-data ./cmd/api/main.go

# Stage final - imagem mínima
FROM scratch

# Labels para metadata
LABEL maintainer="go-gin-api" \
      version="1.0.0" \
      description="Go Gin API com Elasticsearch e Redis"

# Copiar certificados SSL, timezone e passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copiar binário
COPY --from=builder /app/visiona-data /visiona-data

# Copiar .env para a imagem final
COPY --from=builder /app/.env /app/.env

# Expor porta
EXPOSE 8080

# Comando de execução
ENTRYPOINT ["/visiona-data"]