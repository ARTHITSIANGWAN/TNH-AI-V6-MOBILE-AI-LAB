// Project Name: ThitNueaHub-Core-Master
// Folder Path: /backend/core/
// File Name: main.go

package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
)

// AIDispatch - โครงสร้างสื่อสารกลางระหว่าง Agents (A2A Protocol)
type AIDispatch struct {
	Sender    string                 `json:"sender"`
	Action    string                 `json:"action"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp string                 `json:"timestamp"`
}

// initDB - ตั้งค่า Storage Infrastructure (thitnuea_knowledge.db ในโหมด WAL)
func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./thitnuea_knowledge.db")
	if err != nil {
		slog.Error("❌ ไม่สามารถเปิดฐานข้อมูลได้", "err", err)
		os.Exit(1)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		slog.Error("❌ การตั้งค่า WAL Mode ล้มเหลว", "err", err)
	}

	query := `CREATE TABLE IF NOT EXISTS security_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		alert_type TEXT,
		status TEXT,
		created_at DATETIME
	);`
	
	if _, err := db.Exec(query); err != nil {
		slog.Error("❌ ไม่สามารถสร้างตารางได้", "err", err)
	}
	return db
}

func main() {
	db := initDB()
	defer db.Close()

	app := fiber.New(fiber.Config{
		AppName: "ThitNueaHub Unified Engine v2026",
	})
	
	app.Use(recover.New())

	// Health Check Endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"system":   "ThitNueaHub Unified Engine",
			"status":   "100% Operational",
			"database": "thitnuea_knowledge.db (WAL Mode)",
			"port":     2026,
		})
	})

	// Security Alert Logging Endpoint
	app.Post("/api/v1/security/log", func(c *fiber.Ctx) error {
		var body map[string]string
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Payload"})
		}

		alertType := body["alert_type"]
		status := "Dismissed"

		_, err := db.Exec("INSERT INTO security_logs (alert_type, status, created_at) VALUES (?, ?, ?)", alertType, status, time.Now())
		if err != nil {
			slog.Error("❌ บันทึก Security Log ไม่สำเร็จ", "err", err)
		} else {
			fmt.Printf("🛡️ [Security] บันทึกการจัดการแจ้งเตือน '%s' สำเร็จ!\n", alertType)
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"alert":  alertType,
			"action": status,
		})
	})

	port := ":2026"
	slog.Info("🚀 ThitNueaHub Core Engine รันบนพอร์ต", "port", port)
	if err := app.Listen(port); err != nil {
		slog.Error("❌ เซิร์ฟเวอร์หยุดทำงาน", "err", err)
		os.Exit(1)
	}
}
