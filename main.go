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
	GeminiAPIUrl = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash:generateContent?key="
)

// --- LOGGING ---
func logIdentity() {
	fmt.Println("🐅 ThitNueaHub: Dark-Relay Fusion Active V4.2 (Stable Edition)")
	fmt.Println("🚀 System: IGNITE V7 | AI: Gemini 3 Flash | Engine: Go")
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

type LineReplyMessage struct {
	ReplyToken string                   `json:"replyToken"`
	Messages   []map[string]interface{} `json:"messages"`
}

type LinePushMessage struct {
	To       string                   `json:"to"`
	Messages []map[string]interface{} `json:"messages"`
}

// --- CORE FUNCTIONS ---

func generateGeminiContent(apiKey, prompt string) (string, error) {
	url := GeminiAPIUrl + apiKey
	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.8,
			"topP":            0.95,
			"maxOutputTokens": 2048,
		},
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Gemini API Error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("no candidates from Gemini")
	}
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

func darkRelayExecution() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	tgToken := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("CHAT_ID")
	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	targetLineID := os.Getenv("TARGET_LINE_ID")

	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is missing")
	}

	report, err := generateGeminiContent(apiKey, "Task: Generate an elite tech insight for SME. Style: Aggressive, Zero-Garbage, Highly Professional. Language: Thai.")
	if err != nil {
		return err
	}

	fullMsg := "🛡️ **DARK-RELAY FUSION REPORT**\n\n" + report

	// Send Telegram
	if tgToken != "" && chatID != "" {
		tgUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tgToken)
		tgPayload, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": fullMsg, "parse_mode": "Markdown"})
		http.Post(tgUrl, "application/json", bytes.NewBuffer(tgPayload))
	}

	// Send LINE Push
	if lineToken != "" && targetLineID != "" {
		sendLinePush(lineToken, targetLineID, fullMsg)
	}

	return nil
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

// --- HTTP HANDLERS ---

func handleLineCallback(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	lineToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")

	var payload LineWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, event := range payload.Events {
		if event.Type == "message" && event.Message.Type == "text" {
			// บอสทักมา ไอ้จ๊อดตอบกลับผ่าน Gemini
			prompt := fmt.Sprintf("คุณคือ 'น้องน้ำอิง' AI ผู้ช่วยของบอส Art (อาทิตย์). ตอบกลับข้อความนี้ด้วยสไตล์เท่ๆ ตรงไปตรงมา: %s", event.Message.Text)
			reply, err := generateGeminiContent(apiKey, prompt)
			if err != nil {
				log.Printf("Gemini Error: %v", err)
				reply = "โอยบอส! สมองน้ำอิงช็อตนิดหน่อย ลองใหม่นะงับ! ⚡"
			}
			sendLineReply(lineToken, event.ReplyToken, reply)
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

// --- RUNTIME ---

func runScheduler() {
	targetHours := []int{8, 12, 20}
	loc, _ := time.LoadLocation("Asia/Bangkok")
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
		fmt.Printf("😴 [Scheduler]: Next Relay at %s\n", nextRun.Format("15:04:05"))
		time.Sleep(time.Until(nextRun))
		darkRelayExecution()
	}
}

func main() {
	logIdentity()
	go runScheduler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/callback", handleLineCallback)
	http.HandleFunc("/", handleWeb)

	fmt.Printf("🚪 ThitNueaHub Gate Open on Port: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
