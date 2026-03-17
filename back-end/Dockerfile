# Dockerfile construido com ajuda de IA

# Etapa 1 — Build da aplicação
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copia arquivos de dependência primeiro (cache eficiente)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila a aplicação (binário estático, menor e mais rápido)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Etapa 2 — Imagem final enxuta
FROM alpine:latest

WORKDIR /app

# Copia apenas o binário
COPY --from=builder /app/server .

# Expõe a porta da API
EXPOSE 8080:8080

# Comando para rodar o servidor
CMD ["./server"]