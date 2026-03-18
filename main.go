package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// กำหนด Port สำหรับ Cloud Run
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Serve Static Files (เพื่อให้ดึงรูปภาพขึ้นมาโชว์ได้)
	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handler สำหรับหน้าแรก
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// ป้องกันปัญหาเบราว์เซอร์ดาวน์โหลดไฟล์ แทนที่จะแสดงผล
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "index.html")
	})

	// เพื่อให้หน้าเว็บดึงรูปภาพจากโฟลเดอร์เดียวกันได้เลย
	http.HandleFunc("/7873.jpg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "7873.jpg")
	})

	fmt.Printf("ThitNueaHub Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

