package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// --- CONFIGURATION ---
const (
	LineAPIReply = "https://api.line.me/v2/bot/message/reply"
	LineAPIPush  = "https://api.line.me/v2/bot/message/push"
	// ใช้รุ่นล่าสุดตามที่บอสเลือก
	GeminiAPIUrl = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key="
)

// --- LOGGING ---
func logIdentity() {
	fmt.Println("🐅 ThitNueaHub: Dark-Relay Fusion Active V4.2 (Stable Edition)")
	fmt.Println("🚀 System: IGNITE V7 | AI: แก้วตา (Gemini 1.5 Flash) | Engine: Go")
}

// --- STRUCTURES ---
type LineWebhookPayload struct {
	Events []struct {
		Type    string `json:"type"`
		Message struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"message"`
		ReplyToken string `json:"replyToken"`
		Source     struct {
			UserId string `json:"userId"`
		} `json:"source"`
	} `json:"events"`
}

// --- CORE FUNCTIONS ---
func generateGeminiContent(apiKey, prompt string) (string, error) {
	url := GeminiAPIUrl + apiKey
	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"topP":            0.9,
			"maxOutputTokens": 2048,
		},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "แก้วตาขออภัยค่ะ ระบบประมวลผลขัดข้อง", nil
	}
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

// --- HTTP HANDLERS ---

// สำหรับหน้าเว็บ (UI แก้วตา)
func handleAskKaewta(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	query := r.URL.Query().Get("q")
	if query == "" {
		fmt.Fprint(w, "บอสคะ พิมพ์คำสั่งมาได้เลยค่ะ แก้วตารออยู่!")
		return
	}

	prompt := fmt.Sprintf(`คุณคือ "แก้วตา" (Kaewta) AI Agent ของ ThitNueaHub 
	มีหน้าที่จัดการโปรเจกต์ project-6e34f0b2 และชุดข้อมูล THN_VISION_CORE_V1 
	ภายใต้สิทธิ์ narm-ing-agent ตอบคำถามบอส Art อย่างเท่ๆ และทำได้จริง: %s`, query)

	reply, err := generateGeminiContent(apiKey, prompt)
	if err != nil {
		fmt.Fprint(w, "แก้วตาเชื่อมต่อฐานข้อมูลไม่ได้ค่ะบอส!")
		return
	}
	fmt.Fprint(w, reply)
}

func handleLineCallback(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")

	var payload LineWebhookPayload
	json.NewDecoder(r.Body).Decode(&payload)

	for _, event := range payload.Events {
		if event.Type == "message" && event.Message.Type == "text" {
			prompt := fmt.Sprintf("คุณคือ 'แก้วตา' ผู้ช่วยของบอส Art. ตอบข้อความนี้แบบตรงไปตรงมา: %s", event.Message.Text)
			reply, _ := generateGeminiContent(apiKey, prompt)
			// (ใช้ฟังก์ชันส่ง Reply เดิมของบอส)
			sendLineReply(lineToken, event.ReplyToken, reply)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func handleWeb(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func main() {
	logIdentity()
	
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	http.HandleFunc("/callback", handleLineCallback)
	http.HandleFunc("/ask", handleAskKaewta)
	http.HandleFunc("/", handleWeb)
    
    // โหลด Static Files (CSS/JS)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Printf("🚪 ThitNueaHub Gate Open on Port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// (เพิ่มฟังก์ชัน sendLineReply และอื่นๆ ของบอสไว้ด้านล่างตามปกติ)
