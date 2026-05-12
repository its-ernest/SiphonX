package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "modernc.org/sqlite"
)

type LogEntry struct {
	ID        int    `json:"id"`
	Device    string `json:"device"`
	Payload   string `json:"payload"`
	ImagePath string `json:"image_path"`
	CreatedAt string `json:"created_at"`
}

var db *sql.DB

func main() {
	err := ReadAndPrintFile("doom.txt", "cyan")
	if err != nil {
		fmt.Println("Error:", err)
	}
	initDB()
	defer db.Close()
	os.MkdirAll("./images", 0755)

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.File("/", "index.html")
	e.Static("/images", "images")

	// GET ALL LOGS
	e.GET("/api/logs", func(c *echo.Context) error {
		rows, err := db.Query("SELECT id, device, payload, image_path, created_at FROM logs ORDER BY id DESC")
		if err != nil {
			return err
		}
		defer rows.Close()

		entries := []LogEntry{}
		for rows.Next() {
			var l LogEntry
			rows.Scan(&l.ID, &l.Device, &l.Payload, &l.ImagePath, &l.CreatedAt)
			entries = append(entries, l)
		}
		return c.JSON(http.StatusOK, entries)
	})

	// NEW: GET UNIQUE DEVICES for the filter dropdown
	e.GET("/api/devices", func(c *echo.Context) error {
		rows, err := db.Query("SELECT DISTINCT device FROM logs")
		if err != nil {
			return err
		}
		defer rows.Close()

		devices := []string{}
		for rows.Next() {
			var d string
			rows.Scan(&d)
			devices = append(devices, d)
		}
		return c.JSON(http.StatusOK, devices)
	})

	e.POST("/capture", func(c *echo.Context) error {
		var payload map[string]any
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad_json"})
		}

		device, _ := payload["device"].(string)
		if device == "" {
			device = "IOS_SHORTCUT"
		}

		imgB64, _ := payload["img_base64"].(string)
		var imgPath string
		if imgB64 != "" {
			var err error
			imgPath, err = saveBase64Image(device, imgB64)
			if err != nil {
				slog.Error("failed to save image", "err", err)
			}
		}

		rawPayload, _ := json.Marshal(payload)
		_, err := db.Exec(
			"INSERT INTO logs (device, payload, image_path, created_at) VALUES (?, ?, ?, ?)",
			device, string(rawPayload), imgPath, time.Now(),
		)
		if err != nil {
			slog.Error("DATABASE INSERT FAILED", "details", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db_fail"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.Start("0.0.0.0:8080")
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./shortcuts.db")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device TEXT,
		payload TEXT,
		image_path TEXT,
		created_at DATETIME
	);`)
}

func saveBase64Image(device, data string) (string, error) {
	dec, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	os.MkdirAll("./images/"+device, 0755)
	filename := fmt.Sprintf("%s_%d.jpg", device, time.Now().Unix())
	path := filepath.Join("images", device, filename)
	err = os.WriteFile(path, dec, 0644)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(path), nil
}
