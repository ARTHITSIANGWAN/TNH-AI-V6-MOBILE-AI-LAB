package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// --- 💎 โครงสร้าง ---
type Mission struct {
	ReplyToken string
	Message    string
}

type EmpireBot struct {
	MissionChan  chan Mission
	GeminiKey    string
	LineToken    string
	TelegramKey  string
	TelegramChat string
}

// --- 🛡️ แก้วตา AI (Gemini 1.5 Flash) ---
func (eb *EmpireBot) askKaewta(prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + eb.GeminiKey
	fullPrompt := fmt.Sprintf(`คุณคือ 'แก้วตา' เลขา ThitNueaHub ตอบคำถามบอส Art แบบเนี้ยบๆ ดุดัน: %s`, prompt)

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

// --- 🛠️ ส่งงานเข้า Telegram ---
func (eb *EmpireBot) sendTelegramMessage(text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", eb.TelegramKey)
	payload, _ := json.Marshal(map[string]interface{}{
		"chat_id": eb.TelegramChat,
		"text":    text,
	})
	http.Post(url, "application/json", bytes.NewBuffer(payload))
}

// --- 🛠️ ตอบกลับ LINE ---
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

// --- 🏍️ WORKER ---
func (eb *EmpireBot) GeorgeWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for mission := range eb.MissionChan {
		log.Printf("🏍️ [ไอ้จ๊อด-%d]: กำลังประมวลผล...", id)
		content, err := eb.askKaewta(mission.Message)
		if err == nil {
			eb.sendLineReply(mission.ReplyToken, "แก้วตาทำ Content เสร็จแล้วค่ะบอส! ส่งเข้า Telegram ให้แล้วนะคะ ✨")
			eb.sendTelegramMessage(content)
		}
	}
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
		if event.Type == "message" {
			eb.MissionChan <- Mission{ReplyToken: event.ReplyToken, Message: event.Message.Text}
		}
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	eb := &EmpireBot{
		MissionChan:  make(chan Mission, 100),
		GeminiKey:    os.Getenv("AI_NAM_ING_KEY"), // ใช้กุญแจน้ำอิงคุม
		LineToken:    os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		TelegramKey:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChat: os.Getenv("TELEGRAM_CHAT_ID"),
	}

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go eb.GeorgeWorker(i, &wg)
	}

	// หน้าแรกกัน 404
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "🐅 ThitNueaHub V7 Ignite Online! (Kaewta Mode)")
	})
	http.HandleFunc("/callback", eb.handleLineCallback)

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

