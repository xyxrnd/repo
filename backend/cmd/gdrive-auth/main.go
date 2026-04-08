package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Tool untuk mendapatkan Google Drive refresh token via OAuth2.
// Menggunakan local HTTP server untuk menangkap authorization code secara otomatis.
//
// Jalankan sekali: go run ./cmd/gdrive-auth/
// Browser akan terbuka otomatis → login → Allow → token muncul di terminal.

const redirectPort = "8090"

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
		RedirectURL:  fmt.Sprintf("http://localhost:%s/callback", redirectPort),
	}

	// Channel untuk menerima auth code dari HTTP handler
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Setup local HTTP server untuk menangkap callback
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "tidak ada authorization code"
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<html><body style="font-family:sans-serif;text-align:center;padding:50px">
				<h1 style="color:red">❌ Gagal</h1>
				<p>Error: %s</p>
				<p>Silakan coba lagi.</p>
			</body></html>`, errMsg)
			errChan <- fmt.Errorf("OAuth error: %s", errMsg)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body style="font-family:sans-serif;text-align:center;padding:50px">
			<h1 style="color:green">✅ Berhasil!</h1>
			<p>Authorization berhasil. Anda bisa menutup tab ini dan kembali ke terminal.</p>
		</body></html>`)
		codeChan <- code
	})

	server := &http.Server{
		Addr:    ":" + redirectPort,
		Handler: mux,
	}

	// Start server di background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Gagal start local server: %v", err)
		}
	}()

	// Generate authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║       Google Drive OAuth2 - Setup Refresh Token     ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Membuka browser untuk login...")
	fmt.Println()

	// Buka browser otomatis
	if err := openBrowser(authURL); err != nil {
		fmt.Println("⚠️  Tidak bisa membuka browser otomatis.")
		fmt.Println("   Buka link berikut secara manual:")
		fmt.Println()
		fmt.Println("   " + authURL)
	} else {
		fmt.Println("✅ Browser terbuka. Silakan login dan klik 'Allow'.")
	}

	fmt.Println()
	fmt.Println("Menunggu authorization...")

	// Tunggu auth code atau error
	var code string
	select {
	case code = <-codeChan:
		// success
	case err := <-errChan:
		server.Shutdown(context.Background())
		log.Fatalf("❌ %v", err)
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		log.Fatal("❌ Timeout: tidak ada response dalam 5 menit")
	}

	// Shutdown server
	server.Shutdown(context.Background())

	// Exchange code for token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("❌ Gagal mendapatkan token: %v", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ TOKEN BERHASIL DIDAPAT!             ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Tambahkan baris berikut ke file .env Anda:")
	fmt.Println()
	fmt.Printf("   GDRIVE_REFRESH_TOKEN=%s\n", token.RefreshToken)
	fmt.Println()
	fmt.Println("Setelah itu, restart backend server.")
}

// openBrowser membuka URL di browser default
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
