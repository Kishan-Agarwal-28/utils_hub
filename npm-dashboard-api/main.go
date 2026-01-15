package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	_ = godotenv.Load()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	
	// Rate Limit: 100 reqs/min (NPM API is generous but let's be safe)
	r.Use(httprate.Limit(100, 1*time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP)))
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}}))

	r.Get("/", documentationHandler)
	
	r.Get("/api/npm-stats", fetcherHandler)

	port := os.Getenv("PORT")
	if port == "" { port = "8006" }
	
	log.Printf("📦 NPM Server running on port %s", port)
	http.ListenAndServe(":"+port, r)
}

func fetcherHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	format := r.URL.Query().Get("format")
	if format == "" { format = "svg" }

	if username == "" {
		http.Error(w, "Missing 'username' parameter", 400)
		return
	}

	// 1. Fetch NPM Data
	data, err := FetchNPMData(username)
	if err != nil {
		log.Printf("NPM Fetch Error: %v", err)
		http.Error(w, "Failed to fetch NPM data", 500)
		return
	}

	// 2. Render
	var buf bytes.Buffer
	renderer := NewRenderer(&buf)
	renderer.Render(data)

	// 3. Output
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=7200")
		renderPNG(w, buf.Bytes())
	} else {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=7200")
		w.Write(buf.Bytes())
	}
}

// Reuse the exact same PNG renderer from the GitHub project
func renderPNG(w http.ResponseWriter, svgData []byte) error {
	icon, _ := oksvg.ReadIconStream(bytes.NewReader(svgData))
	wInt, hInt := int(icon.ViewBox.W), int(icon.ViewBox.H)
	icon.SetTarget(0, 0, float64(wInt), float64(hInt))
	rgba := image.NewRGBA(image.Rect(0, 0, wInt, hInt))
	scanner := rasterx.NewScannerGV(wInt, hInt, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(wInt, hInt, scanner)
	icon.Draw(raster, 1.0)
	return png.Encode(w, rgba)
}