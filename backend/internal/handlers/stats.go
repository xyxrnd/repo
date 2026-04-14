package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"repository-un/internal/config"
	"repository-un/internal/middleware"

	"github.com/google/uuid"
)

// StatsResponse berisi statistik umum untuk landing page
type StatsResponse struct {
	TotalDocuments int `json:"total_documents"`
	TotalAuthors   int `json:"total_authors"`
	TotalVisitors  int `json:"total_visitors"`
	TodayVisitors  int `json:"today_visitors"`
	TotalUsers     int `json:"total_users"`
}

// StatsHandler menangani request statistik publik
// GET /api/stats
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Catat kunjungan situs (unique per IP per hari)
	recordSiteVisit(r)

	var stats StatsResponse

	// Total dokumen yang dipublish
	err := config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM documents WHERE status = 'publish'`).Scan(&stats.TotalDocuments)
	if err != nil {
		stats.TotalDocuments = 0
	}

	// Total penulis unik dari dokumen yang dipublish
	err = config.DB.QueryRow(context.Background(),
		`SELECT COUNT(DISTINCT penulis) FROM documents WHERE status = 'publish'`).Scan(&stats.TotalAuthors)
	if err != nil {
		stats.TotalAuthors = 0
	}

	// Total pengunjung unik keseluruhan (dari site_visits)
	err = config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM site_visits`).Scan(&stats.TotalVisitors)
	if err != nil {
		stats.TotalVisitors = 0
	}

	// Total pengunjung unik hari ini
	err = config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM site_visits WHERE visited_at::date = CURRENT_DATE`).Scan(&stats.TodayVisitors)
	if err != nil {
		stats.TodayVisitors = 0
	}

	// Total pengguna aktif (dari tabel users)
	err = config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		stats.TotalUsers = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// recordSiteVisit mencatat kunjungan website unik per IP per hari
func recordSiteVisit(r *http.Request) {
	ipAddress := r.RemoteAddr
	// Strip port dari RemoteAddr (contoh: "10.10.4.200:54321" -> "10.10.4.200")
	if host, _, err := net.SplitHostPort(ipAddress); err == nil {
		ipAddress = host
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}

	// Cek apakah IP ini sudah tercatat hari ini
	var count int
	err := config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM site_visits
		 WHERE ip_address = $1 AND visited_at::date = CURRENT_DATE`,
		ipAddress).Scan(&count)

	if err != nil || count > 0 {
		return
	}

	// Belum ada, catat kunjungan baru
	visitID := uuid.New()
	config.DB.Exec(context.Background(),
		`INSERT INTO site_visits (id, ip_address) VALUES ($1, $2)`,
		visitID, ipAddress)
}

// TopFakultasItem berisi data fakultas dengan jumlah dokumen
type TopFakultasItem struct {
	ID             string `json:"id"`
	Nama           string `json:"nama"`
	TotalDocuments int    `json:"total_documents"`
}

// TopFakultasHandler menangani request top fakultas berdasarkan jumlah dokumen
// GET /api/stats/top-fakultas?limit=6
func TopFakultasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 6
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 20 {
			limit = parsed
		}
	}

	rows, err := config.DB.Query(context.Background(),
		`SELECT f.id::text, f.nama, COUNT(d.id) as total
		 FROM fakultas f
		 INNER JOIN documents d ON d.fakultas_id = f.id
		 WHERE d.status = 'publish'
		 GROUP BY f.id, f.nama
		 HAVING COUNT(d.id) > 0
		 ORDER BY total DESC
		 LIMIT $1`, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]TopFakultasItem{})
		return
	}
	defer rows.Close()

	result := []TopFakultasItem{}
	for rows.Next() {
		var item TopFakultasItem
		if err := rows.Scan(&item.ID, &item.Nama, &item.TotalDocuments); err != nil {
			continue
		}
		result = append(result, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
