// ส่งสัญญาณเชื่อมสถาปัตยกรรมข้ามพอร์ตในเครื่องเดียว
func LinkToPythonAgent(action string, project string) {
	targetURL := "http://127.0.0.1:5000/process"
	
	// แพ็คข้อมูลตามระนาบพิมพ์เขียว V6
	payload := map[string]interface{}{
		"sender": "TNH-AI-V6-LAB",
		"action": action,
		"payload": map[string]string{
			"project": project,
			"analysis_type": "3_LEGGED_BALANCER",
		},
	}
    
	// (ใช้ฟังก์ชัน http.NewRequest แนบโทเคน SecureToken2026 ยิงเข้าพอร์ต 5000 ทันที)
}
