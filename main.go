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

type Mission struct {
	Platform   string
	ReplyToken string
	Text       string
	UserID     string
	Timestamp  time.Time
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
		log.Println("⚠️ สถาปนิกเตือน: ท่อ DISCORD_WEBHOOK_URL ยังไม่ได้เชื่อมต่อ!")
		return
	}

	payload := DiscordPayload{
		Content:  message,
		Username: agentName,
		Avatar:   "https://cdn-icons-png.flaticon.com/512/4712/4712109.png", 
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ ยิง Discord พลาด: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

func main() {
	log.Println("🐅 [ทิศเหนือ ฮับ]: IGNITE - Mobile AI Lab Online...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	dbClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("⚠️ Firestore Warning: %v", err)
	}

	hub := &ThitNueaHub{
		db:        dbClient,
		missionCh: make(chan Mission, 1000),
		secret:    os.Getenv("LINE_CHANNEL_SECRET"),
	}

	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	hub.bot, _ = linebot.New(hub.secret, lineToken)

	sendToDiscord("🚀 **[SYSTEM IGNITE]** ThitNueaHub Mobile AI Lab พร้อมทะยานเข้าสู่ Matrix แล้วเจ้านาย!", "🐅 ทิศเหนือ ฮับ (System)")

	for i := 1; i <= 10; i++ {
		hub.wg.Add(1)
		go hub.GeorgeWorker(ctx, i)
	}

	http.HandleFunc("/webhook/line", hub.PhraiThongLine)
	http.HandleFunc("/webhook/facebook", hub.PhraiThongMeta)
	http.HandleFunc("/webhook/telegram", hub.PhraiThongTelegram)
	http.HandleFunc("/api/surgery", hub.NamIngSurgeryHandler)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "✅ ThitNueaHub F-16: Stable & Ignite")
	})

	log.Printf("👑 THITNUEA HUB | 🚀 GLOBAL IGNITE | Port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (h *ThitNueaHub) PhraiThongLine(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	hash := hmac.New(sha256.New, []byte(h.secret))
	hash.Write(body)
	sig := r.Header.Get("X-Line-Signature")
	if base64.StdEncoding.EncodeToString(hash.Sum(nil)) != sig {
		http.Error(w, "Unauthorized", 401)
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
	sendToDiscord("🎨 **[น้ำอิง]** ได้รับคำสั่งผ่าตัด (Surgery) จากหน้าแอปแล้วค่ะเจ้านาย!", "🧑‍🎨 น้ำอิง")
	w.WriteHeader(200)
	fmt.Fprint(w, "น้ำอิง: ทิศเหนือ ฮับ จัดการให้เรียบร้อยแล้วค่ะ!")
}

func (h *ThitNueaHub) PhraiThongMeta(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (h *ThitNueaHub) PhraiThongTelegram(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (h *ThitNueaHub) GeorgeWorker(ctx context.Context, id int) {
	defer h.wg.Done()
	for m := range h.missionCh {
		discordReport := fmt.Sprintf("📡 **[สัญญาณจาก %s]**\n👤 ผู้ใช้: `%s`\n💬 ข้อความ: *%s*", m.Platform, m.UserID, m.Text)
		sendToDiscord(discordReport, "🕵️ แก้วตา (ศูนย์บัญชาการ)")

		if h.db != nil {
			_, _, _ = h.db.Collection("missions").Add(ctx, map[string]interface{}{
				"user_id":   m.UserID,
				"text":      m.Text,
				"platform":  m.Platform,
				"timestamp": m.Timestamp,
			})
		}

		reply := "💎 แก้วตา: ทิศเหนือ ฮับ รับทราบค่ะ! กำลังส่งให้น้ำอิงจัดการนะคะ"
		h.bot.ReplyMessage(m.ReplyToken, linebot.NewTextMessage(reply)).Do()
	}
}
