# Stage 2: Final Image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
# เปลี่ยน WORKDIR เป็น /app เพื่อให้สม่ำเสมอ
WORKDIR /app

# 1. ก๊อปตัวโปรแกรมมา
COPY --from=builder /app/thn-core .

# 2. ก๊อปปี้โฟลเดอร์ web มาทั้งยวง (รวมทั้ง index.html และ static)
# วิธีนี้จะทำให้โครงสร้างไฟล์ใน Container เหมือนกับในเครื่องบอสเป๊ะๆ
COPY --from=builder /app/web/ ./web/

EXPOSE 8080
# รันตัวแปรต้นฉบับที่ build มา
CMD ["./thn-core"]
