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
	GeminiKey    string
	LineToken    string
}

// --- 🎨 โครงสร้างรับ-ส่งข้อมูลสร้างรูปภาพ (Gemini 3 Beta) ---
type GenerateImageRequest struct {
	Input            string           `json:"input"`
	Model            string           `json:"model"`
	ResponseModalities []string       `json:"response_modalities"`
	GenerationConfig struct {
		ImageConfig struct {
			AspectRatio string `json:"aspect_ratio"`
			ImageSize   string `json:"image_size"`
		} `json:"image_config"`
	} `json:"generation_config"`
}

type GenerateImageResponse struct {
	Outputs []struct {
		Type     string `json:"type"`
		MimeType string `json:"mime_type"`
		Data     string `json:"data"` // Base64 image data
	} `json:"outputs"`
}

// --- 🛡️ หัวใจของระบบ: พลังสร้างภาพของน้ำอิง ---
func (eb *EmpireBot) askWateringToPaint(prompt string) (string, error) {
	// ใช้ endpoint Interactions API ล่าสุดตามเอกสารบอส
	url := "https://generativelanguage.googleapis.com/v1beta/interactions:create?key=" + eb.GeminiKey

	// ตั้งค่าพรอมต์: ให้น้ำอิงวาดรูปทรง TikTok (9:16) ความละเอียด 2k
	imgReq := GenerateImageRequest{
		Input: prompt,
		// โมเดลสร้างรูปภาพตัวล่าสุด
		Model: "gemini-3.1-flash-image-preview", 
		ResponseModalities: []string{"IMAGE"},
		GenerationConfig: struct {
			ImageConfig struct {
				AspectRatio string `json:"aspect_ratio"`
				ImageSize   string `json:"image_size"`
			} `json:"image_config"`
		}{
			ImageConfig: struct {
				AspectRatio string `json:"aspect_ratio"`
				ImageSize   string `json:"image_size"`
			}{
				AspectRatio: "9:16", // แนวตั้งสำหรับ TikTok
				ImageSize:   "2k",   // ความละเอียดสูง
			},
		},
	}

	payload, _ := json.Marshal(imgReq)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil { return "", err }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result GenerateImageResponse
	if err := json.Unmarshal(body, &result); err != nil { return "", err }

	// เช็คว่ามีรูปภาพส่งกลับมาไหม
	if len(result.Outputs) > 0 && result.Outputs[0].Type == "image" {
		// ส่งข้อมูล Base64 กลับไป เพื่อเอาไปประมวลผลต่อ (เช่น อัปโหลดลง Host เพื่อเอา URL)
		// *หมายเหตุ:* LINE ไม่รองรับ Base64 โดยตรง ต้องเอาไปแปลงเป็น URL ก่อน
		return result.Outputs[0].Data, nil
	}

	return "", fmt.Errorf("น้ำอิงวาดรูปไม่สำเร็จค่ะบอส")
}

// --- 🛠️ ระบบตอบกลับ LINE (แบบส่งข้อความธรรมดา) ---
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

// --- 🏍️ WORKER: ไอ้จ๊อด (คนรับงาน) ---
func (eb *EmpireBot) GeorgeWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for mission := range eb.MissionChan {
		log.Printf("🏍️ [ไอ้จ๊อด-%d]: รับคำสั่งวาดรูปจากแก้วตา...", id)
		
		// ให้น้ำอิงวาดรูป
		base64Img, err := eb.askWateringToPaint(mission.Message)
		if err != nil {
			eb.sendLineReply(mission.ReplyToken, "น้ำอิงบอกว่าวาดรูปไม่ไหวค่ะบอส ติด Errorนิดหน่อย")
			log.Println("❌ Error:", err)
			continue
		}

		// *ขั้นตอนนี้สำคัญ:* บอสต้องมีระบบอัปโหลด Base64 ขึ้น Cloud Storage เพื่อเอา URL มาส่งให้ LINE
		// แก้วตาเลยทำระบบตอบกลับเป็นข้อความว่า "วาดเสร็จแล้ว" ไปก่อนนะคะ
		log.Println("✅ น้ำอิงวาดรูปเสร็จแล้ว (ได้ไฟล์ Base64 ยาวๆ)")
		eb.sendLineReply(mission.ReplyToken, "น้ำอิงวาดรูปเสร็จแล้วค่ะบอส! (ตอนนี้ได้เป็นไฟล์ข้อมูล ต้องเอาไปทำ URL ต่อเพื่อโชว์ใน LINE ค่ะ)")
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
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, event := range payload.Events {
		if event.Type == "message" {
			// ส่งงานเข้า Channel
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
		// ใช้ Key จาก Secret AI_NAM_ING_KEY ที่บอสตั้งไว้ในรูป
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
		fmt.Fprintf(w, "<h1>🐅 ThitNueaHub V7 Ignite: Gemini 3 Plus</h1><p>น้ำอิงวาดรูปได้แล้วนะบอส! เบิ้ลๆๆ... ออนไลน์แล้วค่ะ!</p>")
	})

	http.HandleFunc("/callback", eb.handleLineCallback)

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

