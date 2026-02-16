package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB adalah koneksi database global
var DB *pgxpool.Pool

// ConnectDB membuat koneksi ke database PostgreSQL
func ConnectDB() {
	// Ambil DSN dari environment variable, atau gunakan default
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:rendhi123@localhost:5432/repository_db"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Gagal koneksi ke database:", err)
	}

	// Test koneksi
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Database tidak dapat dijangkau:", err)
	}

	DB = pool
	log.Println("✅ Database connected successfully")

	// Auto-migrate: buat tabel yang belum ada
	runMigrations(pool)
}

// runMigrations membuat tabel yang belum ada di database
func runMigrations(pool *pgxpool.Pool) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS fakultas (
			id UUID PRIMARY KEY,
			nama VARCHAR(255) NOT NULL,
			kode VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS prodi (
			id UUID PRIMARY KEY,
			nama VARCHAR(255) NOT NULL,
			kode VARCHAR(50) NOT NULL UNIQUE,
			fakultas_id UUID NOT NULL REFERENCES fakultas(id) ON DELETE RESTRICT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fakultas_kode ON fakultas(kode)`,
		`CREATE INDEX IF NOT EXISTS idx_prodi_kode ON prodi(kode)`,
		`CREATE INDEX IF NOT EXISTS idx_prodi_fakultas_id ON prodi(fakultas_id)`,
		// Tambah kolom baru di documents (ALTER TABLE aman jika kolom sudah ada)
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS fakultas_id UUID REFERENCES fakultas(id) ON DELETE SET NULL`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS prodi_id UUID REFERENCES prodi(id) ON DELETE SET NULL`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS dosen_pembimbing TEXT DEFAULT ''`,
		// Tabel untuk multiple files per dokumen
		`CREATE TABLE IF NOT EXISTS document_files (
			id UUID PRIMARY KEY,
			document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			file_name VARCHAR(500) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			file_size BIGINT DEFAULT 0,
			file_order INT DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_document_files_doc_id ON document_files(document_id)`,
		// Tambah kolom is_locked untuk fitur lock file
		`ALTER TABLE document_files ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT FALSE`,
		// Tabel untuk tracking views dokumen (untuk fitur dokumen populer)
		`CREATE TABLE IF NOT EXISTS document_views (
			id UUID PRIMARY KEY,
			document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			ip_address VARCHAR(100) DEFAULT '',
			viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_document_views_doc_id ON document_views(document_id)`,
		// Tambah kolom view_count di documents untuk cache jumlah views
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS view_count INT DEFAULT 0`,
		// Tambah kolom abstrak di documents untuk menyimpan abstrak/ringkasan dokumen
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS abstrak TEXT DEFAULT ''`,
		// Tabel site_settings untuk menyimpan pengaturan situs (key-value)
		`CREATE TABLE IF NOT EXISTS site_settings (
			key VARCHAR(100) PRIMARY KEY,
			value TEXT DEFAULT '',
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		// Tabel access_requests untuk permintaan akses file terkunci
		`CREATE TABLE IF NOT EXISTS access_requests (
			id UUID PRIMARY KEY,
			document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			file_id UUID NOT NULL REFERENCES document_files(id) ON DELETE CASCADE,
			nama VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			ktm_path VARCHAR(500) DEFAULT '',
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_access_requests_doc_id ON access_requests(document_id)`,
		`CREATE INDEX IF NOT EXISTS idx_access_requests_file_id ON access_requests(file_id)`,
		// Tambah kolom access_token untuk menyimpan token akses yang dikirim via email
		`ALTER TABLE access_requests ADD COLUMN IF NOT EXISTS access_token VARCHAR(100) DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_access_requests_token ON access_requests(access_token)`,
	}

	for _, q := range queries {
		_, err := pool.Exec(context.Background(), q)
		if err != nil {
			log.Printf("⚠️ Migration warning: %v", err)
		}
	}
	log.Println("✅ Database migrations completed")
}

// CloseDB menutup koneksi database
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
