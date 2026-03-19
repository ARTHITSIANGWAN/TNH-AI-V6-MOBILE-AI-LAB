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

# ก๊อปไฟล์โปรแกรมมา
COPY --from=builder /app/thn-core .

# --- สำคัญมาก: ก๊อปโฟลเดอร์หน้าเว็บ (HTML/CSS/Images) มาด้วย ---
# ถ้าพี่ใช้ชื่อโฟลเดอร์อื่น (เช่น public) ให้แก้คำว่า static เป็นชื่อนั้นนะครับ
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./thn-core"]
