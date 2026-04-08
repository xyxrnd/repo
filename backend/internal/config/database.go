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
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS dosen_pembimbing_2 TEXT DEFAULT ''`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS kata_kunci TEXT DEFAULT ''`,
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
		// Index untuk unique view per IP per hari
		`CREATE INDEX IF NOT EXISTS idx_document_views_unique_daily ON document_views(document_id, ip_address, (viewed_at::date))`,
		// Migrasi: bersihkan duplikat view lama lalu buat unique constraint
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM site_settings WHERE key = 'views_dedup_v1') THEN
				-- Hapus duplikat, sisakan 1 row per document_id+ip_address+date
				DELETE FROM document_views a USING document_views b
				WHERE a.id > b.id
				AND a.document_id = b.document_id
				AND a.ip_address = b.ip_address
				AND a.viewed_at::date = b.viewed_at::date;
				-- Recalculate view_count agar akurat
				UPDATE documents SET view_count = (
					SELECT COUNT(*) FROM document_views WHERE document_id = documents.id
				);
				INSERT INTO site_settings (key, value) VALUES ('views_dedup_v1', 'done')
				ON CONFLICT (key) DO NOTHING;
			END IF;
		END $$`,
		// Unique constraint untuk mencegah duplikat view per IP per hari (race condition safe)
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_document_views_unique_constraint ON document_views(document_id, ip_address, (viewed_at::date))`,
		// Tambah kolom view_count di documents untuk cache jumlah views
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS view_count INT DEFAULT 0`,

		// Tambah kolom abstrak di documents untuk menyimpan abstrak/ringkasan dokumen
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS abstrak TEXT DEFAULT ''`,
		// Tabel site_visits untuk tracking pengunjung website unik
		`CREATE TABLE IF NOT EXISTS site_visits (
			id UUID PRIMARY KEY,
			ip_address VARCHAR(100) DEFAULT '',
			visited_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_site_visits_ip_date ON site_visits(ip_address, (visited_at::date))`,
		// Tabel site_settings untuk menyimpan pengaturan situs (key-value)
		`CREATE TABLE IF NOT EXISTS site_settings (
			key VARCHAR(100) PRIMARY KEY,
			value TEXT DEFAULT '',
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		// Reset semua data view lama (hanya sekali, dicek via marker di site_settings)
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM site_settings WHERE key = 'views_reset_v1') THEN
				TRUNCATE TABLE document_views;
				UPDATE documents SET view_count = 0;
				INSERT INTO site_settings (key, value) VALUES ('views_reset_v1', 'done')
				ON CONFLICT (key) DO NOTHING;
			END IF;
		END $$`,
		// Migrasi: reset site_visits (data lama korup karena bug port di IP)
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM site_settings WHERE key = 'site_visits_reset_v1') THEN
				TRUNCATE TABLE site_visits;
				INSERT INTO site_settings (key, value) VALUES ('site_visits_reset_v1', 'done')
				ON CONFLICT (key) DO NOTHING;
			END IF;
		END $$`,
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
		// Migrasi: ubah file_id menjadi nullable (sistem token sekarang per-dokumen, bukan per-file)
		`ALTER TABLE access_requests ALTER COLUMN file_id DROP NOT NULL`,
		// Tabel student_registrations untuk pendaftaran mahasiswa
		`CREATE TABLE IF NOT EXISTS student_registrations (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			ktm_path VARCHAR(500) DEFAULT '',
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_student_registrations_email ON student_registrations(email)`,
		`CREATE INDEX IF NOT EXISTS idx_student_registrations_status ON student_registrations(status)`,
		// Update role constraint di users untuk mendukung role 'mahasiswa'
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check`,
		`ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user', 'mahasiswa'))`,
		// Tabel email_otps untuk verifikasi email sebelum submit access request
		`CREATE TABLE IF NOT EXISTS email_otps (
			id UUID PRIMARY KEY,
			email VARCHAR(255) NOT NULL,
			otp_code VARCHAR(10) NOT NULL,
			document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			is_verified BOOLEAN DEFAULT FALSE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_email_otps_email ON email_otps(email)`,
		`CREATE INDEX IF NOT EXISTS idx_email_otps_expires ON email_otps(expires_at)`,
		// Migrasi: ubah document_id di email_otps menjadi nullable (untuk OTP registrasi mahasiswa)
		`ALTER TABLE email_otps ALTER COLUMN document_id DROP NOT NULL`,
		`ALTER TABLE email_otps DROP CONSTRAINT IF EXISTS email_otps_document_id_fkey`,
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
