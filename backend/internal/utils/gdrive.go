package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// ─── Singleton Drive Service ─────────────────────────────────────────────────
// Kita buat satu instance *drive.Service yang bisa dipakai berulang kali,
// karena OAuth2 token source akan auto-refresh access token pakai refresh token.

var (
	driveService    *drive.Service
	driveServiceMu  sync.Mutex
	driveServiceErr error
)

// ResetDriveService memungkinkan re-inisialisasi Drive service
// Panggil ini jika token expired/invalid agar bisa retry dengan token baru
func ResetDriveService() {
	driveServiceMu.Lock()
	defer driveServiceMu.Unlock()
	driveService = nil
	driveServiceErr = nil
	fmt.Println("🔄 Google Drive service di-reset, akan re-inisialisasi pada request berikutnya")
}

// getDriveService mengembalikan singleton Google Drive service.
// Mendukung 2 mode:
//  1. OAuth2 Refresh Token (prioritas) — pakai GDRIVE_CLIENT_ID, GDRIVE_CLIENT_SECRET, GDRIVE_REFRESH_TOKEN
//  2. Service Account (fallback) — pakai file JSON credentials
func getDriveService() (*drive.Service, error) {
	driveServiceMu.Lock()
	defer driveServiceMu.Unlock()

	if driveService != nil {
		return driveService, nil
	}
	if driveServiceErr != nil {
		return nil, driveServiceErr
	}

	ctx := context.Background()

	clientID := os.Getenv("GDRIVE_CLIENT_ID")
	clientSecret := os.Getenv("GDRIVE_CLIENT_SECRET")
	refreshToken := os.Getenv("GDRIVE_REFRESH_TOKEN")

	if clientID != "" && clientSecret != "" && refreshToken != "" {
		// ── Mode OAuth2 Refresh Token ──
		fmt.Println("🔑 Google Drive: menggunakan OAuth2 refresh token")

		config := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{drive.DriveFileScope},
			Endpoint:     google.Endpoint,
		}

		token := &oauth2.Token{
			RefreshToken: refreshToken,
		}

		// TokenSource akan otomatis refresh access token saat expired
		tokenSource := config.TokenSource(ctx, token)

		driveService, driveServiceErr = drive.NewService(ctx,
			option.WithTokenSource(tokenSource),
		)
		if driveServiceErr != nil {
			driveServiceErr = fmt.Errorf("gagal membuat Drive service (OAuth2): %v", driveServiceErr)
		}
	} else {
		// ── Fallback: Service Account ──
		credFile := getCredentialsFile()
		if _, err := os.Stat(credFile); os.IsNotExist(err) {
			driveServiceErr = fmt.Errorf(
				"Google Drive belum dikonfigurasi.\n"+
					"Opsi 1: Set GDRIVE_CLIENT_ID, GDRIVE_CLIENT_SECRET, GDRIVE_REFRESH_TOKEN di .env\n"+
					"Opsi 2: Sediakan file credentials service account: %s", credFile)
			return nil, driveServiceErr
		}

		fmt.Printf("🔑 Google Drive: menggunakan service account (%s)\n", credFile)
		driveService, driveServiceErr = drive.NewService(ctx, option.WithCredentialsFile(credFile))
		if driveServiceErr != nil {
			driveServiceErr = fmt.Errorf("gagal membuat Drive service (service account): %v", driveServiceErr)
		}
	}

	return driveService, driveServiceErr
}

// getCredentialsFile path ke file JSON service account (fallback)
func getCredentialsFile() string {
	if v := os.Getenv("GDRIVE_CREDENTIALS_FILE"); v != "" {
		return v
	}
	return "w-unsub-82fabe19ad9b.json"
}

// getFolderID mengembalikan folder ID Google Drive untuk upload KTM
// Set via env GDRIVE_FOLDER_ID
func getFolderID() string {
	return os.Getenv("GDRIVE_FOLDER_ID")
}

// getDocumentsFolderID mengembalikan folder ID Google Drive khusus untuk dokumen repository
// Set via env GDRIVE_DOCUMENTS_FOLDER_ID. Jika kosong, fallback ke GDRIVE_FOLDER_ID
func getDocumentsFolderID() string {
	if v := os.Getenv("GDRIVE_DOCUMENTS_FOLDER_ID"); v != "" {
		return v
	}
	return getFolderID()
}

// getMimeType mendeteksi MIME type dari ekstensi file
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

// GDriveUploadResult berisi hasil upload ke Google Drive
type GDriveUploadResult struct {
	FileID  string // ID file di Google Drive
	ViewURL string // URL untuk melihat file
}

// UploadToGDrive mengupload file ke Google Drive (ke folder default dari env)
// Parameter:
//   - fileReader: io.Reader dari file yang diupload
//   - fileName: nama file asli (untuk menentukan MIME type dan nama di Drive)
//
// Return: GDriveUploadResult, error
func UploadToGDrive(fileReader io.Reader, fileName string) (*GDriveUploadResult, error) {
	return UploadToGDriveFolder(fileReader, fileName, getFolderID())
}

// UploadToGDriveFolder mengupload file ke folder tertentu di Google Drive
// Menggunakan resumable upload dengan chunked streaming untuk efisiensi memory
func UploadToGDriveFolder(fileReader io.Reader, fileName string, folderID string) (*GDriveUploadResult, error) {
	srv, err := getDriveService()
	if err != nil {
		return nil, err
	}

	mimeType := getMimeType(fileName)

	// Metadata file
	driveFile := &drive.File{
		Name:     fileName,
		MimeType: mimeType,
	}

	// Jika ada folder ID, upload ke folder tersebut
	if folderID != "" {
		driveFile.Parents = []string{folderID}
	}

	// Upload file menggunakan resumable upload dengan chunk 8MB
	// Ini memungkinkan streaming file besar tanpa memuat seluruhnya ke memory
	fmt.Printf("📤 Mengupload '%s' ke Google Drive (streaming)...\n", fileName)
	res, err := srv.Files.Create(driveFile).
		Media(fileReader, googleapi.ChunkSize(8*1024*1024)). // 8MB chunks
		Fields("id, webViewLink, webContentLink").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, fmt.Errorf("gagal upload ke Google Drive: %v", err)
	}

	// Set permission agar bisa diakses publik (view only)
	_, err = srv.Permissions.Create(res.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()
	if err != nil {
		fmt.Printf("⚠️ Gagal set permission public pada file %s: %v\n", res.Id, err)
	}

	viewURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", res.Id)

	fmt.Printf("✅ File berhasil diupload ke Google Drive: %s (ID: %s)\n", fileName, res.Id)

	return &GDriveUploadResult{
		FileID:  res.Id,
		ViewURL: viewURL,
	}, nil
}

// FindOrCreateFolder mencari folder dengan nama tertentu di dalam parent folder.
// Jika folder sudah ada, return ID-nya. Jika belum, buat baru.
func FindOrCreateFolder(parentID, folderName string) (string, error) {
	srv, err := getDriveService()
	if err != nil {
		return "", err
	}

	// Cari folder yang sudah ada
	query := fmt.Sprintf(
		"name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents and trashed = false",
		strings.ReplaceAll(folderName, "'", "\\'"),
		parentID,
	)

	result, err := srv.Files.List().
		Q(query).
		Fields("files(id, name)").
		PageSize(1).
		Do()
	if err != nil {
		return "", fmt.Errorf("gagal mencari folder '%s': %v", folderName, err)
	}

	// Folder sudah ada
	if len(result.Files) > 0 {
		fmt.Printf("📁 Folder '%s' sudah ada (ID: %s)\n", folderName, result.Files[0].Id)
		return result.Files[0].Id, nil
	}

	// Buat folder baru
	folderMeta := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	folder, err := srv.Files.Create(folderMeta).
		Fields("id").
		Do()
	if err != nil {
		return "", fmt.Errorf("gagal membuat folder '%s': %v", folderName, err)
	}

	fmt.Printf("📁 Folder '%s' berhasil dibuat (ID: %s)\n", folderName, folder.Id)
	return folder.Id, nil
}

// GetDocumentFolderID membuat/mendapatkan hierarki folder untuk dokumen:
// Root > Fakultas > Prodi > Tahun > Penulis
// Return folder ID terdalam (folder penulis) tempat file akan disimpan.
func GetDocumentFolderID(fakultasName, prodiName string, tahun int, penulisName string) (string, error) {
	rootID := getDocumentsFolderID()
	if rootID == "" {
		return "", fmt.Errorf("GDRIVE_DOCUMENTS_FOLDER_ID atau GDRIVE_FOLDER_ID belum dikonfigurasi")
	}

	// Level 1: Fakultas
	if fakultasName == "" {
		fakultasName = "Tanpa Fakultas"
	}
	fakultasID, err := FindOrCreateFolder(rootID, fakultasName)
	if err != nil {
		return "", fmt.Errorf("gagal membuat folder fakultas: %v", err)
	}

	// Level 2: Program Studi
	if prodiName == "" {
		prodiName = "Tanpa Prodi"
	}
	prodiID, err := FindOrCreateFolder(fakultasID, prodiName)
	if err != nil {
		return "", fmt.Errorf("gagal membuat folder prodi: %v", err)
	}

	// Level 3: Tahun
	tahunStr := "Tanpa Tahun"
	if tahun > 0 {
		tahunStr = fmt.Sprintf("%d", tahun)
	}
	tahunID, err := FindOrCreateFolder(prodiID, tahunStr)
	if err != nil {
		return "", fmt.Errorf("gagal membuat folder tahun: %v", err)
	}

	// Level 4: Penulis
	if penulisName == "" {
		penulisName = "Tanpa Penulis"
	}
	penulisID, err := FindOrCreateFolder(tahunID, penulisName)
	if err != nil {
		return "", fmt.Errorf("gagal membuat folder penulis: %v", err)
	}

	return penulisID, nil
}

// DeleteFromGDrive menghapus file dari Google Drive berdasarkan file ID
func DeleteFromGDrive(fileID string) error {
	if fileID == "" {
		return nil
	}

	srv, err := getDriveService()
	if err != nil {
		return err
	}

	err = srv.Files.Delete(fileID).Do()
	if err != nil {
		return fmt.Errorf("gagal menghapus file dari Google Drive (ID: %s): %v", fileID, err)
	}

	fmt.Printf("✅ File berhasil dihapus dari Google Drive: %s\n", fileID)
	return nil
}

// IsGDriveID mengecek apakah string tertentu adalah Google Drive file ID
// (bukan path lokal seperti "uploads/ktm/xxx.jpg")
func IsGDriveID(path string) bool {
	// Path lokal biasanya mengandung "/" atau "\" dan ekstensi file
	// Google Drive ID biasanya string alfanumerik tanpa path separator
	if strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return false
	}
	if strings.Contains(path, ".") {
		return false
	}
	// Google Drive IDs biasanya 28-44 karakter
	return len(path) > 20
}

// GetGDriveViewURL menghasilkan URL untuk melihat file di Google Drive
func GetGDriveViewURL(fileID string) string {
	return fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID)
}

// GetGDriveThumbnailURL menghasilkan URL thumbnail dari Google Drive
func GetGDriveThumbnailURL(fileID string) string {
	return fmt.Sprintf("https://drive.google.com/thumbnail?id=%s&sz=w400", fileID)
}

// GetGDriveFileContent mendownload konten file dari Google Drive
// Mengembalikan io.ReadCloser dan MIME type. Caller harus menutup reader.
func GetGDriveFileContent(fileID string) (io.ReadCloser, string, error) {
	srv, err := getDriveService()
	if err != nil {
		return nil, "", err
	}

	// Ambil metadata untuk MIME type
	file, err := srv.Files.Get(fileID).Fields("mimeType").Do()
	if err != nil {
		return nil, "", fmt.Errorf("gagal mendapatkan metadata file: %v", err)
	}

	// Download file content
	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return nil, "", fmt.Errorf("gagal mendownload file: %v", err)
	}

	return resp.Body, file.MimeType, nil
}
