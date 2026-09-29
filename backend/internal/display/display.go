package display

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"personal-library/backend/internal/database"
)

// Handler membungkus dependency yang dibutuhkan handler display.
type Handler struct {
	DB *database.DB
}

// New membuat Handler baru.
func New(db *database.DB) *Handler {
	return &Handler{DB: db}
}

// ============================================================
// GET /api/books
// ============================================================

// ListBooks mengembalikan semua buku (untuk shelf di frontend).
//
// Response:
//
//	{
//	  "total": 24,
//	  "books": [ { ...book }, ... ]
//	}
func (h *Handler) ListBooks(c *gin.Context) {
	books, err := h.DB.ListBooks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengambil daftar buku",
		})
		return
	}
	if books == nil {
		books = []database.Book{} // biar JSON-nya [] bukan null
	}
	c.JSON(http.StatusOK, gin.H{
		"total": len(books),
		"books": books,
	})
}

// ============================================================
// GET /api/books/:id
// ============================================================

// GetBook mengembalikan detail satu buku.
func (h *Handler) GetBook(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id tidak valid",
		})
		return
	}

	book, err := h.DB.GetBook(id)
	if errors.Is(err, database.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "buku tidak ditemukan",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengambil buku",
		})
		return
	}

	c.JSON(http.StatusOK, book)
}

// ============================================================
// GET /api/stats
// ============================================================

// Stats mengembalikan statistik ringan (jumlah buku).
//
// Dipakai untuk badge di hero: "24 books".
func (h *Handler) Stats(c *gin.Context) {
	n, err := h.DB.CountBooks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal menghitung buku",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": n})
}

// ============================================================
// HELPER
// ============================================================

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
