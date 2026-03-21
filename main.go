package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// --- 💎 STRUCTURES ---
type Mission struct {
	ReplyToken string // สำหรับ LINE
	TelegramID string // สำหรับส่งงานเข้า Telegram
	Message    string
	Type       string 
}

type EmpireBot struct {
	MissionChan  chan Mission
	GeminiKey    string
	LineToken    string
	TelegramKey  string
	TelegramChat string // ID ของบอสใน Telegram
}

// --- 🛡️ CORE AI FUNCTIONS ---
func (eb *EmpireBot) askKaewta(prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + eb.GeminiKey
	fullPrompt := fmt.Sprintf(`คุณคือ 'แก้วตา' เลขา ThitNueaHub ตอบคำถามบอส Art แบบเนี้ยบๆ: %s`, prompt)

	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{{"parts": []map[string]string{{"text": fullPrompt}}}},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil { return "", err }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	candidates := result["candidates"].([]interface{})
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

// --- 🏍️ WORKER: บึ่งงานส่งไป Telegram ---
func (eb *EmpireBot) GeorgeWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for mission := range eb.MissionChan {
		log.Printf("🏍️ [ไอ้จ๊อด-%d]: กำลังประมวลผล Content...", id)
		
		content, _ := eb.askKaewta(mission.Message)

		// 1. ตอบกลับทาง LINE สั้นๆ ว่าส่งงานไปที่ Telegram แล้ว
		eb.sendLineReply(mission.ReplyToken, "แก้วตาทำ Content เสร็จแล้วค่ะบอส! ส่งเข้า Telegram ให้แล้วนะคะ ✨")

		// 2. ส่งเนื้อหาเต็มเข้า Telegram เพื่อให้บอสก๊อปวาง
		eb.sendTelegramMessage(content)
	}
}

// --- 🛠️ TELEGRAM SENDER ---
func (eb *EmpireBot) sendTelegramMessage(text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", eb.TelegramKey)
	payload, _ := json.Marshal(map[string]interface{}{
		"chat_id": eb.TelegramChat, // ID ของบอส
		"text":    text,
	})
	http.Post(url, "application/json", bytes.NewBuffer(payload))
}

// --- 🛠️ LINE SENDER ---
func (eb *EmpireBot) sendLineReply(token, text string) {
	url := "https://api.line.me/v2/bot/message/reply"
	payload, _ := json.Marshal(map[string]interface{}{
		"replyToken": token,
		"messages": []map[string]interface{}{{"type": "text", "text": text}},
	})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+eb.LineToken)
	client := &http.Client{}
	client.Do(req)
}

func (eb *EmpireBot) handleLineCallback(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Events []struct {
			Type    string `json:"type"`
			Message struct { Text string `json:"text" `} `json:"message"`
			ReplyToken string `json:"replyToken"`
		} `json:"events"`
	}
	json.NewDecoder(r.Body).Decode(&payload)

	for _, event := range payload.Events {
		eb.MissionChan <- Mission{
			ReplyToken: event.ReplyToken,
			Message:    event.Message.Text,
		}
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	eb := &EmpireBot{
		MissionChan:  make(chan Mission, 100),
		GeminiKey:    os.Getenv("GEMINI_API_KEY"),
		LineToken:    os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		TelegramKey:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChat: os.Getenv("TELEGRAM_CHAT_ID"), // ใส่ ID บอสที่นี่
	}

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go eb.GeorgeWorker(i, &wg)
	}

	http.HandleFunc("/callback", eb.handleLineCallback)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/index.html") })

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
