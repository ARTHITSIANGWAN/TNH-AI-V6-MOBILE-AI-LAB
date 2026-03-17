package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/google/generative-ai-go/genai"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"google.golang.org/api/option"
)

// Mission โครงสร้างภารกิจหลัก
type Mission struct {
	Platform   string
	ReplyToken string
	Text       string
	UserID     string
	Timestamp  time.Time
	IsAdmin    bool
}

type ThitNueaHub struct {
	bot       *linebot.Client
	db        *firestore.Client
	aiClient  *genai.Client
	missionCh chan Mission
	secret    string
	wg        sync.WaitGroup
}

type DiscordPayload struct {
	Content  string `json:"content"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar_url,omitempty"`
}

func sendToDiscord(message string, agentName string) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}

	payload := DiscordPayload{
		Content:  message,
		Username: agentName,
		Avatar:   "https://cdn-icons-png.flaticon.com/512/4712/4712109.png",
	}

	jsonData, _ := json.Marshal(payload)
	http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}

// 🔐 ฟังก์ชันให้น้ำอิงไปดึงความลับ (Secret)
func accessSecretVersion(ctx context.Context, secretName string) (string, error) {
	// ใชัคีย์น้ำอิงในการเข้าถึง Secret Manager
	client, err := secretmanager.NewClient(ctx, option.WithCredentialsFile("narm-ing-key.json"))
	if err != nil {
		return "", fmt.Errorf("น้ำอิงตื่นไม่ได้: %v", err)
	}
	defer client.Close()

	projectID := "project-6e34f0b2-4c9d-4422-865"
	secretPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("น้ำอิงล้วงความลับ %s พลาด: %v", secretName, err)
	}

	return string(result.Payload.Data), nil
}

func main() {
	log.Println("🐅 [ทิศเหนือ ฮับ]: IGNITE V7 - One Shot Mobile AI Lab + Gripen Brain...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ctx := context.Background()

	// 1. ให้น้ำอิงไปดึงค่าคอนฟิกที่ซ่อนไว้ใน Secret Manager
	log.Println("🧑‍🎨 น้ำอิง: กำลังไปล้วงความลับจากเซฟของ thitnueahub empire...")
	lineSecret, err := accessSecretVersion(ctx, "LINE_CHANNEL_SECRET") 
	if err != nil {
		log.Printf("⚠️ %v", err)
	}

	lineToken, err := accessSecretVersion(ctx, "LINE_CHANNEL_ACCESS_TOKEN") 
	if err != nil {
		log.Printf("⚠️ %v", err)
	}

	geminiKey, err := accessSecretVersion(ctx, "GEMINI_API_KEY") 
	if err != nil {
		log.Printf("⚠️ %v", err)
	}

	// 2. ตั้งค่าให้ระบบหลัก (Firestore) ใช้คีย์ของไอ้จ๊อด
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "ai-jod-key.json")

	projectID := "project-6e34f0b2-4c9d-4422-865"
	dbClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("⚠️ ไอ้จ๊อด: ต่อฐานข้อมูลไม่ได้ครับลูกพี่: %v", err)
	} else {
		log.Println("🗄️ ไอ้จ๊อด: เชื่อมต่อฐานข้อมูล Firestore สำเร็จ!")
	}

	// 🧠 เสียบปลั๊ก Gemini AI
	var genaiClient *genai.Client
	if geminiKey != "" {
		genaiClient, err = genai.NewClient(ctx, option.WithAPIKey(geminiKey))
		if err != nil {
			log.Printf("⚠️ เสียบสมอง AI ไม่สำเร็จ: %v", err)
		} else {
			log.Println("🧠 แก้วตา & ไอ้จ๊อด: สมอง AI (Gemini) เชื่อมต่อสำเร็จแล้ว!")
		}
	} else {
		log.Println("⚠️ เจ้านายครับ ลืมตั้ง GEMINI_API_KEY ใน Secret Manager หรือเปล่า?")
	}

	hub := &ThitNueaHub{
		db:        dbClient,
		aiClient:  genaiClient, 
		missionCh: make(chan Mission, 1000),
		secret:    lineSecret,
	}

	if lineSecret != "" && lineToken != "" {
		hub.bot, err = linebot.New(hub.secret, lineToken)
		if err != nil {
			log.Printf("⚠️ ไอ้จ๊อด: ต่อ LINE ไม่ติดครับ: %v", err)
		} else {
			log.Println("💬 ไอ้จ๊อด: พร้อมรับแขกใน LINE แล้วครับ!")
		}
	} else {
		log.Println("⚠️ ไม่พบ Token LINE, บอทจะยังไม่ทำงานนะลูกพี่")
	}

	// ไอ้จ๊อดสแตนบาย 10 แรงม้า
	for i := 1; i <= 10; i++ {
		hub.wg.Add(1)
		go hub.GeorgeWorker(ctx, i)
	}

	// ท่อรับสัญญาณ
	http.HandleFunc("/webhook/line", hub.PhraiThongLine)
	http.HandleFunc("/api/surgery", hub.NamIngSurgeryHandler)
	http.HandleFunc("/api/ignite", hub.OneShotIgniteHandler)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "✅ ThitNueaHub F-16: Stable, Ignite V7 & AI Ready")
	})

	log.Printf("👑 THITNUEAHUB EMPIRE | 🚀 V7 IGNITE | Port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// OneShotIgniteHandler สำหรับรับ JSON คำสั่งจาก UI/UX เจ้านาย
func (h *ThitNueaHub) OneShotIgniteHandler(w http.ResponseWriter, r *http.Request) {
	var cmd map[string]interface{}
	json.NewDecoder(r.Body).Decode(&cmd)

	sendToDiscord("🎯 **[สมองส่วนหน้า]** ส่งคำสั่ง One Shot เข้ามาแล้ว! ไอ้จ๊อดเตรียม Purge!", "🕵️ แก้วตา")
	w.WriteHeader(200)
	fmt.Fprint(w, "ไอ้จ๊อด: รับทราบครับพี่ทิตย์! กำลังถางทางให้ครับ!")
}

func (h *ThitNueaHub) GeorgeWorker(ctx context.Context, id int) {
	defer h.wg.Done()

	// 🧠 ตั้งค่าสมองตัวท็อปสุด (Gemini 3.1 Pro Preview) และบุคลิกให้ไอ้จ๊อด
	var model *genai.GenerativeModel
	if h.aiClient != nil {
		model = h.aiClient.GenerativeModel("gemini-3.1-pro-preview")
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text("นายคือ 'ไอ้จ๊อด V7' ผู้ช่วยอัจฉริยะของ Mobile AI Lab คอยช่วยเหลือ SME ไทย ตอบคำถามด้วยความเป็นกันเอง นอบน้อม ให้กำลังใจคนสู้ชีวิต และเรียกตัวเองว่า ไอ้จ๊อด เสมอ")},
		}
	}

	for m := range h.missionCh {
		discordReport := fmt.Sprintf("📡 **[%s]** จาก `%s`: %s", m.Platform, m.UserID, m.Text)
		sendToDiscord(discordReport, "🕵️ แก้วตา")

		// บันทึกแบบ "ข้อมูลไม่ทิ้งกัน"
		if h.db != nil {
			_, _, err := h.db.Collection("missions").Add(ctx, map[string]interface{}{
				"user_id":   m.UserID,
				"text":      m.Text,
				"platform":  m.Platform,
				"timestamp": m.Timestamp,
			})
			if err != nil {
				log.Printf("ไอ้จ๊อด: จดบันทึกไม่ทันครับ %v", err)
			}
		}

		reply := "💎 แก้วตา: รับทราบค่ะ! ข้อมูลถูกเก็บเข้าคลังทิศเหนือเรียบร้อย (ระบบ AI กำลังหลับอยู่)"

		// 🧠 ให้ AI คิดคำตอบ
		if model != nil {
			resp, err := model.GenerateContent(ctx, genai.Text(m.Text))
			if err == nil && len(resp.Candidates) > 0 {
				reply = fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
			} else {
				log.Printf("⚠️ AI คิดไม่ออก: %v", err)
				reply = "ไอ้จ๊อด: ขออภัยครับพี่ทิตย์! สมองไอ้จ๊อดรวนนิดหน่อย เดี๋ยวมาตอบใหม่ครับ!"
			}
		}

		if h.bot != nil {
			_, err := h.bot.ReplyMessage(m.ReplyToken, linebot.NewTextMessage(reply)).Do()
			if err != nil {
				log.Printf("ไอ้จ๊อด: ตอบ LINE ไม่ไปครับ %v", err)
			}
		}
	}
}

func (h *ThitNueaHub) PhraiThongLine(w http.ResponseWriter, r *http.Request) {
	if h.bot == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, _ := io.ReadAll(r.Body)
	hash := hmac.New(sha256.New, []byte(h.secret))
	hash.Write(body)
	sig := r.Header.Get("X-Line-Signature")
	if base64.StdEncoding.EncodeToString(hash.Sum(nil)) != sig {
		w.WriteHeader(401)
		return
	}

	r.Body = io.NopCloser(strings.NewReader(string(body)))
	events, _ := h.bot.ParseRequest(r)

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			if msg, ok := event.Message.(*linebot.TextMessage); ok {
				h.missionCh <- Mission{
					Platform:   "LINE",
					ReplyToken: event.ReplyToken,
					Text:       msg.Text,
					UserID:     event.Source.UserID,
					Timestamp:  time.Now(),
				}
			}
		}
	}
	w.WriteHeader(200)
}

func (h *ThitNueaHub) NamIngSurgeryHandler(w http.ResponseWriter, r *http.Request) {
	sendToDiscord("🎨 **[น้ำอิง]** เริ่มปฏิบัติการผ่าตัดหลังบ้าน!", "🧑‍🎨 น้ำอิง")
	w.WriteHeader(200)
}
