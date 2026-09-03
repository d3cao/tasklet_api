# --- Etapa 1: Construção e Testes ---
FROM golang:1.27-alpine AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências primeiro (otimização de cache)
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte
COPY . .

# Executa os testes automatizados
RUN CGO_ENABLED=0 go test -v ./...

# Compila o binário otimizado
RUN CGO_ENABLED=0 GOOS=linux go build -o meu-app .

# --- Etapa 2: Imagem de Produção Final ---
FROM alpine:3.19 AS runner

WORKDIR /app

# Copia apenas o binário compilado da etapa 'builder'
COPY --from=builder /app/meu-app .

EXPOSE 8080

CMD ["./meu-app"]
