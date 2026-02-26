package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"repository-un/internal/config"
	"repository-un/internal/middleware"
	"repository-un/internal/models"
	"repository-un/internal/utils"

	"github.com/google/uuid"
)

// ===== SMTP Configuration =====
// Nilai dibaca saat runtime (bukan saat init) agar godotenv.Load() sudah dijalankan terlebih dahulu

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getSmtpHost() string     { return getEnv("SMTP_HOST", "smtp.gmail.com") }
func getSmtpPort() string     { return getEnv("SMTP_PORT", "587") }
func getSmtpEmail() string    { return getEnv("SMTP_EMAIL", "") }
func getSmtpPassword() string { return getEnv("SMTP_PASSWORD", "") }

// getFrontendURL mengembalikan URL frontend untuk link di email
func getFrontendURL() string {
	return getEnv("FRONTEND_URL", "http://localhost:8080")
}

// generateAccessToken membuat token akses acak (16 karakter hex)
func generateAccessToken() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateOTP membuat kode OTP 6 digit
func generateOTP() string {
	mathrand.Seed(time.Now().UnixNano())
	otp := mathrand.Intn(900000) + 100000 // 100000-999999
	return fmt.Sprintf("%d", otp)
}

// sendOTPEmail mengirim email berisi kode OTP untuk verifikasi email
func sendOTPEmail(toEmail, otpCode string) error {
	host := getSmtpHost()
	port := getSmtpPort()
	email := getSmtpEmail()
	password := getSmtpPassword()

	if email == "" || password == "" {
		return fmt.Errorf("SMTP belum dikonfigurasi")
	}

	fmt.Printf("📧 Mengirim OTP ke %s\n", toEmail)

	subject := "Kode Verifikasi Email - Repository Akademik"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc;">
  <div style="background: linear-gradient(135deg, #3b82f6, #6366f1); padding: 30px; border-radius: 16px 16px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0; font-size: 24px;">🔐 Verifikasi Email</h1>
  </div>
  <div style="background: white; padding: 30px; border-radius: 0 0 16px 16px; box-shadow: 0 4px 6px rgba(0,0,0,0.05);">
    <p style="color: #334155; font-size: 16px;">Halo,</p>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Anda meminta akses ke dokumen di Repository Akademik. Gunakan kode OTP berikut untuk memverifikasi email Anda:
    </p>
    <div style="background: linear-gradient(135deg, #f0f9ff, #e0f2fe); border: 2px dashed #3b82f6; border-radius: 12px; padding: 20px; text-align: center; margin: 25px 0;">
      <p style="color: #64748b; font-size: 11px; margin: 0 0 8px 0; text-transform: uppercase; letter-spacing: 1px;">Kode OTP Anda</p>
      <p style="color: #1e40af; font-size: 32px; font-weight: bold; margin: 0; letter-spacing: 8px; font-family: 'Courier New', monospace;">%s</p>
    </div>
    <p style="color: #475569; font-size: 13px; line-height: 1.6;">
      Kode ini berlaku selama <strong>5 menit</strong>. Jangan bagikan kode ini kepada siapa pun.
    </p>
    <hr style="border: none; border-top: 1px solid #e2e8f0; margin: 20px 0;">
    <p style="color: #94a3b8; font-size: 12px; text-align: center;">
      Email ini dikirim secara otomatis oleh sistem Repository Akademik.<br>
      Jika Anda tidak meminta kode ini, abaikan email ini.
    </p>
  </div>
</body>
</html>`, otpCode)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		email, toEmail, subject)

	msg := []byte(headers + body)
	auth := smtp.PlainAuth("", email, password, host)

	err := smtp.SendMail(host+":"+port, auth, email, []string{toEmail}, msg)
	if err != nil {
		fmt.Printf("❌ Gagal kirim OTP: %v\n", err)
	}
	return err
}

// sendAccessTokenEmail mengirim email yang berisi link ke halaman dokumen dengan token
func sendAccessTokenEmail(toEmail, toName, token, documentID, documentTitle string) error {
	host := getSmtpHost()
	port := getSmtpPort()
	email := getSmtpEmail()
	password := getSmtpPassword()
	frontendURL := getFrontendURL()

	// Validasi konfigurasi
	if email == "" || password == "" {
		return fmt.Errorf("SMTP belum dikonfigurasi: SMTP_EMAIL=%q, SMTP_PASSWORD kosong=%v", email, password == "")
	}

	fmt.Printf("📧 Mengirim email ke %s via %s:%s dari %s\n", toEmail, host, port, email)

	from := email
	to := toEmail

	// Link ke halaman dokumen dengan token sebagai query parameter
	documentLink := fmt.Sprintf("%s/#/document/%s?token=%s", frontendURL, documentID, token)

	subject := "Akses Dokumen Disetujui - " + documentTitle
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
      Permintaan akses Anda untuk dokumen <strong>%s</strong> telah <span style="color: #22c55e; font-weight: bold;">disetujui</span>.
    </p>
    <p style="color: #475569; font-size: 14px;">Klik tombol di bawah untuk membuka dokumen:</p>
    <div style="text-align: center; margin: 25px 0;">
      <a href="%s" style="display: inline-block; padding: 14px 32px; background: linear-gradient(135deg, #3b82f6, #6366f1); color: white; text-decoration: none; font-size: 16px; font-weight: bold; border-radius: 12px; box-shadow: 0 4px 12px rgba(59,130,246,0.4);">
        📄 Buka Dokumen
      </a>
    </div>
    <p style="color: #64748b; font-size: 12px; text-align: center; margin-top: 8px;">
      Atau salin link berikut ke browser Anda:<br>
      <span style="color: #3b82f6; word-break: break-all;">%s</span>
    </p>
    <div style="background: linear-gradient(135deg, #f0f9ff, #e0f2fe); border: 2px dashed #3b82f6; border-radius: 12px; padding: 16px; text-align: center; margin: 20px 0;">
      <p style="color: #64748b; font-size: 11px; margin: 0 0 6px 0; text-transform: uppercase; letter-spacing: 1px;">Token Akses Anda (backup)</p>
      <p style="color: #1e40af; font-size: 22px; font-weight: bold; margin: 0; letter-spacing: 3px; font-family: 'Courier New', monospace;">%s</p>
    </div>
    <p style="color: #475569; font-size: 13px; line-height: 1.6;">
      Setelah membuka link di atas, semua file terkunci pada dokumen ini akan otomatis terbuka untuk Anda preview dan download.
    </p>
    <hr style="border: none; border-top: 1px solid #e2e8f0; margin: 20px 0;">
    <p style="color: #94a3b8; font-size: 12px; text-align: center;">
      Email ini dikirim secara otomatis oleh sistem Repository Akademik.<br>
      Jangan bagikan link atau token ini kepada orang lain.
    </p>
  </div>
</body>
</html>`, toName, documentTitle, documentLink, documentLink, token)

	// Compose email
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		from, to, subject)

	msg := []byte(headers + body)

	auth := smtp.PlainAuth("", email, password, host)

	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)
	if err != nil {
		fmt.Printf("❌ Gagal kirim email: %v\n", err)
	}
	return err
}

// ===== HANDLERS =====

// SendOTPHandler mengirim kode OTP ke email user untuk verifikasi
// POST /api/send-otp
// Body: { "email": "xxx", "document_id": "xxx" }
func SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email      string `json:"email"`
		DocumentID string `json:"document_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body tidak valid", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.DocumentID == "" {
		http.Error(w, "Email dan document_id wajib diisi", http.StatusBadRequest)
		return
	}

	// Rate limiting: cek apakah sudah kirim OTP dalam 1 menit terakhir
	var recentCount int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM email_otps WHERE email = $1 AND document_id = $2 AND created_at > NOW() - INTERVAL '1 minute'`,
		body.Email, body.DocumentID).Scan(&recentCount)
	if recentCount > 0 {
		http.Error(w, "Kode OTP sudah dikirim. Tunggu 1 menit sebelum mengirim ulang.", http.StatusTooManyRequests)
		return
	}

	// Generate OTP
	otpCode := generateOTP()
	otpID := uuid.New()
	expiresAt := time.Now().Add(5 * time.Minute)

	// Hapus OTP lama untuk email + dokumen ini
	config.DB.Exec(context.Background(),
		`DELETE FROM email_otps WHERE email = $1 AND document_id = $2`,
		body.Email, body.DocumentID)

	// Simpan OTP baru
	_, err := config.DB.Exec(context.Background(),
		`INSERT INTO email_otps (id, email, otp_code, document_id, expires_at) VALUES ($1, $2, $3, $4, $5)`,
		otpID, body.Email, otpCode, body.DocumentID, expiresAt)
	if err != nil {
		http.Error(w, "Gagal menyimpan OTP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim email OTP (async)
	go func() {
		err := sendOTPEmail(body.Email, otpCode)
		if err != nil {
			fmt.Printf("⚠️ Gagal mengirim OTP ke %s: %v\n", body.Email, err)
		} else {
			fmt.Printf("✅ OTP berhasil dikirim ke %s\n", body.Email)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Kode OTP telah dikirim ke email " + body.Email,
	})
}

// VerifyOTPHandler memverifikasi kode OTP yang dimasukkan user
// POST /api/verify-otp
// Body: { "email": "xxx", "document_id": "xxx", "otp_code": "xxx" }
func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email      string `json:"email"`
		DocumentID string `json:"document_id"`
		OTPCode    string `json:"otp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body tidak valid", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.DocumentID == "" || body.OTPCode == "" {
		http.Error(w, "Email, document_id, dan otp_code wajib diisi", http.StatusBadRequest)
		return
	}

	// Cari OTP yang cocok dan belum expired
	var otpID string
	err := config.DB.QueryRow(context.Background(),
		`SELECT id FROM email_otps
		 WHERE email = $1 AND document_id = $2 AND otp_code = $3
		   AND is_verified = false AND expires_at > NOW()
		 LIMIT 1`,
		body.Email, body.DocumentID, body.OTPCode).Scan(&otpID)

	if err != nil {
		http.Error(w, "Kode OTP tidak valid atau sudah kedaluwarsa.", http.StatusUnauthorized)
		return
	}

	// Tandai OTP sebagai terverifikasi
	config.DB.Exec(context.Background(),
		`UPDATE email_otps SET is_verified = true WHERE id = $1`, otpID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"verified": true,
		"message":  "Email berhasil diverifikasi!",
	})
}

// AccessRequestHandler menangani permintaan akses dokumen terkunci
// POST /api/access-requests - Submit permintaan akses (public, per-dokumen)
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

// VerifyAccessTokenHandler memverifikasi token akses per-dokumen
// POST /api/verify-access-token
// Body: { "document_id": "xxx", "token": "xxx" }
// Response: { valid: true, files: [{ file_id, file_path, file_name }, ...] }
func VerifyAccessTokenHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		DocumentID string `json:"document_id"`
		Token      string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body tidak valid", http.StatusBadRequest)
		return
	}

	if body.DocumentID == "" || body.Token == "" {
		http.Error(w, "document_id dan token wajib diisi", http.StatusBadRequest)
		return
	}

	// Cari access request yang cocok: token benar, status approved, untuk dokumen tersebut
	var arID string
	err := config.DB.QueryRow(context.Background(),
		`SELECT ar.id FROM access_requests ar
		 WHERE ar.document_id = $1 AND ar.access_token = $2 AND ar.status = 'approved'
		 LIMIT 1`,
		body.DocumentID, body.Token).Scan(&arID)

	if err != nil {
		http.Error(w, "Token tidak valid atau belum disetujui", http.StatusUnauthorized)
		return
	}

	// Ambil semua locked files dari dokumen ini
	rows, err := config.DB.Query(context.Background(),
		`SELECT id, file_path, file_name FROM document_files
		 WHERE document_id = $1 AND COALESCE(is_locked, false) = true`,
		body.DocumentID)
	if err != nil {
		http.Error(w, "Gagal mengambil data file", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FileInfo struct {
		FileID   string `json:"file_id"`
		FilePath string `json:"file_path"`
		FileName string `json:"file_name"`
	}

	files := []FileInfo{}
	for rows.Next() {
		var f FileInfo
		if err := rows.Scan(&f.FileID, &f.FilePath, &f.FileName); err == nil {
			files = append(files, f)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"files":   files,
		"message": "Token valid! Semua file terkunci pada dokumen ini dapat diakses.",
	})
}

// ===== INTERNAL FUNCTIONS =====

// createAccessRequest membuat permintaan akses baru (per-dokumen, bukan per-file)
// Sekarang wajib email sudah diverifikasi via OTP sebelum submit
func createAccessRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB max

	documentID := r.FormValue("document_id")
	nama := r.FormValue("nama")
	email := r.FormValue("email")

	if documentID == "" || nama == "" || email == "" {
		http.Error(w, "Semua field wajib diisi (document_id, nama, email)", http.StatusBadRequest)
		return
	}

	// Validasi bahwa email sudah diverifikasi via OTP
	var verifiedCount int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM email_otps WHERE email = $1 AND document_id = $2 AND is_verified = true AND expires_at > NOW() - INTERVAL '10 minutes'`,
		email, documentID).Scan(&verifiedCount)
	if verifiedCount == 0 {
		http.Error(w, "Email belum diverifikasi. Silakan verifikasi email terlebih dahulu.", http.StatusForbidden)
		return
	}

	// Validasi bahwa dokumen memiliki file yang terkunci
	var lockedCount int
	err := config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM document_files WHERE document_id = $1 AND COALESCE(is_locked, false) = true`,
		documentID).Scan(&lockedCount)
	if err != nil || lockedCount == 0 {
		http.Error(w, "Dokumen ini tidak memiliki file terkunci, tidak perlu meminta akses", http.StatusBadRequest)
		return
	}

	// Cek apakah sudah ada request pending dari email yang sama untuk dokumen yang sama
	var existingCount int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM access_requests WHERE document_id = $1 AND email = $2 AND status = 'pending'`,
		documentID, email).Scan(&existingCount)
	if existingCount > 0 {
		http.Error(w, "Anda sudah memiliki permintaan akses yang masih menunggu persetujuan untuk dokumen ini", http.StatusConflict)
		return
	}

	// Handle upload KTM ke Google Drive
	var ktmPath string
	ktmFile, ktmHeader, err := r.FormFile("ktm")
	if err == nil {
		defer ktmFile.Close()

		fileName := uuid.New().String() + filepath.Ext(ktmHeader.Filename)
		result, err := utils.UploadToGDrive(ktmFile, fileName)
		if err != nil {
			fmt.Printf("⚠️ Gagal upload KTM ke Google Drive: %v\n", err)
			http.Error(w, "Gagal mengupload file KTM: "+err.Error(), http.StatusInternalServerError)
			return
		}
		ktmPath = result.FileID // Simpan Google Drive file ID
	}

	// Insert ke database (tanpa file_id, hanya document_id)
	reqID := uuid.New()
	_, err = config.DB.Exec(context.Background(),
		`INSERT INTO access_requests (id, document_id, nama, email, ktm_path, status)
		 VALUES ($1, $2, $3, $4, $5, 'pending')`,
		reqID, documentID, nama, email, ktmPath)

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

// listAccessRequests mengambil semua permintaan akses (dengan info dokumen)
func listAccessRequests(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("document_id")

	query := `SELECT ar.id, ar.document_id, ar.nama, ar.email,
		COALESCE(ar.ktm_path, '') as ktm_path, ar.status,
		COALESCE(ar.access_token, '') as access_token,
		ar.created_at, ar.updated_at,
		COALESCE(d.judul, '') as doc_judul
		FROM access_requests ar
		LEFT JOIN documents d ON d.id = ar.document_id`
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
	}

	requests := []AccessRequestWithInfo{}
	for rows.Next() {
		var ar AccessRequestWithInfo
		err := rows.Scan(
			&ar.ID, &ar.DocumentID, &ar.Nama, &ar.Email,
			&ar.KtmPath, &ar.Status, &ar.AccessToken,
			&ar.CreatedAt, &ar.UpdatedAt,
			&ar.DocJudul,
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
// Jika approve → generate token, kirim email dengan link ke halaman dokumen
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
		var email, nama, docID, docJudul string
		config.DB.QueryRow(context.Background(),
			`SELECT ar.email, ar.nama, ar.document_id, COALESCE(d.judul, '')
			 FROM access_requests ar
			 LEFT JOIN documents d ON d.id = ar.document_id
			 WHERE ar.id = $1`, id).Scan(&email, &nama, &docID, &docJudul)

		// Kirim email (async, jangan block response)
		go func() {
			err := sendAccessTokenEmail(email, nama, token, docID, docJudul)
			if err != nil {
				fmt.Printf("⚠️ Gagal mengirim email ke %s: %v\n", email, err)
			} else {
				fmt.Printf("✅ Email token akses berhasil dikirim ke %s\n", email)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Permintaan akses disetujui. Link akses telah dikirim ke email " + email,
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
	// Hapus file KTM dari Google Drive atau lokal jika ada
	var ktmPath string
	config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(ktm_path, '') FROM access_requests WHERE id = $1`, id).Scan(&ktmPath)
	if ktmPath != "" {
		if utils.IsGDriveID(ktmPath) {
			// Hapus dari Google Drive
			if err := utils.DeleteFromGDrive(ktmPath); err != nil {
				fmt.Printf("⚠️ Gagal hapus KTM dari Google Drive: %v\n", err)
			}
		} else {
			// Hapus dari lokal (file lama)
			os.Remove(ktmPath)
		}
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
