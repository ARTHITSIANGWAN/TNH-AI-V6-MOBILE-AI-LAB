package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

// โครงสร้างรับข้อมูลจาก AI
type AIResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
}

// ฟังก์ชันหัวใจหลัก: โยน Prompt ให้ Llama-3 ประมวลผล
func askKaewtaAI(prompt string) (string, error) {
	// ดึงข้อมูลสำคัญจาก Environment (อย่าลืมไปใส่ใน GitHub Secrets นะคะบอส!)
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")
	model := "@cf/meta/llama-3-8b-instruct" // รุ่นท็อปที่แก้วตาแนะนำ

	url := "https://api.cloudflare.com/client/v4/accounts/" + accountID + "/ai/run/" + model

	payload, _ := json.Marshal(map[string]string{
		"prompt": prompt,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "ระบบ AI ขัดข้องค่ะบอส!", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var aiResp AIResponse
	json.Unmarshal(body, &aiResp)

	if aiResp.Success {
		return aiResp.Result.Response, nil
	}
	return "แก้วตาประมวลผลพลาดไปนิดค่ะบอส!", nil
}
