package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// รับ Port จาก Google Cloud
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// หน้าแรกแสดงผล index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "index.html")
	})

	// ดึงรูปเสือขาว image_0.png
	http.HandleFunc("/image_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "image_0.png")
	})

	fmt.Printf("🚀 ThitNueaHub Engine Starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
