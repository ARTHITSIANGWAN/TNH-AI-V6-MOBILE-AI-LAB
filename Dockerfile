# STEP 1: Build the binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o thn-core .

# STEP 2: Final minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
# ตั้งค่า Timezone ให้ตรงกับไทยในระดับ OS
ENV TZ=Asia/Bangkok
COPY --from=builder /app/thn-core .
COPY --from=builder /app/web ./web

ENV PORT=8080
EXPOSE 8080
CMD ["./thn-core"]
