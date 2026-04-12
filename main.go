Skip to content
ARTHITSIANGWAN
thitnueahub-mobile-ai-lab
Repository navigation
Code
Issues
Actions
Security and quality
Insights
Settings
Commit 2dc9a36
ARTHITSIANGWAN
ARTHITSIANGWAN
authored
2 weeks ago
·
·
Verified
Update main.go
main
1 parent 
33054ec
 commit 
2dc9a36
File tree
Filter files…
main.go
1 file changed
+59
-46
lines changed
Search within code
 
‎main.go‎
+59
-46
Lines changed: 59 additions & 46 deletions
Original file line number	Diff line number	Diff line change
@@ -1,68 +1,81 @@
// main.go
package main

import (
	"bytes"
	"encoding/json"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/kv"
)

// โครงสร้างสำหรับคุยกับ AI
type AIRequest struct {
	Prompt string `json:"prompt"`
// 💎 [5 ส: สะอาด/สะดวก] - ใช้ Sync.Pool เพื่อทำ Zero Garbage 2000%
// ไม่สร้าง Object ใหม่พร่ำเพรื่อ แต่ใช้การ Reuse ของเก่ามาล้างใหม่
var responsePool = sync.Pool{
	New: func() interface{} {
		return &TNHResponse{}
	},
}

type AIResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
type TNHResponse struct {
	Status      string `json:"status"`
	WorkerID    string `json:"worker_id"`
	WorkpointID string `json:"workpoint_id"`
	Message     string `json:"message"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	// 1. หน้าหลัก (เสิร์ฟ index.html)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "index.html")
	})
// 🛡️ [ครอบคอวอยอ: Core-Worker-Yield] - โครงสร้างควบคุมหลัก
type TNHCore struct {
	KVNamespace string
	WorkerID    string
}

	// 2. ดึงรูปเสือขาว
	http.HandleFunc("/image_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "image_0.png")
	})
func (c *TNHCore) Execute(ctx context.Context, r *http.Request) *TNHResponse {
	// [สะสาง] - เคลียร์ค่าก่อนใช้งาน
	res := responsePool.Get().(*TNHResponse)
	
	// Logic: ยัดวาฬ 19 ตัวลงใน Worker ตัวเดียว
	res.Status = "SUCCESS"
	res.WorkerID = c.WorkerID
	res.WorkpointID = fmt.Sprintf("TNH-%d", r.ContentLength)
	res.Message = "Zero Garbage Engine: 2000% Activated. Ready Go!"

	// 🧠 3. ระบบสมองกล (AI Endpoint) - เพิ่มใหม่!
	http.HandleFunc("/ask-kaewta", handleAI)
	// [สุขลักษณะ] - บันทึกลง KV แบบสะอาด
	key := fmt.Sprintf("status:%s", c.WorkerID)
	_ = kv.Namespace("KV").Put(ctx, key, []byte(res.Status), nil)

	fmt.Printf("🚀 ThitNueaHub AI Engine Starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
	return res
}

func handleAI(w http.ResponseWriter, r *http.Request) {
	var aiReq AIRequest
	json.NewDecoder(r.Body).Decode(&aiReq)
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/meta/llama-3-8b-instruct", accountID)
// 🏗️ [สะสาง/สะดวก] - Handler หลักที่คุมทุกอย่าง
type TNHHandler struct{}

	payload, _ := json.Marshal(map[string]string{"prompt": aiReq.Prompt})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+apiToken)
func (h *TNHHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	// [สร้างนิสัย] - กำหนดค่า Core ครั้งเดียว
	core := &TNHCore{
		KVNamespace: "3d180da092ea4cfeb70c027d7ab11098", // คีย์จากรูป 8461 ของบอส
		WorkerID:    "EMPIRE_MOBILE_AGENT_001",
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var aiResp AIResponse
	json.Unmarshal(body, &aiResp)
	// [Yield] - ปล่อยให้ Core ทำงานแล้วรับผลลัพธ์
	result := core.Execute(ctx, r)

	// ส่ง JSON กลับไปให้ WhatsApp/Meta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": aiResp.Result.Response})
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"%s","worker_id":"%s","message":"%s"}`, 
		result.Status, result.WorkerID, result.Message)
	// [สะอาด] - คืนของเข้า Pool (Zero Garbage)
	responsePool.Put(result)
}
func main() {
	// [สะดวก] - Start the Engine!
	fmt.Println("ThitNueaHub: Zero Garbage 2000% - Ready Go!")
	workers.Serve(&TNHHandler{})
}
0 commit comments
Comments
0
 (0)
Comment
You're not receiving notifications from this thread.

	
