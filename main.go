package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// สั่งให้ระบบไปอ่านไฟล์จากโฟลเดอร์เว็บ
	fs := http.FileServer(http.Dir("./")) // ถ้าไฟล์ index.html อยู่ข้างนอกสุด
	// หรือใช้ http.Dir("./web") ถ้าพี่เอาไฟล์ไว้ในโฟลเดอร์ web
	
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🐅 ThitNueaHub Ignite on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

