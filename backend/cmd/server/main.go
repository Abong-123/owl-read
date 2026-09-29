package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"personal-library/backend/internal/database"
	"personal-library/backend/internal/display"
	"personal-library/backend/internal/input"
	"personal-library/backend/internal/middleware"
)

func main() {
	// ---------- 1. Tentukan working dir = root project ----------
	root, err := findProjectRoot()
	if err != nil {
		log.Fatalf("tidak bisa menemukan root project: %v", err)
	}
	log.Printf("project root: %s", root)

	dataDir := filepath.Join(root, "database")
	frontendDir := filepath.Join(root, "frontend")

	// ---------- 2. Buka database ----------
	db, err := database.Open(database.Config{
		Path:    filepath.Join(dataDir, "pustaka.db"),
		DataDir: dataDir,
	})
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}
	defer db.Close()

	// ---------- 3. Setup handler ----------
	displayH := display.New(db)
	inputH := input.New(db)

	// ---------- 4. Setup router ----------
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.BodyLimit(middleware.MaxBodySize))
	r.Use(middleware.Sanitize())
	r.Use(middleware.CSRFProtection())

	// ---------- 5. API routes ----------

	// Endpoint untuk frontend ambil token CSRF (opsional, karena cookie sudah di-set otomatis).
	r.GET("/api/csrf", func(c *gin.Context) {
		token, _ := c.Cookie("csrf_token")
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})

	// Read-only (rate limit ringan).
	apiRead := r.Group("/api", middleware.RateLimit(120, time.Minute))
	{
		apiRead.GET("/books", displayH.ListBooks)
		apiRead.GET("/books/:id", displayH.GetBook)
		apiRead.GET("/stats", displayH.Stats)
	}

	// Write (rate limit ketat + upload code).
	apiWrite := r.Group("/api", middleware.RateLimit(20, time.Minute), middleware.RequireUploadCode())
	{
		apiWrite.POST("/books", inputH.CreateBook)
		apiWrite.PUT("/books/:id", inputH.UpdateBook)
		apiWrite.DELETE("/books/:id", inputH.DeleteBook)
	}

	// ---------- 6. Static files (cover & PDF) ----------
	r.Static("/cover", filepath.Join(dataDir, "cover"))
	r.Static("/book", filepath.Join(dataDir, "book"))

	// ---------- 7. Serve frontend (Svelte build) ----------
	// SvelteKit dengan adapter-static menghasilkan folder "build".
	svelteBuild := filepath.Join(frontendDir, "build")

	if _, err := os.Stat(svelteBuild); err == nil {
		serveFrontend(r, svelteBuild)
	} else {
		log.Printf("⚠️ frontend build belum ada di %s — jalankan `npm run build` dulu", svelteBuild)
		log.Printf("   Untuk sekarang hanya API yang aktif.")
	}

	// ---------- 8. Run ----------
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("server jalan di http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

// ============================================================
// SERVE FRONTEND (Svelte static build)
// ============================================================

// serveFrontend menyajikan hasil build SvelteKit (adapter-static).
//
// - File statis di-serve langsung.
// - Route yang tidak ketemu → fallback ke index.html (SPA mode).
// - File di /api, /cover, /book tidak boleh fallback.
func serveFrontend(r *gin.Engine, buildDir string) {
	log.Printf("📦 serve frontend dari %s", buildDir)

	indexPath := filepath.Join(buildDir, "index.html")

	// Serve asset statis (termasuk _app/, favicon, dll).
	r.Static("/_app", filepath.Join(buildDir, "_app"))
	r.StaticFile("/favicon.png", filepath.Join(buildDir, "favicon.png"))
	r.StaticFile("/favicon.ico", filepath.Join(buildDir, "favicon.ico"))
	r.StaticFile("/robots.txt", filepath.Join(buildDir, "robots.txt"))

	// NoRoute = fallback untuk route SPA.
	// Cek file statis dulu, kalau ada serve langsung.
	// Kalau tidak ada, return index.html supaya SvelteKit router yang handle.
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path

		// Jangan fallback untuk API atau file upload.
		if len(p) >= 4 && (p[:4] == "/api" || p[:5] == "/book" || p[:6] == "/cover") {
			c.JSON(http.StatusNotFound, gin.H{"error": "endpoint tidak ditemukan"})
			return
		}

		// Cek file statis (mis. /input, /read/1 → kalau ada file literal).
		candidate := filepath.Join(buildDir, filepath.Clean(p))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			c.File(candidate)
			return
		}

		// Fallback ke index.html.
		c.File(indexPath)
	})
}

// ============================================================
// FIND PROJECT ROOT
// ============================================================

// findProjectRoot menelusuri dari working dir ke atas sampai menemukan
// folder yang berisi "go.mod" DAN "database".
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if fileExists(filepath.Join(dir, "go.mod")) && dirExists(filepath.Join(dir, "database")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
