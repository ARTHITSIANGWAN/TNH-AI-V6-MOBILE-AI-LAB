# Stage 1: Build โปรแกรม Go
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o thn-core .

# Stage 2: Final Image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# 1. ก๊อปตัวโปรแกรมมา
COPY --from=builder /app/thn-core .

# 2. ก๊อปไฟล์จากโฟลเดอร์ที่ซ้อนอยู่ (web/static) ออกมาวางข้างนอก
# เพื่อให้ตรงกับที่ main.go เรียกหา (index.html และ image_0.png)
COPY --from=builder /app/web/static/index.html .
COPY --from=builder /app/web/static/image_0.png .

EXPOSE 8080
CMD ["./thn-core"]
