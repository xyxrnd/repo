package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	"golang.org/x/crypto/bcrypt"
)

// ===== STUDENT REGISTRATION HANDLERS =====

// OCRKtmHandler mengekstrak data dari foto KTM menggunakan OCR
// POST /api/ocr-ktm
func OCRKtmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB max

	ktmFile, _, err := r.FormFile("ktm")
	if err != nil {
		http.Error(w, `{"error":"File KTM wajib diupload"}`, http.StatusBadRequest)
		return
	}
	defer ktmFile.Close()

	data, err := utils.ExtractKTMData(ktmFile)
	if err != nil {
		fmt.Printf("⚠️ OCR KTM gagal: %v\n", err)
		http.Error(w, `{"error":"Gagal membaca data KTM: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ OCR KTM berhasil - Nama: %s, NPM: %s, Fakultas: %s, Prodi: %s\n",
		data.Name, data.NPM, data.Fakultas, data.Prodi)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// StudentSignupHandler menangani pendaftaran mahasiswa baru (public)
// POST /api/student-signup
func StudentSignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB max

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	npm := r.FormValue("npm")
	fakultas := r.FormValue("fakultas")
	prodi := r.FormValue("prodi")

	if name == "" || email == "" || password == "" {
		http.Error(w, `{"error":"Nama, email, dan password wajib diisi"}`, http.StatusBadRequest)
		return
	}

	if len(password) < 6 {
		http.Error(w, `{"error":"Password minimal 6 karakter"}`, http.StatusBadRequest)
		return
	}

	// Validasi bahwa email sudah diverifikasi via OTP
	var verifiedCount int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM email_otps WHERE email = $1 AND document_id IS NULL AND is_verified = true AND expires_at > NOW() - INTERVAL '10 minutes'`,
		email).Scan(&verifiedCount)
	if verifiedCount == 0 {
		http.Error(w, `{"error":"Email belum diverifikasi. Silakan verifikasi email terlebih dahulu via OTP."}`, http.StatusForbidden)
		return
	}

	// Cek apakah email sudah terdaftar di users
	var existsInUsers bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&existsInUsers)
	if existsInUsers {
		http.Error(w, `{"error":"Email sudah terdaftar sebagai pengguna"}`, http.StatusConflict)
		return
	}

	// Cek apakah email sudah ada di student_registrations dengan status pending
	var existsInRegs bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM student_registrations WHERE email = $1 AND status = 'pending')`, email).Scan(&existsInRegs)
	if existsInRegs {
		http.Error(w, `{"error":"Pendaftaran dengan email ini sudah menunggu verifikasi"}`, http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"Gagal memproses password"}`, http.StatusInternalServerError)
		return
	}

	// Handle upload KTM ke Google Drive
	var ktmPath string
	ktmFile, ktmHeader, err := r.FormFile("ktm")
	if err != nil {
		http.Error(w, `{"error":"Upload KTM wajib dilakukan"}`, http.StatusBadRequest)
		return
	}
	defer ktmFile.Close()

	fileName := uuid.New().String() + filepath.Ext(ktmHeader.Filename)
	uploadResult, err := utils.UploadToGDrive(ktmFile, fileName)
	if err != nil {
		fmt.Printf("⚠️ Gagal upload KTM ke Google Drive: %v\n", err)
		// Reset service agar bisa retry dengan token baru
		utils.ResetDriveService()
		http.Error(w, `{"error":"Gagal mengupload file KTM. Silakan coba lagi nanti."}`, http.StatusInternalServerError)
		return
	}
	ktmPath = uploadResult.FileID // Simpan Google Drive file ID

	// Insert ke database
	regID := uuid.New().String()
	now := time.Now()

	_, err = config.DB.Exec(context.Background(),
		`INSERT INTO student_registrations (id, name, email, password, ktm_path, npm, fakultas, prodi, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending', $9, $10)`,
		regID, name, email, string(hashedPassword), ktmPath, npm, fakultas, prodi, now, now)

	if err != nil {
		http.Error(w, `{"error":"Gagal menyimpan pendaftaran: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      regID,
		"message": "Pendaftaran berhasil! Silakan tunggu verifikasi dari admin. Informasi akun akan dikirim ke email Anda.",
	})
}

// StudentRegistrationsHandler menangani list pendaftaran mahasiswa (admin only)
// GET /api/student-registrations
func StudentRegistrationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	rows, err := config.DB.Query(context.Background(),
		`SELECT id, name, email, COALESCE(ktm_path, '') as ktm_path, COALESCE(npm, '') as npm, COALESCE(fakultas, '') as fakultas, COALESCE(prodi, '') as prodi, status, created_at, updated_at
		 FROM student_registrations
		 ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, `{"error":"Gagal mengambil data: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	registrations := []models.StudentRegistration{}
	for rows.Next() {
		var reg models.StudentRegistration
		err := rows.Scan(&reg.ID, &reg.Name, &reg.Email, &reg.KtmPath, &reg.NPM, &reg.Fakultas, &reg.Prodi, &reg.Status, &reg.CreatedAt, &reg.UpdatedAt)
		if err != nil {
			continue
		}
		registrations = append(registrations, reg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registrations)
}

// StudentRegistrationByIdHandler menangani operasi pada pendaftaran tertentu
// PUT /api/student-registrations/:id - Approve/Reject
// DELETE /api/student-registrations/:id - Delete
func StudentRegistrationByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	id := strings.TrimPrefix(r.URL.Path, "/api/student-registrations/")
	if id == "" {
		http.Error(w, `{"error":"ID tidak valid"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		updateStudentRegistrationStatus(w, r, id)
	case http.MethodDelete:
		deleteStudentRegistration(w, r, id)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// updateStudentRegistrationStatus mengubah status pendaftaran
// Jika approved → buat akun mahasiswa, kirim email
func updateStudentRegistrationStatus(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Body tidak valid"}`, http.StatusBadRequest)
		return
	}

	if body.Status != "approved" && body.Status != "rejected" {
		http.Error(w, `{"error":"Status harus 'approved' atau 'rejected'"}`, http.StatusBadRequest)
		return
	}

	if body.Status == "approved" {
		// Ambil data pendaftaran
		var name, email, hashedPassword string
		err := config.DB.QueryRow(context.Background(),
			`SELECT name, email, password FROM student_registrations WHERE id = $1 AND status = 'pending'`,
			id).Scan(&name, &email, &hashedPassword)

		if err != nil {
			http.Error(w, `{"error":"Pendaftaran tidak ditemukan atau sudah diproses"}`, http.StatusNotFound)
			return
		}

		// Cek lagi apakah email sudah ada di users
		var existsInUsers bool
		config.DB.QueryRow(context.Background(),
			`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&existsInUsers)
		if existsInUsers {
			http.Error(w, `{"error":"Email sudah terdaftar sebagai pengguna"}`, http.StatusConflict)
			return
		}

		// Buat akun user dengan role "mahasiswa"
		userID := uuid.New().String()
		now := time.Now()

		_, err = config.DB.Exec(context.Background(),
			`INSERT INTO users (id, name, email, password, role, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 'mahasiswa', $5, $6)`,
			userID, name, email, hashedPassword, now, now)

		if err != nil {
			http.Error(w, `{"error":"Gagal membuat akun: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		// Update status registrasi
		config.DB.Exec(context.Background(),
			`UPDATE student_registrations SET status = 'approved', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)

		// Kirim email notifikasi (async)
		go func() {
			err := sendStudentApprovalEmail(email, name)
			if err != nil {
				fmt.Printf("⚠️ Gagal mengirim email ke %s: %v\n", email, err)
			} else {
				fmt.Printf("✅ Email persetujuan pendaftaran berhasil dikirim ke %s\n", email)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Pendaftaran disetujui. Akun mahasiswa telah dibuat dan email notifikasi dikirim ke " + email,
		})
	} else {
		// Reject
		result, err := config.DB.Exec(context.Background(),
			`UPDATE student_registrations SET status = 'rejected', updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			id)
		if err != nil {
			http.Error(w, `{"error":"Gagal update status: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		if result.RowsAffected() == 0 {
			http.Error(w, `{"error":"Pendaftaran tidak ditemukan"}`, http.StatusNotFound)
			return
		}

		// Ambil data untuk notifikasi
		var email, name string
		config.DB.QueryRow(context.Background(),
			`SELECT email, name FROM student_registrations WHERE id = $1`, id).Scan(&email, &name)

		// Kirim email penolakan (async)
		go func() {
			err := sendStudentRejectionEmail(email, name)
			if err != nil {
				fmt.Printf("⚠️ Gagal mengirim email penolakan ke %s: %v\n", email, err)
			} else {
				fmt.Printf("✅ Email penolakan pendaftaran dikirim ke %s\n", email)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Pendaftaran ditolak.",
		})
	}
}

// deleteStudentRegistration menghapus pendaftaran
func deleteStudentRegistration(w http.ResponseWriter, r *http.Request, id string) {
	// Hapus file KTM dari Google Drive atau lokal jika ada
	var ktmPath string
	config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(ktm_path, '') FROM student_registrations WHERE id = $1`, id).Scan(&ktmPath)
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
		`DELETE FROM student_registrations WHERE id = $1`, id)
	if err != nil {
		http.Error(w, `{"error":"Gagal menghapus pendaftaran"}`, http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, `{"error":"Pendaftaran tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Pendaftaran berhasil dihapus",
	})
}

// ===== EMAIL FUNCTIONS =====

// sendStudentApprovalEmail mengirim email pemberitahuan bahwa pendaftaran disetujui
func sendStudentApprovalEmail(toEmail, toName string) error {
	host := getSmtpHost()
	port := getSmtpPort()
	senderEmail := getSmtpEmail()
	senderPassword := getSmtpPassword()

	if senderEmail == "" || senderPassword == "" {
		return fmt.Errorf("SMTP belum dikonfigurasi")
	}

	fmt.Printf("📧 Mengirim email persetujuan ke %s via %s:%s\n", toEmail, host, port)

	subject := "Pendaftaran Mahasiswa Disetujui - Repository Akademik"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc;">
  <div style="background: linear-gradient(135deg, #10b981, #059669); padding: 30px; border-radius: 16px 16px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0; font-size: 24px;">🎓 Selamat! Pendaftaran Disetujui</h1>
  </div>
  <div style="background: white; padding: 30px; border-radius: 0 0 16px 16px; box-shadow: 0 4px 6px rgba(0,0,0,0.05);">
    <p style="color: #334155; font-size: 16px;">Halo <strong>%s</strong>,</p>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Pendaftaran akun mahasiswa Anda telah <span style="color: #10b981; font-weight: bold;">disetujui</span> oleh administrator.
    </p>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Anda sekarang dapat login ke sistem Repository Akademik menggunakan akun berikut:
    </p>
    <div style="background: linear-gradient(135deg, #f0fdf4, #dcfce7); border: 2px solid #86efac; border-radius: 12px; padding: 20px; margin: 20px 0;">
      <table style="width: 100%%;">
        <tr>
          <td style="color: #64748b; font-size: 13px; padding: 6px 0; width: 80px;">Email</td>
          <td style="color: #1e293b; font-size: 14px; font-weight: bold; padding: 6px 0;">%s</td>
        </tr>
        <tr>
          <td style="color: #64748b; font-size: 13px; padding: 6px 0;">Password</td>
          <td style="color: #1e293b; font-size: 14px; font-weight: bold; padding: 6px 0;">(Password yang Anda gunakan saat mendaftar)</td>
        </tr>
      </table>
    </div>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Dengan akun ini, Anda dapat mengakses semua dokumen yang tersedia di Repository Akademik, termasuk dokumen yang terkunci.
    </p>
    <hr style="border: none; border-top: 1px solid #e2e8f0; margin: 20px 0;">
    <p style="color: #94a3b8; font-size: 12px; text-align: center;">
      Email ini dikirim secara otomatis oleh sistem Repository Akademik.<br>
      Jangan bagikan informasi akun Anda kepada orang lain.
    </p>
  </div>
</body>
</html>`, toName, toEmail)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		senderEmail, toEmail, subject)

	msg := []byte(headers + body)

	auth := smtp.PlainAuth("", senderEmail, senderPassword, host)
	return smtp.SendMail(host+":"+port, auth, senderEmail, []string{toEmail}, msg)
}

// sendStudentRejectionEmail mengirim email pemberitahuan bahwa pendaftaran ditolak
func sendStudentRejectionEmail(toEmail, toName string) error {
	host := getSmtpHost()
	port := getSmtpPort()
	senderEmail := getSmtpEmail()
	senderPassword := getSmtpPassword()

	if senderEmail == "" || senderPassword == "" {
		return fmt.Errorf("SMTP belum dikonfigurasi")
	}

	subject := "Pendaftaran Mahasiswa Ditolak - Repository Akademik"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc;">
  <div style="background: linear-gradient(135deg, #ef4444, #dc2626); padding: 30px; border-radius: 16px 16px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0; font-size: 24px;">Pendaftaran Tidak Disetujui</h1>
  </div>
  <div style="background: white; padding: 30px; border-radius: 0 0 16px 16px; box-shadow: 0 4px 6px rgba(0,0,0,0.05);">
    <p style="color: #334155; font-size: 16px;">Halo <strong>%s</strong>,</p>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Mohon maaf, pendaftaran akun mahasiswa Anda <span style="color: #ef4444; font-weight: bold;">tidak disetujui</span> oleh administrator.
    </p>
    <p style="color: #475569; font-size: 14px; line-height: 1.6;">
      Hal ini mungkin disebabkan oleh dokumen KTM yang tidak valid atau data yang tidak lengkap. Silakan coba mendaftar kembali dengan data yang benar atau hubungi administrator untuk informasi lebih lanjut.
    </p>
    <hr style="border: none; border-top: 1px solid #e2e8f0; margin: 20px 0;">
    <p style="color: #94a3b8; font-size: 12px; text-align: center;">
      Email ini dikirim secara otomatis oleh sistem Repository Akademik.
    </p>
  </div>
</body>
</html>`, toName)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		senderEmail, toEmail, subject)

	msg := []byte(headers + body)

	auth := smtp.PlainAuth("", senderEmail, senderPassword, host)
	return smtp.SendMail(host+":"+port, auth, senderEmail, []string{toEmail}, msg)
}
