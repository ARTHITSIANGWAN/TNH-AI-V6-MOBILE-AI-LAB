# BUILD STAGE
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o thn-core .

# FINAL STAGE
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/thn-core .
# ก๊อปทั้งโฟลเดอร์ web เข้าไปเพื่อให้ Path ตรงกับ main.go
COPY --from=builder /app/web ./web
ENV PORT=8080
EXPOSE 8080
CMD ["./thn-core"]
