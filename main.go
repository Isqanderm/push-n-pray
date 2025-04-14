package main

import (
	"bytes"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"net/http"
	"os"
	"sync"
)

type Candle struct {
	ID      string
	Message string
}

var (
	candles = make(map[string]Candle)
	lock    sync.RWMutex
)

func generateCandleImage(candle Candle) ([]byte, error) {
	const width = 160
	const height = 60
	const frames = 8
	const delay = 8

	palette := []color.Color{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 215, 0, 255},
		color.RGBA{255, 140, 0, 255},
		color.RGBA{255, 255, 0, 255},
		color.RGBA{0x33, 0x33, 0x33, 255},
	}

	var images []*image.Paletted
	var delays []int

	for i := 0; i < frames; i++ {
		img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
		draw.Draw(img, img.Bounds(), &image.Uniform{palette[0]}, image.Point{}, draw.Src)

		candleX := width/2 - 2
		candleY := height - 20
		for y := 0; y < 10; y++ {
			for x := 0; x < 5; x++ {
				img.SetColorIndex(candleX+x, candleY+y, 1)
			}
		}

		t := float64(i) / float64(frames)
		flickerX := math.Sin(t*6*math.Pi) * 1.5
		flickerY := math.Cos(t*3*math.Pi) * 0.8
		cx := float64(width / 2)
		cy := float64(height - 22)

		for y := -4; y <= 4; y++ {
			for x := -4; x <= 4; x++ {
				dx := float64(x) + flickerX
				dy := float64(y) + flickerY
				if dx*dx+dy*dy < 10 {
					xi := int(cx + dx)
					yi := int(cy + dy)
					if xi >= 0 && yi >= 0 && xi < width && yi < height {
						img.SetColorIndex(xi, yi, 2)
					}
				}
			}
		}
		for y := -2; y <= 2; y++ {
			for x := -2; x <= 2; x++ {
				dx := float64(x) + flickerX/2
				dy := float64(y) + flickerY/2
				if dx*dx+dy*dy < 3 {
					xi := int(cx + dx)
					yi := int(cy + dy)
					if xi >= 0 && yi >= 0 && xi < width && yi < height {
						img.SetColorIndex(xi, yi, 3)
					}
				}
			}
		}

		dc := gg.NewContextForImage(img)
		dc.SetRGB255(100, 100, 100)
		_ = dc.LoadFontFace("assets/font.ttf", 12)
		dc.DrawStringAnchored("Blessed", float64(width/2), 12, 0.5, 0.5)

		images = append(images, imageToPaletted(dc.Image(), palette))
		delays = append(delays, delay)
	}

	outGif := &gif.GIF{
		Image:     images,
		Delay:     delays,
		LoopCount: 0,
	}

	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, outGif)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func imageToPaletted(img image.Image, palette color.Palette) *image.Paletted {
	paletted := image.NewPaletted(img.Bounds(), palette)
	draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.Point{})
	return paletted
}

func main() {
	r := gin.Default()
	r.Static("/static", "./static")

	// Подгружаем все HTML-шаблоны
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	// Главная страница с формой
	r.GET("/", func(c *gin.Context) {
		tmpl.ExecuteTemplate(c.Writer, "index.html", nil)
	})

	// Создание свечки (hx-post)
	r.POST("/candles", func(c *gin.Context) {
		message := c.PostForm("message")

		id := uuid.NewString()
		candle := Candle{
			ID:      id,
			Message: message,
		}

		lock.Lock()
		candles[id] = candle
		lock.Unlock()

		tmpl.ExecuteTemplate(c.Writer, "candle.html", map[string]string{
			"ID":       id,
			"ImageURL": "/candles/" + id + "/image",
		})
	})

	r.GET("/candles/:id/image", func(c *gin.Context) {
		id := c.Param("id")

		lock.RLock()
		candle, ok := candles[id]
		lock.RUnlock()

		if !ok {
			c.String(http.StatusNotFound, "Свечка не найдена")
			return
		}

		img, err := generateCandleImage(candle)
		if err != nil {
			c.String(http.StatusInternalServerError, "Ошибка генерации картинки: %v", err)
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
