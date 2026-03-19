# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o thn-core .

# Stage 2: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# 1. ก๊อปตัวโปรแกรมมา
COPY --from=builder /app/thn-core .

# 2. ก๊อปไฟล์หน้าเว็บและรูป (เพราะพี่วางไว้ข้างนอก ไม่ได้ใส่โฟลเดอร์)
COPY --from=builder /app/index.html .
COPY --from=builder /app/image_0.png .

EXPOSE 8080
CMD ["./thn-core"]
