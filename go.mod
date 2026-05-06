module github.com/thitnueahub/thitnueahub-mobile-ai-lab

go 1.22 // อัปเพื่อใช้ฟีเจอร์ For-Loop semantics ใหม่ และรองรับ Library ปี 2026

require (
	cloud.google.com/go/firestore v1.15.0 // เสถียรขึ้นเพื่อเก็บ State ของขุนพล
	github.com/line/line-bot-sdk-go/v8 v8.0.1 // อัปเป็น V8 เพื่อรองรับ Rich Menu แบบ Flex ล่าสุด
	github.com/google/generative-ai-go v0.12.0 // สำคัญ! ต้องเวอร์ชันนี้ถึงจะใช้ Multimodal File Search & Webhooks ได้
	google.golang.org/api v0.170.0 // เพื่อคุยกับ Google Cloud แบบ Sub-50ms
)

