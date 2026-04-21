package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// KTMData berisi data yang diekstrak dari foto KTM
type KTMData struct {
	Name     string `json:"name"`
	NPM      string `json:"npm"`
	Fakultas string `json:"fakultas"`
	Prodi    string `json:"prodi"`
	RawText  string `json:"raw_text"`
}

// ── Google Vision API types ──

type visionRequest struct {
	Requests []visionAnnotateRequest `json:"requests"`
}

type visionAnnotateRequest struct {
	Image    visionImage    `json:"image"`
	Features []visionFeature `json:"features"`
}

type visionImage struct {
	Content string `json:"content"` // base64 encoded
}

type visionFeature struct {
	Type string `json:"type"`
}

type visionResponse struct {
	Responses []visionAnnotateResponse `json:"responses"`
}

type visionAnnotateResponse struct {
	TextAnnotations []visionTextAnnotation `json:"textAnnotations"`
	Error           *visionError           `json:"error"`
}

type visionTextAnnotation struct {
	Description string `json:"description"`
}

type visionError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ExtractKTMData mengekstrak data dari foto KTM menggunakan Google Cloud Vision API
// Memerlukan env GOOGLE_VISION_API_KEY
func ExtractKTMData(imageReader io.Reader) (*KTMData, error) {
	apiKey := os.Getenv("GOOGLE_VISION_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_VISION_API_KEY belum dikonfigurasi. Set di file .env")
	}

	// Baca seluruh file image ke memory
	imageBytes, err := io.ReadAll(imageReader)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file gambar: %v", err)
	}

	// Base64 encode
	imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)

	// Buat request ke Vision API
	reqBody := visionRequest{
		Requests: []visionAnnotateRequest{
			{
				Image: visionImage{Content: imageBase64},
				Features: []visionFeature{
					{Type: "TEXT_DETECTION"},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request JSON: %v", err)
	}

	// Kirim ke Google Vision API
	url := fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Google Vision API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google Vision API error (status %d): %s", resp.StatusCode, string(body))
	}

	var visionResp visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&visionResp); err != nil {
		return nil, fmt.Errorf("gagal parsing response Vision API: %v", err)
	}

	if len(visionResp.Responses) == 0 {
		return nil, fmt.Errorf("tidak ada response dari Vision API")
	}

	apiResp := visionResp.Responses[0]
	if apiResp.Error != nil {
		return nil, fmt.Errorf("Vision API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.TextAnnotations) == 0 {
		return &KTMData{RawText: ""}, nil // Tidak ada teks terdeteksi
	}

	// TextAnnotations[0] berisi SELURUH teks yang terdeteksi
	fullText := apiResp.TextAnnotations[0].Description
	fmt.Printf("📄 OCR Raw Text:\n%s\n", fullText)

	// Parse data dari teks
	data := parseKTMText(fullText)
	data.RawText = fullText

	return data, nil
}

// parseKTMText mengekstrak nama, NPM, fakultas, dan prodi dari teks OCR KTM
func parseKTMText(text string) *KTMData {
	data := &KTMData{}
	lines := strings.Split(text, "\n")

	// Normalisasi setiap line
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	// ── Extract NPM/NIM/NRP (nomor mahasiswa) ──
	// Cari pola angka panjang (8-15 digit) yang biasanya adalah NPM
	npmRegex := regexp.MustCompile(`\b(\d{8,15})\b`)
	for _, line := range cleanLines {
		upper := strings.ToUpper(line)
		// Cari line yang mengandung keyword NPM/NIM/NRP
		if strings.Contains(upper, "NPM") || strings.Contains(upper, "NIM") || strings.Contains(upper, "NRP") || strings.Contains(upper, "NO. MAHASISWA") {
			matches := npmRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				data.NPM = matches[1]
			} else {
				// Cek line berikutnya
				idx := indexOf(cleanLines, line)
				if idx >= 0 && idx+1 < len(cleanLines) {
					matches = npmRegex.FindStringSubmatch(cleanLines[idx+1])
					if len(matches) > 1 {
						data.NPM = matches[1]
					}
				}
			}
			break
		}
	}
	// Fallback: jika belum ditemukan, cari angka panjang pertama
	if data.NPM == "" {
		for _, line := range cleanLines {
			matches := npmRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				data.NPM = matches[1]
				break
			}
		}
	}

	// ── Extract Nama ──
	for i, line := range cleanLines {
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "NAMA") && !strings.Contains(upper, "NAMA UNIVERSITAS") {
			// Format: "Nama : John Doe" atau "NAMA" diikuti value di line berikutnya
			parts := splitKeyValue(line)
			if parts != "" {
				data.Name = toTitleCase(parts)
			} else if i+1 < len(cleanLines) {
				// Value di line berikutnya
				nextLine := cleanLines[i+1]
				if !isKeyword(nextLine) {
					data.Name = toTitleCase(nextLine)
				}
			}
			break
		}
	}

	// ── Extract Fakultas ──
	for i, line := range cleanLines {
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "FAKULTAS") {
			// Bisa jadi "Fakultas : Teknik" atau "FAKULTAS TEKNIK"
			parts := splitKeyValue(line)
			if parts != "" {
				data.Fakultas = toTitleCase(parts)
			} else {
				// Cek apakah "FAKULTAS TEKNIK" format (value dalam line yang sama)
				re := regexp.MustCompile(`(?i)fakultas\s*[:\-]?\s*(.+)`)
				m := re.FindStringSubmatch(line)
				if len(m) > 1 && strings.TrimSpace(m[1]) != "" {
					data.Fakultas = toTitleCase(strings.TrimSpace(m[1]))
				} else if i+1 < len(cleanLines) && !isKeyword(cleanLines[i+1]) {
					data.Fakultas = toTitleCase(cleanLines[i+1])
				}
			}
			break
		}
	}

	// ── Extract Prodi / Program Studi / Jurusan ──
	for i, line := range cleanLines {
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "PROGRAM STUDI") || strings.Contains(upper, "PRODI") || strings.Contains(upper, "JURUSAN") {
			parts := splitKeyValue(line)
			if parts != "" {
				data.Prodi = toTitleCase(parts)
			} else {
				re := regexp.MustCompile(`(?i)(?:program\s+studi|prodi|jurusan)\s*[:\-]?\s*(.+)`)
				m := re.FindStringSubmatch(line)
				if len(m) > 1 && strings.TrimSpace(m[1]) != "" {
					data.Prodi = toTitleCase(strings.TrimSpace(m[1]))
				} else if i+1 < len(cleanLines) && !isKeyword(cleanLines[i+1]) {
					data.Prodi = toTitleCase(cleanLines[i+1])
				}
			}
			break
		}
	}

	return data
}

// splitKeyValue memisahkan "Key : Value" atau "Key: Value" dan return value
func splitKeyValue(line string) string {
	// Coba split dengan ":" atau "-"
	for _, sep := range []string{":", "-"} {
		idx := strings.Index(line, sep)
		if idx > 0 && idx < len(line)-1 {
			value := strings.TrimSpace(line[idx+1:])
			if value != "" {
				return value
			}
		}
	}
	return ""
}

// indexOf mencari index string dalam slice
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// isKeyword mengecek apakah line adalah keyword/label (bukan nilai)
func isKeyword(line string) bool {
	upper := strings.ToUpper(strings.TrimSpace(line))
	keywords := []string{"NAMA", "NPM", "NIM", "NRP", "FAKULTAS", "PROGRAM STUDI", "PRODI", "JURUSAN",
		"TEMPAT", "TANGGAL", "ALAMAT", "JENIS KELAMIN", "AGAMA", "NO", "BERLAKU", "KARTU"}
	for _, k := range keywords {
		if strings.HasPrefix(upper, k) {
			return true
		}
	}
	return false
}

// toTitleCase mengubah string ke Title Case
func toTitleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}
