# 🔧 Стадия сборки
FROM golang:1.24.1 AS builder

WORKDIR /app

# Кэширование зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Собираем бинарник
RUN go build -o release-candle .

# 🧼 Минимальный рантайм
FROM debian:bullseye-slim

# Добавим сертификаты и шрифт, если надо
RUN apt-get update && apt-get install -y \
    ca-certificates \
    fonts-dejavu-core \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем шаблоны, статику и бинарник
COPY --from=builder /app/release-candle .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./release-candle"]
