// main.go - thitnueahub mobile ai lab engine
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/kv"
)

// 💎 [5 ส: สะอาด/สะดวก] - ใช้ Sync.Pool เพื่อทำ Zero Garbage 2000%
// ไม่สร้าง Object ใหม่พร่ำเพรื่อ แต่ใช้การ Reuse ของเก่ามาล้างใหม่
var responsePool = sync.Pool{
	New: func() interface{} {
		return &TNHResponse{}
	},
}

// โครงสร้างสำหรับคุยกับ AI
type AIRequest struct {
	Prompt string `json:"prompt"`
}

type AIResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
}

type TNHResponse struct {
	Status      string `json:"status"`
	WorkerID    string `json:"worker_id"`
	WorkpointID string `json:"workpoint_id"`
	Message     string `json:"message"`
}

// 🛡️ [ครอบคอวอยอ: Core-Worker-Yield] - โครงสร้างควบคุมหลัก
type TNHCore struct {
	KVNamespace string
	WorkerID    string
}

func (c *TNHCore) Execute(ctx context.Context, r *http.Request) *TNHResponse {
	// [สะสาง] - เคลียร์ค่าจาก Pool ก่อนใช้งาน
	res := responsePool.Get().(*TNHResponse)
	
	// Logic: Zero Garbage Engine Activated
	res.Status = "success"
	res.WorkerID = c.WorkerID
	res.WorkpointID = fmt.Sprintf("tnh-%d", r.ContentLength)
	res.Message = "zero garbage engine: 2000% activated. ready go!"

	// [สุขลักษณะ] - บันทึกลง KV แบบสะอาด (ใช้ชื่อ id ที่ถูกต้องของบอส)
	key := fmt.Sprintf("status:%s", c.WorkerID)
	_ = kv.Namespace("KV").Put(ctx, key, []byte(res.Status), nil)

	return res
}

// 🏗️ [สะสาง/สะดวก] - Handler หลักที่คุมทุกอย่าง (Cloudflare Worker Interface)
type TNHHandler struct{}

func (h *TNHHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 🧠 ตรวจสอบว่าเป็น Path สำหรับ AI หรือไม่
	if r.URL.Path == "/ask-kaewta" {
		handleAI(w, r)
		return
	}

	// [สร้างนิสัย] - กำหนดค่า Core ครั้งเดียว (เช็คตัว 'a' ให้แล้วครับบอส!)
	core := &TNHCore{
		KVNamespace: "3d180da092ea4cfeb70c027d7ab11098", 
		WorkerID:    "empire_mobile_agent_001",
	}

	// [Yield] - ปล่อยให้ Core ทำงานแล้วรับผลลัพธ์
	result := core.Execute(ctx, r)

	// ส่ง JSON กลับไป
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	// [สะอาด] - คืนของเข้า Pool (Zero Garbage)
	responsePool.Put(result)
}

func handleAI(w http.ResponseWriter, r *http.Request) {
	var aiReq AIRequest
	if err := json.NewDecoder(r.Body).Decode(&aiReq); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/meta/llama-3-8b-instruct", accountID)

	payload, _ := json.Marshal(map[string]string{"prompt": aiReq.Prompt})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "ai connection error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var aiResp AIResponse
	json.Unmarshal(body, &aiResp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": aiResp.Result.Response})
}

func main() {
	// [สะดวก] - Start the Engine on Cloudflare!
	fmt.Println("thitnueahub: zero garbage 2000% - ready go!")
	workers.Serve(&TNHHandler{})
}
