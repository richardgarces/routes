# Etapa 1: Construcción de la aplicación
FROM golang:1.20 AS builder

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos del proyecto al contenedor
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Construir el binario de la aplicación
RUN go build -o main .

# Etapa 2: Imagen final para producción
FROM debian:bullseye-slim

# Instalar dependencias necesarias
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Establecer el directorio de trabajo
WORKDIR /root/

# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/main .

# Exponer el puerto de la aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]