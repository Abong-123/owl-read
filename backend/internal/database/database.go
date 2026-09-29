package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Book merepresentasikan satu baris di tabel books.
type Book struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Year       *int      `json:"year,omitempty"`
	Synopsis   string    `json:"synopsis"`
	Genre      string    `json:"genre"`
	CoverType  string    `json:"cover_type"`  // "uploaded" | "generated"
	CoverPath  string    `json:"cover_path"`  // relatif, mis. "cover/xxx.png"
	CoverColor string    `json:"cover_color"` // hex, mis. "#aabbcc"
	CoverText  string    `json:"cover_text"`  // maks 18 char
	PDFPath    string    `json:"pdf_path"`    // relatif, mis. "book/xxx.pdf"
	CreatedAt  time.Time `json:"created_at"`
}

// DB adalah wrapper tipis di atas *sql.DB.
type DB struct {
	conn *sql.DB

	// Root folder data (berisi pustaka.db, cover/, book/).
	DataDir string
}

// Config untuk membuka database.
type Config struct {
	// Path file SQLite, mis. "database/pustaka.db".
	Path string
	// Root folder data (untuk resolve cover/ dan book/).
	DataDir string
}

// Open membuka koneksi SQLite dan menjalankan migrasi.
func Open(cfg Config) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return nil, fmt.Errorf("buat folder db: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, "cover"), 0o755); err != nil {
		return nil, fmt.Errorf("buat folder cover: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, "book"), 0o755); err != nil {
		return nil, fmt.Errorf("buat folder book: %w", err)
	}

	// _pragma untuk WAL + foreign keys + busy timeout.
	dsn := cfg.Path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("buka sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1) // SQLite: satu writer, cukup.
	conn.SetConnMaxLifetime(0)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{conn: conn, DataDir: cfg.DataDir}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrasi: %w", err)
	}
	return db, nil
}

// Close menutup koneksi.
func (db *DB) Close() error { return db.conn.Close() }

// migrate membuat tabel kalau belum ada.
func (db *DB) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS books (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    title         TEXT NOT NULL,
    author        TEXT NOT NULL DEFAULT '',
    year          INTEGER,
    synopsis      TEXT NOT NULL DEFAULT '',
    genre         TEXT NOT NULL CHECK(genre IN ('novel','comic','majalah','non-fiksi')),

    cover_type    TEXT NOT NULL DEFAULT 'generated'
                  CHECK(cover_type IN ('uploaded','generated')),
    cover_path    TEXT NOT NULL DEFAULT '',
    cover_color   TEXT NOT NULL DEFAULT '',
    cover_text    TEXT NOT NULL DEFAULT '',

    pdf_path      TEXT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);
CREATE INDEX IF NOT EXISTS idx_books_created ON books(created_at DESC);
`
	_, err := db.conn.Exec(schema)
	return err
}

// ---------- CRUD ----------

var ErrNotFound = errors.New("buku tidak ditemukan")

// ListBooks mengembalikan semua buku, terbaru dulu.
func (db *DB) ListBooks() ([]Book, error) {
	rows, err := db.conn.Query(`
		SELECT id, title, author, year, synopsis, genre,
		       cover_type, cover_path, cover_color, cover_text,
		       pdf_path, created_at
		FROM books
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Book
	for rows.Next() {
		var b Book
		var year sql.NullInt64
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &year, &b.Synopsis, &b.Genre,
			&b.CoverType, &b.CoverPath, &b.CoverColor, &b.CoverText,
			&b.PDFPath, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		if year.Valid {
			y := int(year.Int64)
			b.Year = &y
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBook mengambil satu buku berdasarkan id.
func (db *DB) GetBook(id int64) (*Book, error) {
	row := db.conn.QueryRow(`
		SELECT id, title, author, year, synopsis, genre,
		       cover_type, cover_path, cover_color, cover_text,
		       pdf_path, created_at
		FROM books WHERE id = ?
	`, id)

	var b Book
	var year sql.NullInt64
	err := row.Scan(
		&b.ID, &b.Title, &b.Author, &year, &b.Synopsis, &b.Genre,
		&b.CoverType, &b.CoverPath, &b.CoverColor, &b.CoverText,
		&b.PDFPath, &b.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if year.Valid {
		y := int(year.Int64)
		b.Year = &y
	}
	return &b, nil
}

// CreateBook menyimpan buku baru, mengembalikan id.
func (db *DB) CreateBook(b *Book) (int64, error) {
	res, err := db.conn.Exec(`
		INSERT INTO books
		  (title, author, year, synopsis, genre,
		   cover_type, cover_path, cover_color, cover_text, pdf_path)
		VALUES (?,?,?,?,?,?,?,?,?,?)
	`,
		b.Title, b.Author, b.Year, b.Synopsis, b.Genre,
		b.CoverType, b.CoverPath, b.CoverColor, b.CoverText, b.PDFPath,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateBook memperbarui buku berdasarkan id.
func (db *DB) UpdateBook(b *Book) error {
	res, err := db.conn.Exec(`
		UPDATE books SET
		  title=?, author=?, year=?, synopsis=?, genre=?,
		  cover_type=?, cover_path=?, cover_color=?, cover_text=?,
		  pdf_path=?
		WHERE id=?
	`,
		b.Title, b.Author, b.Year, b.Synopsis, b.Genre,
		b.CoverType, b.CoverPath, b.CoverColor, b.CoverText,
		b.PDFPath, b.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteBook menghapus buku dan file fisiknya.
func (db *DB) DeleteBook(id int64) error {
	b, err := db.GetBook(id)
	if err != nil {
		return err
	}

	// Hapus file fisik (best-effort, jangan gagal kalau file sudah hilang).
	if b.CoverType == "uploaded" && b.CoverPath != "" {
		_ = os.Remove(filepath.Join(db.DataDir, b.CoverPath))
	}
	if b.PDFPath != "" {
		_ = os.Remove(filepath.Join(db.DataDir, b.PDFPath))
	}

	res, err := db.conn.Exec(`DELETE FROM books WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// CountBooks mengembalikan jumlah buku (dipakai di hero "N books").
func (db *DB) CountBooks() (int, error) {
	var n int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&n)
	return n, err
}
