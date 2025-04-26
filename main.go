package main

import (
	"bytes"
	"context"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type Candle struct {
	ID      string
	Message string
}

var (
	db *pgx.Conn
)

func wrapText(dc *gg.Context, text string, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var current string
	for _, word := range words {
		test := strings.TrimSpace(current + " " + word)
		w, _ := dc.MeasureString(test)
		if w > maxWidth && current != "" {
			lines = append(lines, current)
			current = word
		} else {
			current = test
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func generateCandleImage(candle Candle) ([]byte, error) {
	const baseWidth = 160
	const lineHeight = 14
	const padding = 8
	const fontSize = 12.0
	const candleOffset = 3
	palette := []color.Color{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 215, 0, 255},
		color.RGBA{255, 140, 0, 255},
		color.RGBA{255, 255, 0, 255},
		color.RGBA{0x33, 0x33, 0x33, 255},
	}
	msg := candle.Message
	if len(msg) > 100 {
		msg = msg[:100]
	}
	tempCtx := gg.NewContext(baseWidth, 100)
	if err := tempCtx.LoadFontFace("assets/font.ttf", fontSize); err != nil {
		log.Println("font load error:", err)
		return nil, err
	}
	lines := wrapText(tempCtx, msg, float64(baseWidth-padding*2))
	height := padding*2 + len(lines)*lineHeight + candleOffset + 13

	const frames = 8
	const delay = 8
	var images []*image.Paletted
	var delays []int

	for i := 0; i < frames; i++ {
		img := image.NewPaletted(image.Rect(0, 0, baseWidth, height), palette)
		draw.Draw(img, img.Bounds(), &image.Uniform{palette[0]}, image.Point{}, draw.Src)

		dc := gg.NewContextForImage(img)
		dc.SetRGB255(50, 50, 50)
		_ = dc.LoadFontFace("assets/font.ttf", fontSize)

		for idx, line := range lines {
			y := float64(padding + padding + idx*lineHeight)
			dc.DrawStringAnchored(line, float64(baseWidth/2), y, 0.5, 0)
		}

		// Draw candle
		candleY := padding + len(lines)*lineHeight + candleOffset
		candleX := baseWidth/2 - 2
		for y := 0; y < 10; y++ {
			for x := 0; x < 5; x++ {
				dc.SetColor(palette[1])
				dc.SetPixel(candleX+x, candleY+y)
			}
		}

		t := float64(i) / float64(frames)
		flickerX := math.Sin(t*6*math.Pi) * 1.5
		flickerY := math.Cos(t*3*math.Pi) * 0.8
		cx := float64(baseWidth / 2)
		cy := float64(candleY - 2)

		for y := -4; y <= 4; y++ {
			for x := -4; x <= 4; x++ {
				dx := float64(x) + flickerX
				dy := float64(y) + flickerY
				if dx*dx+dy*dy < 10 {
					dc.SetColor(palette[2])
					dc.SetPixel(int(cx+dx), int(cy+dy))
				}
			}
		}
		for y := -2; y <= 2; y++ {
			for x := -2; x <= 2; x++ {
				dx := float64(x) + flickerX/2
				dy := float64(y) + flickerY/2
				if dx*dx+dy*dy < 3 {
					dc.SetColor(palette[3])
					dc.SetPixel(int(cx+dx), int(cy+dy))
				}
			}
		}

		images = append(images, imageToPaletted(dc.Image(), palette))
		delays = append(delays, delay)
	}

	outGif := &gif.GIF{
		Image:     images,
		Delay:     delays,
		LoopCount: 0,
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, outGif); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func imageToPaletted(img image.Image, palette color.Palette) *image.Paletted {
	paletted := image.NewPaletted(img.Bounds(), palette)
	draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.Point{})
	return paletted
}

func initDB() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err = pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS candles (
			id UUID PRIMARY KEY,
			message TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Unable to create table: %v", err)
	}
}

func main() {
	_ = godotenv.Load()
	initDB()
	defer db.Close(context.Background())

	gin.SetMode(gin.ReleaseMode)
	domain := os.Getenv("DOMAIN")
	r := gin.Default()
	r.Static("/static", "./static")
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	r.GET("/", func(c *gin.Context) {
		tmpl.ExecuteTemplate(c.Writer, "index.html", nil)
	})

	r.POST("/candles", func(c *gin.Context) {
		message := c.PostForm("message")
		message = strings.TrimSpace(message)
		if len(message) > 100 {
			message = message[:100]
		}
		id := uuid.NewString()
		_, err := db.Exec(context.Background(), "INSERT INTO candles (id, message) VALUES ($1, $2)", id, message)
		if err != nil {
			c.String(http.StatusInternalServerError, "DB insert error")
			return
		}
		tmpl.ExecuteTemplate(c.Writer, "candle.html", map[string]interface{}{
			"ID":       id,
			"ImageURL": template.URL(domain + "/candles/" + id + "/image"),
		})
	})

	r.GET("/candles/:id/image", func(c *gin.Context) {
		id := c.Param("id")
		var message string
		err := db.QueryRow(context.Background(), "SELECT message FROM candles WHERE id=$1", id).Scan(&message)
		if err != nil {
			c.String(http.StatusNotFound, "Candle not found")
			return
		}
		candle := Candle{ID: id, Message: message}
		img, err := generateCandleImage(candle)
		if err != nil {
			c.String(http.StatusInternalServerError, "Image generation error: %v", err)
			return
		}
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("Content-Type", "image/gif")
		c.Writer.Write(img)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
