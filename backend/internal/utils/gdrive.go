package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GDriveCredentialsFile path ke file JSON service account
// Bisa di-override via env GDRIVE_CREDENTIALS_FILE
func getCredentialsFile() string {
	if v := os.Getenv("GDRIVE_CREDENTIALS_FILE"); v != "" {
		return v
	}
	return "w-unsub-82fabe19ad9b.json"
}

// getFolderID mengembalikan folder ID Google Drive untuk upload
// Set via env GDRIVE_FOLDER_ID
func getFolderID() string {
	return os.Getenv("GDRIVE_FOLDER_ID")
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

// UploadToGDrive mengupload file ke Google Drive
// Parameter:
//   - fileReader: io.Reader dari file yang diupload
//   - fileName: nama file asli (untuk menentukan MIME type dan nama di Drive)
//
// Return: GDriveUploadResult, error
func UploadToGDrive(fileReader io.Reader, fileName string) (*GDriveUploadResult, error) {
	ctx := context.Background()

	credFile := getCredentialsFile()
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("file credentials Google Drive tidak ditemukan: %s", credFile)
	}

	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat service Google Drive: %v", err)
	}

	mimeType := getMimeType(fileName)

	// Metadata file
	driveFile := &drive.File{
		Name:     fileName,
		MimeType: mimeType,
	}

	// Jika ada folder ID, upload ke folder tersebut
	folderID := getFolderID()
	if folderID != "" {
		driveFile.Parents = []string{folderID}
	}

	// Upload file
	res, err := srv.Files.Create(driveFile).
		Media(fileReader).
		Fields("id, webViewLink, webContentLink").
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
		// Tidak return error, file sudah terupload
	}

	viewURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", res.Id)

	fmt.Printf("✅ File berhasil diupload ke Google Drive: %s (ID: %s)\n", fileName, res.Id)

	return &GDriveUploadResult{
		FileID:  res.Id,
		ViewURL: viewURL,
	}, nil
}

// DeleteFromGDrive menghapus file dari Google Drive berdasarkan file ID
func DeleteFromGDrive(fileID string) error {
	if fileID == "" {
		return nil
	}

	ctx := context.Background()

	credFile := getCredentialsFile()
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return fmt.Errorf("file credentials Google Drive tidak ditemukan: %s", credFile)
	}

	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		return fmt.Errorf("gagal membuat service Google Drive: %v", err)
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
