package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// โครงสร้างสำหรับคุยกับ AI
type AIRequest struct {
	Prompt string `json:"prompt"`
}

type AIResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// 1. หน้าหลัก (เสิร์ฟ index.html)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "index.html")
	})

	// 2. ดึงรูปเสือขาว
	http.HandleFunc("/image_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "image_0.png")
	})

	// 🧠 3. ระบบสมองกล (AI Endpoint) - เพิ่มใหม่!
	http.HandleFunc("/ask-kaewta", handleAI)

	fmt.Printf("🚀 ThitNueaHub AI Engine Starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleAI(w http.ResponseWriter, r *http.Request) {
	var aiReq AIRequest
	json.NewDecoder(r.Body).Decode(&aiReq)

	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/meta/llama-3-8b-instruct", accountID)

	payload, _ := json.Marshal(map[string]string{"prompt": aiReq.Prompt})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var aiResp AIResponse
	json.Unmarshal(body, &aiResp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": aiResp.Result.Response})
}
