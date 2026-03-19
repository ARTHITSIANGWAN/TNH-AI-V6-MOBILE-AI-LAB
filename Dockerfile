FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o thn-core .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
# ก๊อปตัวโปรแกรมมา
COPY --from=builder /app/thn-core .
# ต้องก๊อปโฟลเดอร์หน้าเว็บ (Static files) มาด้วยนะบอส! 
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./thn-core"]
