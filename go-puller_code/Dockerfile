# Etapa de compilación
FROM golang:1.26.5 AS builder

WORKDIR /app

# Descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

# Imagen final
FROM alpine:latest

WORKDIR /app

# Certificados para HTTPS
RUN apk --no-cache add ca-certificates

# Copiar el binario
COPY --from=builder /app/app .

# Copiar el archivo de configuración
COPY --from=builder /app/ollama_prompt.txt .

# Exponer el puerto
EXPOSE 8030

# Ejecutar la aplicación
CMD ["./app"]