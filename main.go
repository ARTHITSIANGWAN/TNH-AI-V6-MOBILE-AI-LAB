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
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// Mission โครงสร้างภารกิจหลัก
type Mission struct {
	Platform   string
	ReplyToken string
	Text       string
	UserID     string
	Timestamp  time.Time
	IsAdmin    bool // เพิ่มเพื่อเช็คสิทธิ์สมองส่วนหน้า
}

type ThitNueaHub struct {
	bot       *linebot.Client
	db        *firestore.Client
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
		return // เงียบไว้ถ้ายังไม่เสียบท่อ Discord
	}

	payload := DiscordPayload{
		Content:  message,
		Username: agentName,
		Avatar:   "https://cdn-icons-png.flaticon.com/512/4712/4712109.png", 
	}

	jsonData, _ := json.Marshal(payload)
	http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}

func main() {
	log.Println("🐅 [ทิศเหนือ ฮับ]: IGNITE V7 - One Shot Mobile AI Lab...")

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	ctx := context.Background()

	// ใช้ Project ID จากถังเหลือง
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	dbClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("⚠️ Firestore Ready Check: %v", err)
	}

	hub := &ThitNueaHub{
		db:        dbClient,
		missionCh: make(chan Mission, 1000),
		secret:    os.Getenv("LINE_CHANNEL_SECRET"),
	}

	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	hub.bot, _ = linebot.New(hub.secret, lineToken)

	// ไอ้จ๊อดสแตนบาย 10 แรงม้า
	for i := 1; i <= 10; i++ {
		hub.wg.Add(1)
		go hub.GeorgeWorker(ctx, i)
	}

	// ท่อรับสัญญาณ
	http.HandleFunc("/webhook/line", hub.PhraiThongLine)
	http.HandleFunc("/api/surgery", hub.NamIngSurgeryHandler)
	http.HandleFunc("/api/ignite", hub.OneShotIgniteHandler) // ท่อใหม่สำหรับ JSON One Shot

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "✅ ThitNueaHub F-16: Stable & Ignite V7")
	})

	log.Printf("👑 THITNUEA HUB | 🚀 V7 IGNITE | Port: %s\n", port)
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
	for m := range h.missionCh {
		// รายงานเข้า Matrix
		discordReport := fmt.Sprintf("📡 **[%s]** จาก `%s`: %s", m.Platform, m.UserID, m.Text)
		sendToDiscord(discordReport, "🕵️ แก้วตา")

		// บันทึกแบบ "ข้อมูลไม่ทิ้งกัน"
		if h.db != nil {
			h.db.Collection("missions").Add(ctx, map[string]interface{}{
				"user_id":   m.UserID,
				"text":      m.Text,
				"platform":  m.Platform,
				"timestamp": m.Timestamp,
			})
		}

		// ตอบกลับนิ่มๆ
		reply := "💎 แก้วตา: รับทราบค่ะ! ข้อมูลถูกเก็บเข้าคลังทิศเหนือเรียบร้อย"
		h.bot.ReplyMessage(m.ReplyToken, linebot.NewTextMessage(reply)).Do()
	}
}

func (h *ThitNueaHub) PhraiThongLine(w http.ResponseWriter, r *http.Request) {
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
					Platform: "LINE",
					ReplyToken: event.ReplyToken,
					Text: msg.Text,
					UserID: event.Source.UserID,
					Timestamp: time.Now(),
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
