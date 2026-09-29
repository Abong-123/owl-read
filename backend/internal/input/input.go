package input

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"personal-library/backend/internal/database"
)

// ============================================================
// KONSTANTA
// ============================================================

const (
	maxPDFSize   = 50 << 20 // 50 MiB
	maxCoverSize = 5 << 20  // 5 MiB
	maxCoverText = 18
)

var (
	validGenres = map[string]bool{
		"novel":     true,
		"comic":     true,
		"majalah":   true,
		"non-fiksi": true,
	}

	// hexColorRegex: #rgb atau #rrggbb (case-insensitive).
	hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

	// pdfMime: beberapa browser kirim mime berbeda.
	pdfMimes = map[string]bool{
		"application/pdf":   true,
		"application/x-pdf": true,
	}

	// imageMimes.
	imageMimes = map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/jpg":  ".jpg",
	}
)

// ============================================================
// HANDLER
// ============================================================

type Handler struct {
	DB *database.DB
}

func New(db *database.DB) *Handler {
	return &Handler{DB: db}
}

// ============================================================
// POST /api/books
// ============================================================

// CreateBook menerima multipart/form-data, validasi, simpan file, insert DB.
//
// Field form:
//
//	title        (string, wajib)
//	author       (string, opsional)
//	year         (int, opsional)
//	synopsis     (string, opsional)
//	genre        (string, wajib) salah satu: novel|comic|majalah|non-fiksi
//
//	cover_type   (string, wajib) "uploaded" | "generated"
//	  kalau "uploaded":
//	     cover_file  (file, wajib) png/jpg/jpeg ≤ 5 MB
//	  kalau "generated":
//	     cover_color (string, wajib) hex #rgb atau #rrggbb
//	     cover_text  (string, wajib) ≤ 18 karakter
//
//	pdf_file     (file, wajib) application/pdf ≤ 50 MB
func (h *Handler) CreateBook(c *gin.Context) {
	form, err := parseForm(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.buildBookFromForm(c, form, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.DB.CreateBook(book)
	if err != nil {
		// Rollback file yang sudah disimpan.
		cleanupFiles(h.DB.DataDir, book)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan buku"})
		return
	}
	book.ID = id

	created, err := h.DB.GetBook(id)
	if err != nil {
		c.JSON(http.StatusOK, book)
		return
	}
	c.JSON(http.StatusCreated, created)
}

// ============================================================
// PUT /api/books/:id
// ============================================================

// UpdateBook memperbarui buku. File baru opsional.
//
// Kalau tidak upload file baru, path lama dipertahankan.
func (h *Handler) UpdateBook(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	existing, err := h.DB.GetBook(id)
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "buku tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil buku"})
		return
	}

	form, err := parseForm(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.buildBookFromForm(c, form, existing)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	book.ID = id

	// Deteksi file lama yang diganti → hapus setelah update sukses.
	oldFilesToRemove := collectReplacedFiles(existing, book)

	if err := h.DB.UpdateBook(book); err != nil {
		// Rollback file baru kalau update gagal.
		cleanupFiles(h.DB.DataDir, book)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui buku"})
		return
	}

	// Update sukses → hapus file lama yang sudah tidak dipakai.
	for _, rel := range oldFilesToRemove {
		_ = os.Remove(filepath.Join(h.DB.DataDir, rel))
	}

	updated, err := h.DB.GetBook(id)
	if err != nil {
		c.JSON(http.StatusOK, book)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// ============================================================
// DELETE /api/books/:id
// ============================================================

func (h *Handler) DeleteBook(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	if err := h.DB.DeleteBook(id); errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "buku tidak ditemukan"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus buku"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "dihapus"})
}

// ============================================================
// BUILD BOOK DARI FORM
// ============================================================

// buildBookFromForm:
//   - Kalau existing != nil → mode update.
//     Field kosong di form → pakai nilai existing.
//     File kosong di form  → pakai path existing.
//   - Kalau existing == nil → mode create.
//     Wajib isi semua field esensial.
func (h *Handler) buildBookFromForm(
	c *gin.Context,
	form *multipart.Form,
	existing *database.Book,
) (*database.Book, error) {

	b := &database.Book{}
	isUpdate := existing != nil

	// ---------- 1. Field teks ----------
	get := func(key string) string {
		if vs, ok := form.Value[key]; ok && len(vs) > 0 {
			return strings.TrimSpace(vs[0])
		}
		return ""
	}

	b.Title = get("title")
	if b.Title == "" && isUpdate {
		b.Title = existing.Title
	}
	if b.Title == "" {
		return nil, errors.New("judul buku wajib diisi")
	}
	if len(b.Title) > 300 {
		return nil, errors.New("judul maksimal 300 karakter")
	}

	b.Author = get("author")
	if b.Author == "" && isUpdate {
		b.Author = existing.Author
	}
	if len(b.Author) > 200 {
		return nil, errors.New("nama penulis maksimal 200 karakter")
	}

	b.Synopsis = get("synopsis")
	if b.Synopsis == "" && isUpdate {
		b.Synopsis = existing.Synopsis
	}
	if len(b.Synopsis) > 5000 {
		return nil, errors.New("sinopsis maksimal 5000 karakter")
	}

	// Year opsional.
	yearStr := get("year")
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil || y < -3000 || y > 9999 {
			return nil, errors.New("tahun tidak valid")
		}
		b.Year = &y
	} else if isUpdate {
		b.Year = existing.Year
	}

	// Genre wajib.
	b.Genre = get("genre")
	if b.Genre == "" && isUpdate {
		b.Genre = existing.Genre
	}
	if !validGenres[b.Genre] {
		return nil, errors.New("genre wajib dipilih: novel, comic, majalah, atau non-fiksi")
	}

	// ---------- 2. Cover ----------
	coverType := get("cover_type")
	if coverType == "" && isUpdate {
		coverType = existing.CoverType
	}
	if coverType == "" {
		return nil, errors.New("tipe cover wajib diisi: uploaded atau generated")
	}

	switch coverType {
	case "uploaded":
		b.CoverType = "uploaded"
		file, header, err := getFile(form, "cover_file")
		if err != nil {
			// Tidak upload file baru → mode update boleh pakai path lama.
			if isUpdate && existing.CoverType == "uploaded" && existing.CoverPath != "" {
				b.CoverPath = existing.CoverPath
			} else {
				return nil, errors.New("file cover wajib diupload")
			}
		} else {
			path, err := saveImage(h.DB.DataDir, file, header)
			if err != nil {
				return nil, err
			}
			b.CoverPath = path
		}

	case "generated":
		b.CoverType = "generated"
		color := get("cover_color")
		text := get("cover_text")

		if color == "" && isUpdate {
			color = existing.CoverColor
		}
		if text == "" && isUpdate {
			text = existing.CoverText
		}
		if color == "" {
			return nil, errors.New("warna cover wajib diisi (format hex, mis. #aabbcc)")
		}
		if !hexColorRegex.MatchString(color) {
			return nil, errors.New("warna cover harus format hex: #rgb atau #rrggbb")
		}
		if text == "" {
			return nil, errors.New("teks cover wajib diisi")
		}
		if len([]rune(text)) > maxCoverText {
			return nil, fmt.Errorf("teks cover maksimal %d karakter", maxCoverText)
		}

		b.CoverColor = strings.ToLower(color)
		b.CoverText = text

	default:
		return nil, errors.New("tipe cover tidak valid")
	}

	// ---------- 3. PDF ----------
	pdfFile, pdfHeader, err := getFile(form, "pdf_file")
	if err != nil {
		if isUpdate && existing.PDFPath != "" {
			b.PDFPath = existing.PDFPath
		} else {
			return nil, errors.New("file PDF wajib diupload")
		}
	} else {
		path, err := savePDF(h.DB.DataDir, pdfFile, pdfHeader)
		if err != nil {
			return nil, err
		}
		b.PDFPath = path
	}

	// ---------- 4. Field yang tidak boleh diubah user ----------
	if isUpdate {
		b.CreatedAt = existing.CreatedAt
	}

	return b, nil
}

// ============================================================
// FILE HANDLING
// ============================================================

// getFile mengambil file dari form. Return error kalau tidak ada.
func getFile(form *multipart.Form, key string) (multipart.File, *multipart.FileHeader, error) {
	files, ok := form.File[key]
	if !ok || len(files) == 0 {
		return nil, nil, errors.New("file tidak ada")
	}
	header := files[0]
	if header.Size == 0 {
		return nil, nil, errors.New("file kosong")
	}
	f, err := header.Open()
	if err != nil {
		return nil, nil, err
	}
	return f, header, nil
}

// savePDF validasi & simpan PDF. Return path relatif terhadap DataDir.
func savePDF(dataDir string, file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	if header.Size > maxPDFSize {
		return "", fmt.Errorf("ukuran PDF maksimal %d MB", maxPDFSize>>20)
	}

	// Cek mime dari 512 byte pertama.
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	mime := http.DetectContentType(buf[:n])
	if !pdfMimes[mime] && !strings.HasPrefix(mime, "application/pdf") {
		// Fallback: cek magic bytes "%PDF-".
		if n < 5 || string(buf[:5]) != "%PDF-" {
			return "", errors.New("file harus berformat PDF")
		}
	}

	// Reset posisi file ke awal.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	name := randomName() + ".pdf"
	rel := filepath.Join("book", name)
	abs := filepath.Join(dataDir, rel)

	if err := writeFile(abs, file); err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// saveImage validasi & simpan gambar cover.
func saveImage(dataDir string, file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	if header.Size > maxCoverSize {
		return "", fmt.Errorf("ukuran cover maksimal %d MB", maxCoverSize>>20)
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	mime := http.DetectContentType(buf[:n])
	ext, ok := imageMimes[mime]
	if !ok {
		return "", errors.New("cover harus PNG, JPG, atau JPEG")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	name := randomName() + ext
	rel := filepath.Join("cover", name)
	abs := filepath.Join(dataDir, rel)

	if err := writeFile(abs, file); err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// writeFile menulis isi reader ke path.
func writeFile(path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Batasi ukuran aktual (defense in depth).
	if _, err := io.Copy(out, r); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}

// randomName menghasilkan nama file unik: {unixnano}-{8 hex}.
func randomName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

// parseForm parse multipart atau urlencoded.
func parseForm(c *gin.Context) (*multipart.Form, error) {
	if err := c.Request.ParseMultipartForm(maxPDFSize + (1 << 20)); err != nil {
		return nil, errors.New("gagal membaca form")
	}
	if c.Request.MultipartForm == nil {
		return nil, errors.New("form kosong")
	}
	return c.Request.MultipartForm, nil
}

// cleanupFiles menghapus file yang baru saja disimpan (untuk rollback).
func cleanupFiles(dataDir string, b *database.Book) {
	if b.CoverType == "uploaded" && b.CoverPath != "" {
		_ = os.Remove(filepath.Join(dataDir, b.CoverPath))
	}
	if b.PDFPath != "" {
		_ = os.Remove(filepath.Join(dataDir, b.PDFPath))
	}
}

// collectReplacedFiles return path relatif file lama yang sudah diganti.
func collectReplacedFiles(old, new *database.Book) []string {
	var out []string
	if old.CoverType == "uploaded" && old.CoverPath != "" &&
		(new.CoverType != "uploaded" || new.CoverPath != old.CoverPath) {
		out = append(out, old.CoverPath)
	}
	if old.PDFPath != "" && new.PDFPath != old.PDFPath {
		out = append(out, old.PDFPath)
	}
	return out
}
