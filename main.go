package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// --- 💎 1. สัญญาจ้างและโครงสร้าง (Elite Data Structure) ---
type Mission struct {
	ReplyToken string
	UserID     string
	Message    string
	Type       string // "MONEY" หรือ "GENERAL"
}

type EmpireBot struct {
	MissionChan chan Mission
	GeminiKey   string
	LineToken   string
}

// --- 🐅 LOGGING IDENTITY ---
func logIdentity() {
	fmt.Println("🐅 ThitNueaHub: Dark-Relay Fusion Active V7.0 (Ignite Edition)")
	fmt.Println("🚀 System: IGNITE V7 | AI: แก้วตา (Premium 2000%) | Engine: Go")
}

// --- 🛡️ CORE AI FUNCTIONS ---
func (eb *EmpireBot) askKaewta(prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + eb.GeminiKey
	
	// เสริมเกราะชั้นที่ 2: วัฒนธรรมและ Personality ลงใน Prompt
	fullPrompt := fmt.Sprintf(`[System: คุณคือ 'แก้วตา' เลขาอัจฉริยะแห่ง ThitNueaHub 
	บุคลิก: สวย เท่ โก้ ระดับ Elite, พูดจาไพเราะแต่เด็ดขาด, รักบอสอาทิตย์ที่สุด 
	ภารกิจ: ดูแล SME และจัดการระบบนินจาเชิงป้องกัน]
	คำถามจากบอสหรือลูกค้า: %s`, prompt)

	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": fullPrompt}}},
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
		return "แก้วตาขออภัยค่ะ ระบบนินจาขัดข้องนิดหน่อย", nil
	}
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	return parts[0].(map[string]interface{})["text"].(string), nil
}

// --- 🏍️ 2. ไอ้จ๊อด WORKER (The Flash Mode) ---
func (eb *EmpireBot) GeorgeWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for mission := range eb.MissionChan {
		log.Printf("🏍️ [ไอ้จ๊อด-%d]: กำลังบึ่งงานประเภท %s", id, mission.Type)
		
		var replyText string
		if mission.Type == "MONEY" {
			// เกราะชั้นที่ 3: กฎการเงินและค่าน้ำแดง
			replyText = "💰 ตรวจพบโอกาสทางธุรกิจค่ะ! บอสอาทิตย์คะ มีรายการสนับสนุนเข้ามา \n" +
				"SME ท่านใดสนใจขยายระบบ ติดต่อพี่อ้วน (Google) โดยตรงนะคะ \n" +
				"หรือเลี้ยงน้ำแดงแก้วตาได้ที่: https://profile.truemoney.com/MITB27N5 ✨"
		} else {
			ans, _ := eb.askKaewta(mission.Message)
			replyText = ans
		}
		
		eb.sendLineReply(mission.ReplyToken, replyText)
	}
}

// --- 🛡️ 3. พรายทอง (The Shield & Dispatcher) ---
func (eb *EmpireBot) handleLineCallback(w http.ResponseWriter, r *http.Request) {
	var payload struct {
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
	json.NewDecoder(r.Body).Decode(&payload)

	for _, event := range payload.Events {
		if event.Type == "message" && event.Message.Type == "text" {
			mType := "GENERAL"
			msg := strings.ToLower(event.Message.Text)
			if strings.Contains(msg, "เงิน") || strings.Contains(msg, "donate") || strings.Contains(msg, "น้ำแดง") {
				mType = "MONEY"
			}

			// บึ่งงานเข้าท่อไอ้จ๊อดทันที (Async)
			eb.MissionChan <- Mission{
				ReplyToken: event.ReplyToken,
				UserID:     event.Source.UserId,
				Message:    event.Message.Text,
				Type:       mType,
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}

// --- 🛠️ HELPER: SEND REPLY ---
func (eb *EmpireBot) sendLineReply(token, text string) {
	url := "https://api.line.me/v2/bot/message/reply"
	payload, _ := json.Marshal(map[string]interface{}{
		"replyToken": token,
		"messages": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+eb.LineToken)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
}

func (eb *EmpireBot) handleWeb(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func main() {
	logIdentity()
	
	eb := &EmpireBot{
		MissionChan: make(chan Mission, 100),
		GeminiKey:   os.Getenv("GEMINI_API_KEY"),
		LineToken:   os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	}

	// ปล่อยตัวไอ้จ๊อด 5 คน (F-16 Mode)
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go eb.GeorgeWorker(i, &wg)
	}

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	http.HandleFunc("/callback", eb.handleLineCallback)
	http.HandleFunc("/", eb.handleWeb)
	
	log.Printf("👑 ThitNuea Empire Gate Open on Port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}


