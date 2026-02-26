package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Tool sederhana untuk mendapatkan Google Drive refresh token.
// Jalankan sekali: go run ./cmd/gdrive-auth/
// Ikuti instruksi di terminal, lalu simpan refresh token ke .env

func main() {
	godotenv.Load()

	clientID := os.Getenv("GDRIVE_CLIENT_ID")
	clientSecret := os.Getenv("GDRIVE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("GDRIVE_CLIENT_ID dan GDRIVE_CLIENT_SECRET harus diisi di .env\n" +
			"Buat OAuth2 credentials di https://console.cloud.google.com/apis/credentials")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{drive.DriveFileScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}

	// Generate authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       Google Drive OAuth2 - Setup Refresh Token     ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("1. Buka link berikut di browser:")
	fmt.Println()
	fmt.Println("   " + authURL)
	fmt.Println()
	fmt.Println("2. Login dengan akun Google Anda dan klik 'Allow'")
	fmt.Println("3. Copy authorization code yang muncul")
	fmt.Println()
	fmt.Print("Paste authorization code di sini: ")

	var code string
	fmt.Scan(&code)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Gagal mendapatkan token: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Berhasil mendapatkan token!")
	fmt.Println()
	fmt.Println("Tambahkan baris berikut ke file .env Anda:")
	fmt.Println()
	fmt.Printf("GDRIVE_REFRESH_TOKEN=%s\n", token.RefreshToken)
	fmt.Println()
	fmt.Println("Setelah itu, restart backend server.")
}
