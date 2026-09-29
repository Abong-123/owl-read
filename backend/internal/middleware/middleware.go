package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// KONFIGURASI
// ============================================================

// UploadCode adalah 4 digit kode rahasia untuk operasi write.
// GANTI sesuai keinginanmu.
const UploadCode = "1234"

// MaxBodySize adalah batas ukuran body request (50 MB + overhead form).
const MaxBodySize = 52 << 20 // 52 MiB

// csrfCookieName adalah nama cookie untuk token CSRF.
const csrfCookieName = "csrf_token"

// csrfHeaderName adalah header tempat frontend mengirim token.
const csrfHeaderName = "X-CSRF-Token"

// ============================================================
// SECURITY HEADERS (pengganti Helmet)
// ============================================================

// SecurityHeaders menambahkan header keamanan dasar ke setiap respons.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// CSP: izinkan self + Google Fonts (Playfair & DM Sans).
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"img-src 'self' data: blob:; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src 'self' https://fonts.gstatic.com; "+
				"script-src 'self' 'unsafe-inline'; "+
				"connect-src 'self'; "+
				"frame-ancestors 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'",
		)

		c.Next()
	}
}

// ============================================================
// BODY LIMIT
// ============================================================

// BodyLimit membatasi ukuran body request agar tidak bisa di-abuse.
func BodyLimit(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}

// ============================================================
// CSRF PROTECTION (double-submit cookie)
// ============================================================

// CSRFProtection:
//   - GET  : pastikan cookie token ada (buat kalau belum).
//   - POST/PUT/DELETE: bandingkan header X-CSRF-Token dengan cookie.
//
// Frontend harus:
//  1. GET /api/csrf (atau GET halaman apa saja) → cookie ter-set.
//  2. Baca cookie `csrf_token`.
//  3. Kirim header X-CSRF-Token berisi nilai cookie saat write.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Ambil cookie existing.
		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			// Buat token baru.
			cookieToken = randomToken(32)
			// SameSite=Lax: aman untuk top-level navigation, cukup untuk personal.
			// Secure=false karena dev HTTP. Di produksi HTTPS, set true.
			c.SetCookie(
				csrfCookieName,
				cookieToken,
				60*60*24*7, // 7 hari
				"/",
				"",
				false, // Secure
				false, // HttpOnly — HARUS false agar JS bisa baca
			)
		}

		// Cek hanya untuk method write.
		if method == http.MethodPost ||
			method == http.MethodPut ||
			method == http.MethodPatch ||
			method == http.MethodDelete {

			headerToken := c.GetHeader(csrfHeaderName)
			if headerToken == "" || headerToken != cookieToken {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "CSRF token tidak valid",
				})
				return
			}
		}

		c.Next()
	}
}

// ============================================================
// UPLOAD CODE
// ============================================================

// RequireUploadCode memastikan header X-Upload-Code == UploadCode.
// Dipasang hanya di route write (POST/PUT/DELETE).
func RequireUploadCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.GetHeader("X-Upload-Code")
		if code != UploadCode {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "kode upload salah",
			})
			return
		}
		c.Next()
	}
}

// ============================================================
// RATE LIMIT (in-memory, per IP + per endpoint)
// ============================================================

type rateEntry struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
}

var limiter = &rateLimiter{entries: make(map[string]*rateEntry)}

// RateLimit membatasi N request per window per IP per route.
//
// Contoh: RateLimit(10, time.Minute) → maks 10 request/menit per IP per endpoint.
func RateLimit(max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + "|" + c.FullPath()

		limiter.mu.Lock()
		now := time.Now()
		e, ok := limiter.entries[key]
		if !ok || now.After(e.resetAt) {
			e = &rateEntry{count: 0, resetAt: now.Add(window)}
			limiter.entries[key] = e
		}
		e.count++
		remaining := max - e.count
		resetAt := e.resetAt
		limiter.mu.Unlock()

		c.Header("X-RateLimit-Limit", itoa(max))
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", itoa(remaining))
		c.Header("X-RateLimit-Reset", itoa(int(resetAt.Unix())))

		if e.count > max {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "terlalu banyak request, coba lagi nanti",
			})
			return
		}
		c.Next()
	}
}

// ============================================================
// SANITIZE
// ============================================================

// Sanitize membersihkan nilai string di form/query:
//   - Trim spasi.
//   - Hapus karakter kontrol (kecuali newline & tab untuk sinopsis).
//   - Batasi panjang maksimum.
//
// Dipasang global; hanya memodifikasi PostForm & Query.
func Sanitize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse form dulu (multipart atau urlencoded).
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			_ = c.Request.ParseMultipartForm(32 << 20)
		}
		_ = c.Request.ParseForm()

		// Bersihkan PostForm.
		for k, vs := range c.Request.PostForm {
			for i, v := range vs {
				c.Request.PostForm[k][i] = cleanString(v)
			}
		}
		// Bersihkan query.
		for k, vs := range c.Request.URL.Query() {
			for i, v := range vs {
				vs[i] = cleanString(v)
			}
			c.Request.URL.Query()[k] = vs
		}

		c.Next()
	}
}

// cleanString menghapus karakter kontrol, trim, dan batasi 5000 char.
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		// Buang karakter kontrol kecuali tab, newline, carriage return.
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			continue
		}
		if r == 0x7f { // DEL
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	if len(out) > 5000 {
		out = out[:5000]
	}
	return out
}

// ============================================================
// HELPER
// ============================================================

// randomToken menghasilkan string hex acak sepanjang n byte (2n char).
func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Fallback: pakai waktu (jarang terjadi).
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

// itoa cepat tanpa import strconv di banyak tempat.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
