# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .
RUN go build -o backend-server

# Финальный образ
FROM alpine:latest

# Добавляем необходимые сертификаты для SMTP соединений через TLS
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/backend-server .

# Установка переменной окружения для порта
ENV PORT=8080

# Открываем порт
EXPOSE 8080

# Запуск приложения
CMD ["./backend-server"]