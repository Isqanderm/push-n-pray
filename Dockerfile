FROM golang:1.24.1 AS builder
WORKDIR /app
COPY . .
RUN go build -o release-candle .


FROM golang:1.24.1
WORKDIR /app
COPY --from=builder /app/release-candle .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/assets ./assets
EXPOSE 8080
CMD ["./release-candle"]
