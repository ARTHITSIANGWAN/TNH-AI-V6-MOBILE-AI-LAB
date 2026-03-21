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

// --- 💎 STRUCTURES ---
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

// --- 🛡️ CORE AI FUNCTIONS (แก้วตา AI) ---
func (eb *EmpireBot) askKaewta(prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + eb.GeminiKey
	fullPrompt := fmt.Sprintf(`คุณคือ 'แก้วตา' เลขา ThitNueaHub ตอบคำถามบอส Art แบบเนี้ยบๆ: %s`, prompt)

	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{{"parts": []map[string]string{{"text": fullPrompt}}}},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil { return "ขออภัยค่ะบอส ระบบ AI ติดขัดนิดหน่อย", err }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// เช็ค Error ป้องกัน Panic
	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 { return "แก้วตาคิดไม่ออกค่ะบอส ลองใหม่อีกทีนะ", nil }
	
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

// --- 🛠️ TELEGRAM & LINE SENDER ---
func (eb *EmpireBot) sendTelegramMessage(text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", eb.TelegramKey)
	payload, _ := json.Marshal(map[string]interface{}{
		"chat_id": eb.TelegramChat,
		"text":    text,
	})
	http.Post(url, "application/json", bytes.NewBuffer(payload))
}

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
		log.Printf("🏍️ [ไอ้จ๊อด-%d]: ลุยงานให้บอส Art อยู่ค่ะ!", id)
		content, _ := eb.askKaewta(mission.Message)
		eb.sendLineReply(mission.ReplyToken, "แก้วตาจัดการ Content ให้บอสเรียบร้อย! ส่งเข้า Telegram แล้วค่ะ ✨")
		eb.sendTelegramMessage(content)
	}
}

// --- 🌐 HANDLERS ---
func (eb *EmpireBot) handleLineCallback(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Events []struct {
			Type    string `json:"type"`
			Message struct { Text string `json:"text" `} `json:"message"`
			ReplyToken string `json:"replyToken"`
		} `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, event := range payload.Events {
		if event.Type == "message" {
			eb.MissionChan <- Mission{
				ReplyToken: event.ReplyToken,
				Message:    event.Message.Text,
			}
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
		TelegramChat: os.Getenv("TELEGRAM_CHAT_ID"),
	}

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go eb.GeorgeWorker(i, &wg)
	}

	// --- [🚀 แก้ปัญหา 404: หน้าแรกแบบไม่ต้องใช้ไฟล์ index.html] ---
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "ไม่พบหน้าเว็บค่ะบอส")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>🐅 ThitNueaHub V7 Ignite (Go) Online!</h1><p>บ้านหลังใหม่รันสำเร็จแล้วค่ะบอส Art!</p>")
	})

	http.HandleFunc("/callback", eb.handleLineCallback)

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
