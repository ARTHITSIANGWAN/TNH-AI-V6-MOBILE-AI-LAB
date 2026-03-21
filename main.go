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
	"time"
)

// --- CONFIGURATION ---
const (
	LineAPIReply = "https://api.line.me/v2/bot/message/reply"
	LineAPIPush  = "https://api.line.me/v2/bot/message/push"
	GeminiAPIUrl = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash:generateContent?key="
)

// --- LOGGING ---
func logIdentity() {
	fmt.Println("🐅 ThitNueaHub: Dark-Relay Fusion Active V4.0 (LINE & Telegram)")
	fmt.Println("🚀 System: IGNITE V7 | AI: Gemini 3 Flash | Engine: Go")
}

// --- STRUCTURES ---

// LINE Webhook Payload
type LineWebhookPayload struct {
	Destination string `json:"destination"`
	Events      []struct {
		Type    string `json:"type"`
		Message struct {
			Type string `json:"type"`
			Id   string `json:"id"`
			Text string `json:"text"`
		} `json:"message"`
		Timestamp  int64 `json:"timestamp"`
		Source     struct {
			Type   string `json:"type"`
			UserId string `json:"userId"`
		} `json:"source"`
		ReplyToken string `json:"replyToken"`
		Mode       string `json:"mode"`
	} `json:"events"`
}

// LINE Reply Message
type LineReplyMessage struct {
	ReplyToken string `json:"replyToken"`
	Messages   []map[string]interface{} `json:"messages"`
}

// LINE Push Message (สำหรับส่งรายงาน)
type LinePushMessage struct {
	To       string `json:"to"`
	Messages []map[string]interface{} `json:"messages"`
}

// --- EXECUTION (LINE & TELEGRAM FUSION) ---

func darkRelayExecution() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	tgToken := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("CHAT_ID")
	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	targetLineID := os.Getenv("TARGET_LINE_ID") // ไอดี LINE ของบอสที่จะรับรายงาน

	if apiKey == "" { return fmt.Errorf("secure credentials missing") }

	// 1. เจนข้อความรายงานด้วย Gemini
	finalText, err := generateGeminiContent(apiKey, "Task: Generate an elite tech insight for SME. Style: Aggressive, Zero-Garbage, Highly Professional. Language: Thai.")
	if err != nil { return err }

	// 2. ส่ง Telegram (ถ้ามีกุญแจ)
	if tgToken != "" && chatID != "" {
		sendTelegram(tgToken, chatID, "🛡️ **DARK-RELAY FUSION REPORT (Gemini 3)**\n\n"+finalText)
	}

	// 3. ส่ง LINE (ถ้ามีกุญแจและไอดีเป้าหมาย)
	if lineToken != "" && targetLineID != "" {
		sendLinePush(lineToken, targetLineID, "🛡️ **DARK-RELAY FUSION REPORT (Gemini 3)**\n\n"+finalText)
	}

	return nil
}

// --- HELPER FUNCTIONS ---

func generateGeminiContent(apiKey, prompt string) (string, error) {
	url := GeminiAPIUrl + apiKey
	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.7,
			"topP": 0.95,
			"maxOutputTokens": 2048, // ปรับ Max Tokens ให้ยาวขึ้นหน่อย
		},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil { return "", err }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// สกัดเนื้อหา (Gemini 3 Structure)
	candidates := result["candidates"].([]interface{})
	if len(candidates) == 0 { return "", fmt.Errorf("no candidates found") }
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

func sendTelegram(token, chatID, text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text, "parse_mode": "Markdown"})
	http.Post(url, "application/json", bytes.NewBuffer(payload))
}

func sendLineReply(token, replyToken, text string) {
	payload, _ := json.Marshal(LineReplyMessage{
		ReplyToken: replyToken,
		Messages: []map[string]interface{}{
			{"type": "text", "text": text},
		},
	})
	req, _ := http.NewRequest("POST", LineAPIReply, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	client.Do(req)
}

func sendLinePush(token, toID, text string) {
	payload, _ := json.Marshal(LinePushMessage{
		To:       toID,
		Messages: []map[string]interface{}{
			{"type": "text", "text": text},
		},
	})
	req, _ := http.NewRequest("POST", LineAPIPush, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	client.Do(req)
}

// --- MAIN HANDLERS ---

func handleLineCallback(w http.ResponseWriter, r *http.Request) {
	// 1. อ่านค่าจาก Webhook
	apiKey := os.Getenv("GEMINI_API_KEY")
	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	if apiKey == "" || lineToken == "" {
		log.Println("❌ Secure credentials missing for LINE")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var payload LineWebhookPayload
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("❌ Line Webhook Decode Error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. ประมวลผลแต่ละ Event
	for _, event := range payload.Events {
		if event.Type == "message" && event.Message.Type == "text" {
			userText := event.Message.Text
			replyToken := event.ReplyToken
			log.Printf("📩 Received LINE Message: %s", userText)

			// 3. เอาข้อความส่งให้ Gemini (เพิ่ม Prompt ให้ฉลาดขึ้น)
			geminiResponse, err := generateGeminiContent(apiKey, "บอสชื่อ อาทิตย์ พฤกษาอุทัยโฆษิต (Art). เจ้าของ ThitNueaHub (ตึก18ชั้น). Style: Aggressive, Zero-Garbage, Highly Professional. Language: Thai. ตอบข้อความบอส Art ดังนี้: "+userText)
			if err != nil {
				log.Printf("❌ Gemini Error during reply: %v", err)
				sendLineReply(lineToken, replyToken, "โอยบอส! เครื่อง Gemini ร้อนนิดหน่อยงับ!! ลองใหม่นะบอส!! 🤣🚀")
				continue
			}

			// 4. ตอบกลับ LINE
			sendLineReply(lineToken, replyToken, geminiResponse)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func handleWeb(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.FileServer(http.Dir("web")).ServeHTTP(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "web/index.html")
}

// --- SCHEDULER (คงเดิม) ---

func runScheduler() {
	targetHours := []int{8, 12, 20}
	loc := time.FixedZone("Asia/Bangkok", 7*60*60)
	for {
		now := time.Now().In(loc)
		nextRun := time.Time{}
		found := false
		for _, hour := range targetHours {
			targetTime := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, loc)
			if targetTime.After(now) {
				nextRun = targetTime
				found = true
				break
			}
		}
		if !found {
			nextRun = time.Date(now.Year(), now.Month(), now.Day()+1, targetHours[0], 0, 0, 0, loc)
		}
		fmt.Printf("😴 [Gas Station]: Next Relay at %s\n", nextRun.Format("15:04:05"))
		time.Sleep(time.Until(nextRun))
		
		if err := darkRelayExecution(); err != nil {
			log.Printf("❌ Snake Nudge Error: %v", err)
			time.Sleep(2 * time.Minute) 
			darkRelayExecution()
		}
	}
}

// --- MAIN (เพิ่ม Handle /callback) ---

func main() {
	logIdentity()
	go runScheduler()
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// ทางเข้าลับสำหรับ LINE (Webhook URL)
	http.HandleFunc("/callback", handleLineCallback)

	// ทางเข้าหน้าเว็บปกติ
	http.HandleFunc("/", handleWeb)

	fmt.Printf("🚪 ThitNueaHub Gate Open on Port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
