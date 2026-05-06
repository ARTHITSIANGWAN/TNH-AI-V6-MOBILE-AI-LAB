# --- [Stage 1: Build the Go Engine] ---
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# คอมไพล์ Go Engine ให้เป็นไบนารีที่เบาหวิว
RUN go build -o tnh-engine ./cmd/main.go

# --- [Stage 2: Final Sovereign Image] ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /root/

# ดึงไบนารีจาก Stage 1
COPY --from=builder /app/tnh-engine .
# ดึงไฟล์ Web UI (V6 ที่เราแก้ใหม่เป็น V8.3)
COPY ./web ./web

# ตั้งค่า Environment สำหรับ Cloud Run
ENV PORT=2026
EXPOSE 2026

# จุดไฟจักรวรรดิ!
CMD ["./tnh-engine"]
