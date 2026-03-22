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

// --- 💎 โครงสร้างภารกิจ ---
type Mission struct {
	ReplyToken string
	Message    string
}

type EmpireBot struct {
	MissionChan  chan Mission
	GeminiKey    string // นี่คือพลังของ "น้ำอิง"
	LineToken    string
	TelegramKey  string
	TelegramChat string
}

// --- 🛡️ หัวใจของระบบ: น้ำอิง x แก้วตา ---
func (eb *EmpireBot) askAI(prompt string) (string, error) {
	// ใช้ Key จาก Secret AI_NAM_ING_KEY ที่บอสตั้งไว้
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + eb.GeminiKey
	
	// คอนเซปต์: แก้วตาหน้าบ้านรับคำสั่งบอส แล้วให้น้ำอิงหลังบ้านประมวลผล
	fullPrompt := fmt.Sprintf(`คุณคือ 'แก้วตา' เลขาหน้าบ้านของ ThitNueaHub โดยมี 'น้ำอิง' คุมระบบหลังบ้าน ตอบคำถามบอส Art แบบเนี้ยบๆ ดุดันแต่จริงใจ: %s`, prompt)

	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{{"parts": []map[string]string{{"text": fullPrompt}}}},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil { return "ขออภัยค่ะบอส น้ำอิงหลังบ้านแจ้งว่าระบบติดขัด", err }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 { return "แก้วตาติดต่อคุณน้ำอิงไม่ได้ค่ะบอส ลองใหม่อีกทีนะ", nil }
	
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

// --- 🛠️ ระบบส่งงาน (LINE & Telegram) ---
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

// --- 🏍️ WORKER: ไอ้จ๊อด (คนส่งของ) ---
func (eb *EmpireBot) GeorgeWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for mission := range eb.MissionChan {
		log.Printf("🏍️ [ไอ้จ๊อด-%d]: รับงานจากแก้วตา ส่งให้น้ำอิงปั่น...", id)
		content, _ := eb.askAI(mission.Message)
		eb.sendLineReply(mission.ReplyToken, "แก้วตาประสานงานคุณน้ำอิงให้เรียบร้อย! ข้อมูลพร้อมแล้วค่ะบอส ✨")
		// บอสสามารถเพิ่มฟังก์ชันส่งเข้า Telegram ตรงนี้ได้เลยถ้าตั้งค่า Key ไว้
		log.Println("✅ งานเสร็จสิ้น:", content[:20], "...")
	}
}

// --- 🌐 ระบบรับแขก (Handlers) ---
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
		// ดึงค่าจาก Secret ที่บอสตั้งไว้ในรูป (AI_NAM_ING_KEY)
		GeminiKey:    os.Getenv("AI_NAM_ING_KEY"), 
		LineToken:    os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	}

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go eb.GeorgeWorker(i, &wg)
	}

	// หน้าแรก (ป้องกัน 404)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>🐅 ThitNueaHub V7: แก้วตา x น้ำอิง</h1><p>หลังบ้านน้ำอิงคุม หน้าบ้านแก้วตาดูแล... ออนไลน์แล้วค่ะบอส!</p>")
	})

	http.HandleFunc("/callback", eb.handleLineCallback)

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
