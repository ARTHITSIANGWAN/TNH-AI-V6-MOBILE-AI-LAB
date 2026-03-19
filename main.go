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

// --- [🛡️ PROTOCOL: DARK-RELAY FUSION v3.5] ---
// ระบบหลอมรวม: Backend Scheduler + Frontend UI
func logIdentity() {
	fmt.Println("🐅 --- ThitNueaHub: Dark-Relay Fusion Active ---")
	fmt.Println("🚀 System Status: IGNITE V7 | Engine: Go | Cloud: Run")
}

// ระบบตั้งเวลาพ่นคอนเทนต์ (Gas Station)
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

		fmt.Printf("😴 [Gas Station]: พักเครื่อง.. รอบถัดไปคือ %s\n", nextRun.Format("15:04:05"))
		time.Sleep(time.Until(nextRun))

		// 🏁 START RELAY SEQUENCE
		err := darkRelayExecution()
		if err != nil {
			log.Printf("❌ Snake Nudge Recall: %v", err)
			time.Sleep(5 * time.Minute) 
			darkRelayExecution()
		}
	}
}

func darkRelayExecution() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	tgToken := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("CHAT_ID")
	if apiKey == "" { return fmt.Errorf("secure credentials missing") }

	prompt := "Task: Generate an elite tech insight for ThitNueaHub. Style: Aggressive, Zero-Garbage. Output: Thai Language."
	
	// ใช้ Header แทนการต่อ String เพื่อความปลอดภัย (Security Best Practice)
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{{"parts": []map[string]string{{"text": prompt}}}},
		"system_instruction": map[string]interface{}{"parts": []map[string]interface{}{{"text": "Dark-Relay Finisher Style: Elite Professional Thai Content Only."}}},
	})

	req, _ := http.NewRequest("POST", url+"?key="+apiKey, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// สกัดเอาเนื้อหาออกมา
	candidates := result["candidates"].([]interface{})
	content := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	finalText := parts[0].(map[string]interface{})["text"].(string)

	if tgToken != "" {
		sendTelegram(tgToken, chatID, "🛡️ **DARK-RELAY REPORT**\n\n"+finalText)
	}
	return nil
}

func sendTelegram(token, chatID, text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text, "parse_mode": "Markdown"})
	http.Post(url, "application/json", bytes.NewBuffer(payload))
}

func main() {
	logIdentity()
	go runScheduler() // รันระบบ AI เบื้องหลัง

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// --- [🌐 FRONTEND SERVING] ---
	// ให้บริการไฟล์หน้าเว็บจากโฟลเดอร์ /web
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// ถ้าเป็นไฟล์อื่นๆ เช่น .png ให้ไปหาในโฟลเดอร์ web
			http.FileServer(http.Dir("web")).ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "web/index.html")
	})

	fmt.Printf("🚪 ThitNueaHub Gate Open on Port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
