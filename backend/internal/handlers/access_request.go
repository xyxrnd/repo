package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"repository-un/internal/config"
	"repository-un/internal/middleware"
	"repository-un/internal/models"

	"github.com/google/uuid"
)

// ===== SMTP Configuration =====
// Konfigurasikan ini sesuai dengan provider email Anda
var (
	smtpHost     = getEnv("SMTP_HOST", "smtp.gmail.com")
	smtpPort     = getEnv("SMTP_PORT", "587")
	smtpEmail    = getEnv("SMTP_EMAIL", "your-email@gmail.com")
	smtpPassword = getEnv("SMTP_PASSWORD", "your-app-password")
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// generateAccessToken membuat token akses acak (16 karakter hex)
func generateAccessToken() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// sendAccessTokenEmail mengirim email yang berisi token akses
func sendAccessTokenEmail(toEmail, toName, token, documentTitle, fileName string) error {
	from := smtpEmail
	to := toEmail

	subject := "Akses File Disetujui - " + documentTitle
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc;">
  <div style="background: linear-gradient(135deg, #3b82f6, #6366f1); padding: 30px; border-radius: 16px 16px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0; font-size: 24px;">✅ Akses Disetujui</h1>
  </div>
  <div style="background: white; padding: 30px; border-radius: 0 0 16px 16px; box-shadow: 0 4px 6px rgba(0,0,0,0.05);">
    <p style="color: #334155; font-size: 16px;">Halo <strong>%s</strong>,</p>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Permintaan akses Anda untuk file <strong>%s</strong> pada dokumen <strong>%s</strong> telah <span style="color: #22c55e; font-weight: bold;">disetujui</span>.
    </p>
    <p style="color: #475569; font-size: 14px;">Gunakan token berikut untuk membuka file:</p>
    <div style="background: linear-gradient(135deg, #f0f9ff, #e0f2fe); border: 2px dashed #3b82f6; border-radius: 12px; padding: 20px; text-align: center; margin: 20px 0;">
      <p style="color: #64748b; font-size: 12px; margin: 0 0 8px 0; text-transform: uppercase; letter-spacing: 1px;">Token Akses Anda</p>
      <p style="color: #1e40af; font-size: 28px; font-weight: bold; margin: 0; letter-spacing: 3px; font-family: 'Courier New', monospace;">%s</p>
    </div>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Masukkan token ini pada halaman detail dokumen untuk membuka akses preview dan download file.
    </p>
    <hr style="border: none; border-top: 1px solid #e2e8f0; margin: 20px 0;">
    <p style="color: #94a3b8; font-size: 12px; text-align: center;">
      Email ini dikirim secara otomatis oleh sistem Repository Akademik.<br>
      Jangan bagikan token ini kepada orang lain.
    </p>
  </div>
</body>
</html>`, toName, fileName, documentTitle, token)

	// Compose email
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		from, to, subject)

	msg := []byte(headers + body)

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
	return err
}

// ===== HANDLERS =====

// AccessRequestHandler menangani permintaan akses file terkunci
// POST /api/access-requests - Submit permintaan akses (public)
// GET  /api/access-requests - List permintaan akses (admin)
func AccessRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	switch r.Method {
	case http.MethodPost:
		createAccessRequest(w, r)
	case http.MethodGet:
		listAccessRequests(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AccessRequestByIdHandler menangani operasi pada access request tertentu
// PUT /api/access-requests/:id - Update status (approve/reject) → generate token & kirim email
// DELETE /api/access-requests/:id - Hapus access request
func AccessRequestByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	id := strings.TrimPrefix(r.URL.Path, "/api/access-requests/")
	if id == "" {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		updateAccessRequestStatus(w, r, id)
	case http.MethodDelete:
		deleteAccessRequest(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// VerifyAccessTokenHandler memverifikasi token akses
// POST /api/access-requests/verify-token
// Body: { "file_id": "xxx", "token": "xxx" }
func VerifyAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		FileID string `json:"file_id"`
		Token  string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body tidak valid", http.StatusBadRequest)
		return
	}

	if body.FileID == "" || body.Token == "" {
		http.Error(w, "file_id dan token wajib diisi", http.StatusBadRequest)
		return
	}

	// Cari access request yang cocok: token benar, status approved, untuk file tersebut
	var arID, filePath, fileName string
	err := config.DB.QueryRow(context.Background(),
		`SELECT ar.id, df.file_path, df.file_name
		 FROM access_requests ar
		 JOIN document_files df ON df.id = ar.file_id
		 WHERE ar.file_id = $1 AND ar.access_token = $2 AND ar.status = 'approved'
		 LIMIT 1`,
		body.FileID, body.Token).Scan(&arID, &filePath, &fileName)

	if err != nil {
		http.Error(w, "Token tidak valid atau belum disetujui", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":     true,
		"file_path": filePath,
		"file_name": fileName,
		"message":   "Token valid! Anda dapat mengakses file ini.",
	})
}

// ===== INTERNAL FUNCTIONS =====

// createAccessRequest membuat permintaan akses baru
func createAccessRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB max

	documentID := r.FormValue("document_id")
	fileID := r.FormValue("file_id")
	nama := r.FormValue("nama")
	email := r.FormValue("email")

	if documentID == "" || fileID == "" || nama == "" || email == "" {
		http.Error(w, "Semua field wajib diisi (document_id, file_id, nama, email)", http.StatusBadRequest)
		return
	}

	// Validasi bahwa file memang terkunci
	var isLocked bool
	err := config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(is_locked, false) FROM document_files WHERE id = $1 AND document_id = $2`,
		fileID, documentID).Scan(&isLocked)
	if err != nil {
		http.Error(w, "File tidak ditemukan", http.StatusNotFound)
		return
	}
	if !isLocked {
		http.Error(w, "File ini tidak terkunci, tidak perlu meminta akses", http.StatusBadRequest)
		return
	}

	// Cek apakah sudah ada request pending dari email yang sama untuk file yang sama
	var existingCount int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM access_requests WHERE file_id = $1 AND email = $2 AND status = 'pending'`,
		fileID, email).Scan(&existingCount)
	if existingCount > 0 {
		http.Error(w, "Anda sudah memiliki permintaan akses yang masih menunggu persetujuan untuk file ini", http.StatusConflict)
		return
	}

	// Handle upload KTM
	var ktmPath string
	ktmFile, ktmHeader, err := r.FormFile("ktm")
	if err == nil {
		defer ktmFile.Close()

		// Pastikan folder uploads/ktm ada
		os.MkdirAll("uploads/ktm", os.ModePerm)

		ext := filepath.Ext(ktmHeader.Filename)
		storedName := uuid.New().String() + ext
		ktmPath = "uploads/ktm/" + storedName

		dst, err := os.Create(ktmPath)
		if err != nil {
			http.Error(w, "Gagal menyimpan file KTM", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, ktmFile)
	}

	// Insert ke database
	reqID := uuid.New()
	_, err = config.DB.Exec(context.Background(),
		`INSERT INTO access_requests (id, document_id, file_id, nama, email, ktm_path, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending')`,
		reqID, documentID, fileID, nama, email, ktmPath)

	if err != nil {
		http.Error(w, "Gagal menyimpan permintaan akses: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      reqID.String(),
		"message": "Permintaan akses berhasil dikirim. Silakan tunggu persetujuan admin.",
	})
}

// listAccessRequests mengambil semua permintaan akses (dengan info dokumen dan file)
func listAccessRequests(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("document_id")

	query := `SELECT ar.id, ar.document_id, ar.file_id, ar.nama, ar.email,
		COALESCE(ar.ktm_path, '') as ktm_path, ar.status,
		COALESCE(ar.access_token, '') as access_token,
		ar.created_at, ar.updated_at,
		COALESCE(d.judul, '') as doc_judul,
		COALESCE(df.file_name, '') as file_name
		FROM access_requests ar
		LEFT JOIN documents d ON d.id = ar.document_id
		LEFT JOIN document_files df ON df.id = ar.file_id`
	args := []interface{}{}

	if documentID != "" {
		query += ` WHERE ar.document_id = $1`
		args = append(args, documentID)
	}
	query += ` ORDER BY ar.created_at DESC`

	rows, err := config.DB.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AccessRequestWithInfo struct {
		models.AccessRequest
		DocJudul string `json:"doc_judul"`
		FileName string `json:"file_name"`
	}

	requests := []AccessRequestWithInfo{}
	for rows.Next() {
		var ar AccessRequestWithInfo
		err := rows.Scan(
			&ar.ID, &ar.DocumentID, &ar.FileID, &ar.Nama, &ar.Email,
			&ar.KtmPath, &ar.Status, &ar.AccessToken,
			&ar.CreatedAt, &ar.UpdatedAt,
			&ar.DocJudul, &ar.FileName,
		)
		if err != nil {
			continue
		}
		requests = append(requests, ar)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// updateAccessRequestStatus mengubah status access request
// Jika approve → generate token, kirim email
func updateAccessRequestStatus(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body tidak valid", http.StatusBadRequest)
		return
	}

	if body.Status != "approved" && body.Status != "rejected" {
		http.Error(w, "Status harus 'approved' atau 'rejected'", http.StatusBadRequest)
		return
	}

	if body.Status == "approved" {
		// Generate token dan update
		token := generateAccessToken()

		result, err := config.DB.Exec(context.Background(),
			`UPDATE access_requests SET status = 'approved', access_token = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
			token, id)
		if err != nil {
			http.Error(w, "Gagal update status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if result.RowsAffected() == 0 {
			http.Error(w, "Permintaan akses tidak ditemukan", http.StatusNotFound)
			return
		}

		// Ambil info untuk email
		var email, nama, docJudul, fileName string
		config.DB.QueryRow(context.Background(),
			`SELECT ar.email, ar.nama, COALESCE(d.judul, ''), COALESCE(df.file_name, '')
			 FROM access_requests ar
			 LEFT JOIN documents d ON d.id = ar.document_id
			 LEFT JOIN document_files df ON df.id = ar.file_id
			 WHERE ar.id = $1`, id).Scan(&email, &nama, &docJudul, &fileName)

		// Kirim email (async, jangan block response)
		go func() {
			err := sendAccessTokenEmail(email, nama, token, docJudul, fileName)
			if err != nil {
				fmt.Printf("⚠️ Gagal mengirim email ke %s: %v\n", email, err)
			} else {
				fmt.Printf("✅ Email token akses berhasil dikirim ke %s\n", email)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Permintaan akses disetujui. Token telah dikirim ke email " + email,
			"token":   token,
		})
	} else {
		// Reject
		result, err := config.DB.Exec(context.Background(),
			`UPDATE access_requests SET status = 'rejected', updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			id)
		if err != nil {
			http.Error(w, "Gagal update status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if result.RowsAffected() == 0 {
			http.Error(w, "Permintaan akses tidak ditemukan", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Permintaan akses ditolak.",
		})
	}
}

// deleteAccessRequest menghapus access request
func deleteAccessRequest(w http.ResponseWriter, r *http.Request, id string) {
	// Hapus file KTM jika ada
	var ktmPath string
	config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(ktm_path, '') FROM access_requests WHERE id = $1`, id).Scan(&ktmPath)
	if ktmPath != "" {
		os.Remove(ktmPath)
	}

	result, err := config.DB.Exec(context.Background(),
		`DELETE FROM access_requests WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Gagal menghapus permintaan akses", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Permintaan akses tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Permintaan akses berhasil dihapus",
	})
}
