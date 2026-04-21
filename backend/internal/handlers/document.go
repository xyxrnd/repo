package handlers

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"repository-un/internal/config"
	"repository-un/internal/middleware"
	"repository-un/internal/models"
	"repository-un/internal/utils"

	"github.com/google/uuid"
)

// DocumentsHandler menangani operasi list dan create dokumen
// GET /api/documents - List semua dokumen
// POST /api/documents - Upload dokumen baru
func DocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	switch r.Method {
	case http.MethodGet:
		listDocuments(w, r)
	case http.MethodPost:
		createDocument(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DocumentByIdHandler menangani operasi pada dokumen tertentu
// GET /api/documents/:id - Get dokumen by ID
// PUT /api/documents/:id - Update dokumen
// DELETE /api/documents/:id - Delete dokumen
func DocumentByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	path := strings.TrimPrefix(r.URL.Path, "/api/documents/")

	// Cek apakah ini request download-all: /api/documents/{id}/download-all
	if strings.HasSuffix(path, "/download-all") {
		DownloadAllHandler(w, r)
		return
	}

	id := path
	if id == "" {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getDocumentById(w, r, id)
	case http.MethodPut:
		updateDocument(w, r, id)
	case http.MethodDelete:
		deleteDocument(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listDocuments mengambil semua dokumen dari database beserta fakultas, prodi, dan files
func listDocuments(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(context.Background(),
		`SELECT d.id, d.judul, d.penulis, COALESCE(d.npm, '') as npm, COALESCE(d.abstrak, '') as abstrak, d.jenis_file, d.status, d.created_at,
		        COALESCE(d.fakultas_id::text, '') as fakultas_id,
		        COALESCE(f.nama, '') as fakultas_nama,
		        COALESCE(d.prodi_id::text, '') as prodi_id,
		        COALESCE(p.nama, '') as prodi_nama,
		        COALESCE(d.dosen_pembimbing, '') as dosen_pembimbing,
		        COALESCE(d.dosen_pembimbing_2, '') as dosen_pembimbing_2,
		        COALESCE(d.kata_kunci, '') as kata_kunci,
		        COALESCE(d.tahun, 0) as tahun,
		        COALESCE(d.view_count, 0) as view_count
		 FROM documents d
		 LEFT JOIN fakultas f ON d.fakultas_id = f.id
		 LEFT JOIN prodi p ON d.prodi_id = p.id
		 ORDER BY d.created_at DESC`)
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	documents := []models.Document{}

	for rows.Next() {
		var d models.Document
		err := rows.Scan(
			&d.ID,
			&d.Judul,
			&d.Penulis,
			&d.NPM,
			&d.Abstrak,
			&d.JenisFile,
			&d.Status,
			&d.CreatedAt,
			&d.FakultasID,
			&d.FakultasNama,
			&d.ProdiID,
			&d.ProdiNama,
			&d.DosenPembimbing,
			&d.DosenPembimbing2,
			&d.KataKunci,
			&d.Tahun,
			&d.ViewCount,
		)
		if err != nil {
			http.Error(w, "Gagal membaca data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Ambil files untuk dokumen ini
		d.Files = getDocumentFiles(d.ID)

		documents = append(documents, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// getDocumentFiles mengambil semua file terkait dokumen
func getDocumentFiles(documentID string) []models.DocumentFile {
	rows, err := config.DB.Query(context.Background(),
		`SELECT id, document_id, file_name, file_path, file_size, file_order, COALESCE(is_locked, false)
		 FROM document_files
		 WHERE document_id = $1
		 ORDER BY file_order ASC`, documentID)
	if err != nil {
		return []models.DocumentFile{}
	}
	defer rows.Close()

	files := []models.DocumentFile{}
	for rows.Next() {
		var f models.DocumentFile
		err := rows.Scan(&f.ID, &f.DocumentID, &f.FileName, &f.FilePath, &f.FileSize, &f.FileOrder, &f.IsLocked)
		if err != nil {
			continue
		}
		files = append(files, f)
	}
	return files
}

// getDocumentById mengambil dokumen berdasarkan ID
func getDocumentById(w http.ResponseWriter, r *http.Request, id string) {
	var d models.Document
	err := config.DB.QueryRow(context.Background(),
		`SELECT d.id, d.judul, d.penulis, COALESCE(d.npm, '') as npm, COALESCE(d.abstrak, '') as abstrak, d.jenis_file, d.status, d.created_at,
		        COALESCE(d.fakultas_id::text, '') as fakultas_id,
		        COALESCE(f.nama, '') as fakultas_nama,
		        COALESCE(d.prodi_id::text, '') as prodi_id,
		        COALESCE(p.nama, '') as prodi_nama,
		        COALESCE(d.dosen_pembimbing, '') as dosen_pembimbing,
		        COALESCE(d.dosen_pembimbing_2, '') as dosen_pembimbing_2,
		        COALESCE(d.kata_kunci, '') as kata_kunci,
		        COALESCE(d.tahun, 0) as tahun,
		        COALESCE(d.view_count, 0) as view_count
		 FROM documents d
		 LEFT JOIN fakultas f ON d.fakultas_id = f.id
		 LEFT JOIN prodi p ON d.prodi_id = p.id
		 WHERE d.id = $1`, id).Scan(
		&d.ID,
		&d.Judul,
		&d.Penulis,
		&d.NPM,
		&d.Abstrak,
		&d.JenisFile,
		&d.Status,
		&d.CreatedAt,
		&d.FakultasID,
		&d.FakultasNama,
		&d.ProdiID,
		&d.ProdiNama,
		&d.DosenPembimbing,
		&d.DosenPembimbing2,
		&d.KataKunci,
		&d.Tahun,
		&d.ViewCount,
	)

	if err != nil {
		http.Error(w, "Dokumen tidak ditemukan", http.StatusNotFound)
		return
	}

	// Catat view dokumen
	recordDocumentView(id, r)

	// Ambil files
	d.Files = getDocumentFiles(d.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// createDocument membuat dokumen baru — file disimpan lokal dulu, lalu upload ke Google Drive di background
func createDocument(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(100 << 20) // 100 MB max

	judul := r.FormValue("title")
	penulis := r.FormValue("author")
	npm := r.FormValue("npm")
	abstrak := r.FormValue("abstrak")
	jenisFile := r.FormValue("category")
	status := r.FormValue("status")
	fakultasID := r.FormValue("fakultas_id")
	prodiID := r.FormValue("prodi_id")
	dosenPembimbing := r.FormValue("dosen_pembimbing")
	dosenPembimbing2 := r.FormValue("dosen_pembimbing_2")
	kataKunci := r.FormValue("kata_kunci")
	tahunStr := r.FormValue("tahun")
	tahun := 0
	if tahunStr != "" {
		if parsed, err := strconv.Atoi(tahunStr); err == nil {
			tahun = parsed
		}
	}

	if judul == "" || penulis == "" || jenisFile == "" {
		http.Error(w, "Metadata tidak lengkap (judul, penulis, jenis_file wajib)", http.StatusBadRequest)
		return
	}

	if status == "" {
		status = "draft"
	}

	// Pastikan folder uploads ada (untuk penyimpanan sementara)
	os.MkdirAll("uploads", os.ModePerm)

	docID := uuid.New()

	// Parse lock info per file (comma-separated: "true,false,true")
	fileLocks := strings.Split(r.FormValue("file_locks"), ",")

	// STEP 1: Simpan file ke LOKAL dulu (agar response cepat)
	type savedFile struct {
		FileName  string
		LocalPath string // path lokal sementara
		FileSize  int64
		Order     int
		IsLocked  bool
	}

	var mainFilePath string
	var savedFiles []savedFile
	multiFiles := r.MultipartForm.File["files"]

	if len(multiFiles) == 0 {
		// Fallback: coba single file dengan key "file"
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File tidak ditemukan. Minimal 1 file harus diupload.", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		storedName := uuid.New().String() + ext
		localPath := "uploads/" + storedName

		dst, err := os.Create(localPath)
		if err != nil {
			http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)

		mainFilePath = localPath
		isLocked := len(fileLocks) > 0 && fileLocks[0] == "true"
		savedFiles = append(savedFiles, savedFile{
			FileName:  header.Filename,
			LocalPath: localPath,
			FileSize:  header.Size,
			Order:     0,
			IsLocked:  isLocked,
		})
	} else {
		// Multiple files — simpan ke lokal
		for i, fileHeader := range multiFiles {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			ext := filepath.Ext(fileHeader.Filename)
			storedName := uuid.New().String() + ext
			localPath := "uploads/" + storedName

			dst, err := os.Create(localPath)
			if err != nil {
				file.Close()
				continue
			}

			io.Copy(dst, file)
			dst.Close()
			file.Close()

			if i == 0 {
				mainFilePath = localPath
			}

			isLocked := i < len(fileLocks) && fileLocks[i] == "true"
			savedFiles = append(savedFiles, savedFile{
				FileName:  fileHeader.Filename,
				LocalPath: localPath,
				FileSize:  fileHeader.Size,
				Order:     i,
				IsLocked:  isLocked,
			})
		}
	}

	// STEP 2: Insert dokumen ke database dengan path lokal (response cepat)
	query := `
		INSERT INTO documents (id, judul, penulis, npm, abstrak, jenis_file, file_path, status, fakultas_id, prodi_id, dosen_pembimbing, dosen_pembimbing_2, kata_kunci, tahun)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
			CASE WHEN $9 = '' THEN NULL ELSE $9::uuid END,
			CASE WHEN $10 = '' THEN NULL ELSE $10::uuid END,
			$11, $12, $13, $14)
	`
	_, err := config.DB.Exec(context.Background(), query,
		docID, judul, penulis, npm, abstrak, jenisFile, mainFilePath, status, fakultasID, prodiID, dosenPembimbing, dosenPembimbing2, kataKunci, tahun)

	if err != nil {
		http.Error(w, "Gagal menyimpan metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// STEP 3: Insert file records ke document_files dengan path lokal
	type fileRecord struct {
		DbFileID  string
		LocalPath string
		FileName  string
	}
	var fileRecords []fileRecord

	for _, sf := range savedFiles {
		dbFileID := uuid.New().String()
		config.DB.Exec(context.Background(),
			`INSERT INTO document_files (id, document_id, file_name, file_path, file_size, file_order, is_locked)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			dbFileID, docID, sf.FileName, sf.LocalPath, sf.FileSize, sf.Order, sf.IsLocked)

		fileRecords = append(fileRecords, fileRecord{
			DbFileID:  dbFileID,
			LocalPath: sf.LocalPath,
			FileName:  sf.FileName,
		})
	}

	// STEP 4: Response langsung ke user (tidak menunggu upload GDrive)
	doc := models.Document{
		ID:               docID.String(),
		Judul:            judul,
		Penulis:          penulis,
		NPM:              npm,
		Abstrak:          abstrak,
		JenisFile:        jenisFile,
		FakultasID:       fakultasID,
		ProdiID:          prodiID,
		DosenPembimbing:  dosenPembimbing,
		DosenPembimbing2: dosenPembimbing2,
		KataKunci:        kataKunci,
		Status:           status,
		Files:            getDocumentFiles(docID.String()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)

	// STEP 5: Upload ke Google Drive di BACKGROUND (async)
	go func() {
		// Resolve nama fakultas dan prodi
		var fakultasNama, prodiNama string
		if fakultasID != "" {
			config.DB.QueryRow(context.Background(),
				`SELECT COALESCE(nama, '') FROM fakultas WHERE id = $1`, fakultasID).Scan(&fakultasNama)
		}
		if prodiID != "" {
			config.DB.QueryRow(context.Background(),
				`SELECT COALESCE(nama, '') FROM prodi WHERE id = $1`, prodiID).Scan(&prodiNama)
		}

		// Buat folder hierarchy di Google Drive
		gdriveFolderID, err := utils.GetDocumentFolderID(fakultasNama, prodiNama, tahun, penulis)
		if err != nil {
			fmt.Printf("⚠️ [Background] Gagal membuat folder GDrive untuk doc %s: %v\n", docID, err)
			return
		}

		fmt.Printf("📤 [Background] Mulai upload %d file ke Google Drive untuk dokumen '%s'...\n", len(fileRecords), judul)

		for _, fr := range fileRecords {
			// Buka file lokal
			localFile, err := os.Open(fr.LocalPath)
			if err != nil {
				fmt.Printf("⚠️ [Background] Gagal membuka file lokal %s: %v\n", fr.LocalPath, err)
				continue
			}

			// Upload ke Google Drive
			uploadResult, err := utils.UploadToGDriveFolder(localFile, fr.FileName, gdriveFolderID)
			localFile.Close()
			if err != nil {
				fmt.Printf("⚠️ [Background] Gagal upload '%s' ke GDrive: %v\n", fr.FileName, err)
				continue
			}

			// Update database: ganti path lokal → GDrive file ID
			config.DB.Exec(context.Background(),
				`UPDATE document_files SET file_path = $1 WHERE id = $2`,
				uploadResult.FileID, fr.DbFileID)

			// Update main file_path juga jika ini file pertama
			if fr.LocalPath == mainFilePath {
				config.DB.Exec(context.Background(),
					`UPDATE documents SET file_path = $1 WHERE id = $2`,
					uploadResult.FileID, docID)
			}

			// Hapus file lokal setelah berhasil upload ke GDrive
			os.Remove(fr.LocalPath)
			fmt.Printf("✅ [Background] '%s' berhasil diupload ke GDrive & file lokal dihapus\n", fr.FileName)
		}

		fmt.Printf("✅ [Background] Semua file dokumen '%s' selesai diupload ke Google Drive\n", judul)
	}()
}

// updateDocument mengupdate dokumen
func updateDocument(w http.ResponseWriter, r *http.Request, id string) {
	r.ParseMultipartForm(100 << 20) // 100 MB max

	judul := r.FormValue("title")
	penulis := r.FormValue("author")
	npm := r.FormValue("npm")
	abstrak := r.FormValue("abstrak")
	jenisFile := r.FormValue("category")
	status := r.FormValue("status")
	fakultasID := r.FormValue("fakultas_id")
	prodiID := r.FormValue("prodi_id")
	dosenPembimbing := r.FormValue("dosen_pembimbing")
	dosenPembimbing2 := r.FormValue("dosen_pembimbing_2")
	kataKunci := r.FormValue("kata_kunci")
	tahunStr := r.FormValue("tahun")
	tahun := 0
	if tahunStr != "" {
		if parsed, err := strconv.Atoi(tahunStr); err == nil {
			tahun = parsed
		}
	}

	if judul == "" || penulis == "" || jenisFile == "" {
		http.Error(w, "Metadata tidak lengkap", http.StatusBadRequest)
		return
	}

	if status == "" {
		status = "draft"
	}

	// Update metadata dokumen
	query := `
		UPDATE documents
		SET judul = $1, penulis = $2, npm = $3, abstrak = $4, jenis_file = $5, status = $6,
		    fakultas_id = CASE WHEN $7 = '' THEN NULL ELSE $7::uuid END,
		    prodi_id = CASE WHEN $8 = '' THEN NULL ELSE $8::uuid END,
		    dosen_pembimbing = $9,
		    dosen_pembimbing_2 = $10,
		    kata_kunci = $11,
		    tahun = $12
		WHERE id = $13
	`
	_, err := config.DB.Exec(context.Background(), query,
		judul, penulis, npm, abstrak, jenisFile, status, fakultasID, prodiID, dosenPembimbing, dosenPembimbing2, kataKunci, tahun, id)

	if err != nil {
		http.Error(w, "Gagal update dokumen: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle file management
	existingFilesJSON := r.FormValue("existing_files")
	fileLocks := strings.Split(r.FormValue("file_locks"), ",")

	// Resolve nama fakultas dan prodi untuk folder Google Drive
	var fakultasNama, prodiNama string
	if fakultasID != "" {
		config.DB.QueryRow(context.Background(),
			`SELECT COALESCE(nama, '') FROM fakultas WHERE id = $1`, fakultasID).Scan(&fakultasNama)
	}
	if prodiID != "" {
		config.DB.QueryRow(context.Background(),
			`SELECT COALESCE(nama, '') FROM prodi WHERE id = $1`, prodiID).Scan(&prodiNama)
	}

	// Dapatkan folder ID di Google Drive
	gdriveFolderID, gdriveErr := utils.GetDocumentFolderID(fakultasNama, prodiNama, tahun, penulis)
	if gdriveErr != nil {
		fmt.Printf("⚠️ Gagal membuat folder hierarchy di Google Drive: %v\n", gdriveErr)
	}

	if existingFilesJSON != "" {
		// Mode baru: kelola file per-item
		type keepFile struct {
			ID    string `json:"id"`
			Order int    `json:"order"`
		}
		var keepFiles []keepFile
		if err := json.Unmarshal([]byte(existingFilesJSON), &keepFiles); err == nil {
			oldFiles := getDocumentFiles(id)

			keepMap := map[string]int{}
			for _, kf := range keepFiles {
				keepMap[kf.ID] = kf.Order
			}

			// Hapus file yang TIDAK ada di keepFiles
			for _, of := range oldFiles {
				if _, keep := keepMap[of.ID]; !keep {
					if utils.IsGDriveID(of.FilePath) {
						utils.DeleteFromGDrive(of.FilePath)
					} else {
						os.Remove(of.FilePath)
					}
					config.DB.Exec(context.Background(),
						`DELETE FROM document_files WHERE id = $1`, of.ID)
				}
			}

			for _, kf := range keepFiles {
				config.DB.Exec(context.Background(),
					`UPDATE document_files SET file_order = $1 WHERE id = $2`,
					kf.Order, kf.ID)
			}

			// Upload file baru ke Google Drive
			multiFiles := r.MultipartForm.File["files"]
			newFileOrders := strings.Split(r.FormValue("new_file_orders"), ",")
			for i, fileHeader := range multiFiles {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}

				var filePath string
				if gdriveFolderID != "" {
					uploadResult, err := utils.UploadToGDriveFolder(file, fileHeader.Filename, gdriveFolderID)
					file.Close()
					if err != nil {
						fmt.Printf("⚠️ Gagal upload file '%s' ke Google Drive: %v\n", fileHeader.Filename, err)
						continue
					}
					filePath = uploadResult.FileID
				} else {
					file.Close()
					continue
				}

				order := i
				if i < len(newFileOrders) {
					if parsed, err := strconv.Atoi(strings.TrimSpace(newFileOrders[i])); err == nil {
						order = parsed
					}
				}

				isLocked := i < len(fileLocks) && fileLocks[i] == "true"
				fileID := uuid.New()
				config.DB.Exec(context.Background(),
					`INSERT INTO document_files (id, document_id, file_name, file_path, file_size, file_order, is_locked)
					 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
					fileID, id, fileHeader.Filename, filePath, fileHeader.Size, order, isLocked)
			}

			updatedFiles := getDocumentFiles(id)
			if len(updatedFiles) > 0 {
				config.DB.Exec(context.Background(),
					`UPDATE documents SET file_path = $1 WHERE id = $2`, updatedFiles[0].FilePath, id)
			}
		}
	} else {
		// Mode lama: jika ada files baru, ganti semua
		multiFiles := r.MultipartForm.File["files"]
		if len(multiFiles) > 0 {
			oldFiles := getDocumentFiles(id)
			for _, of := range oldFiles {
				if utils.IsGDriveID(of.FilePath) {
					utils.DeleteFromGDrive(of.FilePath)
				} else {
					os.Remove(of.FilePath)
				}
			}
			config.DB.Exec(context.Background(),
				`DELETE FROM document_files WHERE document_id = $1`, id)

			var firstFilePath string
			for i, fileHeader := range multiFiles {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}

				var filePath string
				if gdriveFolderID != "" {
					uploadResult, err := utils.UploadToGDriveFolder(file, fileHeader.Filename, gdriveFolderID)
					file.Close()
					if err != nil {
						fmt.Printf("⚠️ Gagal upload file '%s' ke Google Drive: %v\n", fileHeader.Filename, err)
						continue
					}
					filePath = uploadResult.FileID
				} else {
					file.Close()
					continue
				}

				if i == 0 {
					firstFilePath = filePath
				}

				isLocked := i < len(fileLocks) && fileLocks[i] == "true"
				fileID := uuid.New()
				config.DB.Exec(context.Background(),
					`INSERT INTO document_files (id, document_id, file_name, file_path, file_size, file_order, is_locked)
					 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
					fileID, id, fileHeader.Filename, filePath, fileHeader.Size, i, isLocked)
			}

			if firstFilePath != "" {
				config.DB.Exec(context.Background(),
					`UPDATE documents SET file_path = $1 WHERE id = $2`, firstFilePath, id)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                 id,
		"judul":              judul,
		"penulis":            penulis,
		"npm":                npm,
		"abstrak":            abstrak,
		"jenis_file":         jenisFile,
		"status":             status,
		"fakultas_id":        fakultasID,
		"prodi_id":           prodiID,
		"dosen_pembimbing":   dosenPembimbing,
		"dosen_pembimbing_2": dosenPembimbing2,
		"kata_kunci":         kataKunci,
	})
}

// deleteDocument menghapus dokumen dan file-filenya (dari Google Drive atau lokal)
func deleteDocument(w http.ResponseWriter, r *http.Request, id string) {
	// Hapus semua file (Google Drive atau lokal)
	files := getDocumentFiles(id)
	for _, f := range files {
		if utils.IsGDriveID(f.FilePath) {
			if err := utils.DeleteFromGDrive(f.FilePath); err != nil {
				fmt.Printf("⚠️ Gagal hapus file dari Google Drive: %v\n", err)
			}
		} else {
			os.Remove(f.FilePath)
		}
	}

	// Hapus juga file_path lama (jika berbeda dari document_files)
	var filePath string
	config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(file_path, '') FROM documents WHERE id = $1`, id).Scan(&filePath)
	if filePath != "" {
		if utils.IsGDriveID(filePath) {
			utils.DeleteFromGDrive(filePath)
		} else {
			os.Remove(filePath)
		}
	}

	// document_files otomatis terhapus karena ON DELETE CASCADE
	result, err := config.DB.Exec(context.Background(),
		`DELETE FROM documents WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Gagal menghapus data", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Dokumen tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Dokumen berhasil dihapus"}`))
}

// recordDocumentView mencatat view dokumen unik per IP per hari
// Jika IP yang sama sudah melihat dokumen ini hari ini, tidak dihitung lagi
func recordDocumentView(documentID string, r *http.Request) {
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}

	// Atomic insert: jika sudah ada view dari IP ini hari ini, DO NOTHING
	viewID := uuid.New()
	result, err := config.DB.Exec(context.Background(),
		`INSERT INTO document_views (id, document_id, ip_address)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (document_id, ip_address, (viewed_at::date)) DO NOTHING`,
		viewID, documentID, ipAddress)

	if err != nil || result.RowsAffected() == 0 {
		// Error atau sudah pernah melihat hari ini
		return
	}

	// Baru pertama kali melihat hari ini, increment view_count
	config.DB.Exec(context.Background(),
		`UPDATE documents SET view_count = COALESCE(view_count, 0) + 1 WHERE id = $1`,
		documentID)
}

// PopularDocumentsHandler menangani request dokumen populer
// GET /api/documents/popular?limit=6
func PopularDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
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
		`SELECT d.id, d.judul, d.penulis, COALESCE(d.npm, '') as npm, COALESCE(d.abstrak, '') as abstrak, d.jenis_file, d.status, d.created_at,
		        COALESCE(d.fakultas_id::text, '') as fakultas_id,
		        COALESCE(f.nama, '') as fakultas_nama,
		        COALESCE(d.prodi_id::text, '') as prodi_id,
		        COALESCE(p.nama, '') as prodi_nama,
		        COALESCE(d.dosen_pembimbing, '') as dosen_pembimbing,
		        COALESCE(d.dosen_pembimbing_2, '') as dosen_pembimbing_2,
		        COALESCE(d.kata_kunci, '') as kata_kunci,
		        COALESCE(d.view_count, 0) as view_count
		 FROM documents d
		 LEFT JOIN fakultas f ON d.fakultas_id = f.id
		 LEFT JOIN prodi p ON d.prodi_id = p.id
		 ORDER BY d.view_count DESC NULLS LAST, d.created_at DESC
		 LIMIT $1`, limit)
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	documents := []models.Document{}
	for rows.Next() {
		var d models.Document
		err := rows.Scan(
			&d.ID,
			&d.Judul,
			&d.Penulis,
			&d.NPM,
			&d.Abstrak,
			&d.JenisFile,
			&d.Status,
			&d.CreatedAt,
			&d.FakultasID,
			&d.FakultasNama,
			&d.ProdiID,
			&d.ProdiNama,
			&d.DosenPembimbing,
			&d.DosenPembimbing2,
			&d.KataKunci,
			&d.ViewCount,
		)
		if err != nil {
			continue
		}
		d.Files = getDocumentFiles(d.ID)
		documents = append(documents, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// DownloadHandler menangani download dokumen
// GET /download/:id
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/download/")
	id := strings.TrimSuffix(path, "/download")

	if id == "" {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	// Catat view dokumen saat download
	recordDocumentView(id, r)

	var fp string
	err := config.DB.QueryRow(context.Background(),
		`SELECT file_path FROM documents WHERE id = $1`, id).Scan(&fp)

	if err != nil {
		http.Error(w, "File tidak ditemukan", http.StatusNotFound)
		return
	}

	// Jika file di Google Drive, download dari sana
	if utils.IsGDriveID(fp) {
		body, mimeType, err := utils.GetGDriveFileContent(fp)
		if err != nil {
			http.Error(w, "Gagal mengambil file dari Google Drive: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer body.Close()

		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Content-Disposition", "attachment")
		io.Copy(w, body)
		return
	}

	// File lokal (legacy)
	w.Header().Set("Content-Disposition", "attachment")
	http.ServeFile(w, r, fp)
}

// DownloadAllHandler menangani download semua file dokumen dalam bentuk ZIP
// GET /api/documents/:id/download-all
func DownloadAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID dari URL: /api/documents/{id}/download-all
	path := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	id := strings.TrimSuffix(path, "/download-all")

	if id == "" {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	// Ambil judul dokumen untuk nama file ZIP
	var judul string
	err := config.DB.QueryRow(context.Background(),
		`SELECT judul FROM documents WHERE id = $1`, id).Scan(&judul)
	if err != nil {
		http.Error(w, "Dokumen tidak ditemukan", http.StatusNotFound)
		return
	}

	// Ambil semua file terkait dokumen
	files := getDocumentFiles(id)

	// Jika tidak ada file di document_files, fallback ke file_path utama
	if len(files) == 0 {
		var fp string
		config.DB.QueryRow(context.Background(),
			`SELECT COALESCE(file_path, '') FROM documents WHERE id = $1`, id).Scan(&fp)
		if fp != "" {
			if utils.IsGDriveID(fp) {
				body, mimeType, err := utils.GetGDriveFileContent(fp)
				if err != nil {
					http.Error(w, "Gagal mengambil file dari Google Drive", http.StatusInternalServerError)
					return
				}
				defer body.Close()
				w.Header().Set("Content-Type", mimeType)
				w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, judul))
				io.Copy(w, body)
				return
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, judul))
			http.ServeFile(w, r, fp)
			return
		}
		http.Error(w, "Tidak ada file untuk dokumen ini", http.StatusNotFound)
		return
	}

	// Jika hanya 1 file, langsung download
	if len(files) == 1 {
		if utils.IsGDriveID(files[0].FilePath) {
			body, mimeType, err := utils.GetGDriveFileContent(files[0].FilePath)
			if err != nil {
				http.Error(w, "Gagal mengambil file dari Google Drive", http.StatusInternalServerError)
				return
			}
			defer body.Close()
			w.Header().Set("Content-Type", mimeType)
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, files[0].FileName))
			io.Copy(w, body)
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, files[0].FileName))
		http.ServeFile(w, r, files[0].FilePath)
		return
	}

	// Multiple files: buat ZIP
	safeJudul := strings.ReplaceAll(judul, " ", "_")
	safeJudul = strings.ReplaceAll(safeJudul, "/", "_")
	zipFileName := safeJudul + ".zip"

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, zipFileName))

	// Stream ZIP langsung ke response writer
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, file := range files {
		var reader io.ReadCloser

		if utils.IsGDriveID(file.FilePath) {
			// Download dari Google Drive
			body, _, err := utils.GetGDriveFileContent(file.FilePath)
			if err != nil {
				fmt.Printf("⚠️ Gagal download file dari GDrive untuk ZIP: %v\n", err)
				continue
			}
			reader = body
		} else {
			// Buka file dari disk (legacy)
			f, err := os.Open(file.FilePath)
			if err != nil {
				continue
			}
			reader = f
		}

		// Buat entry di ZIP dengan nama file asli
		zEntry, err := zipWriter.Create(file.FileName)
		if err != nil {
			reader.Close()
			continue
		}

		// Copy isi file ke ZIP
		io.Copy(zEntry, reader)
		reader.Close()
	}
}

// UploadHandler menangani upload file legacy
// POST /uploads
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	judul := r.FormValue("judul")
	penulis := r.FormValue("penulis")
	jenisFile := r.FormValue("jenis_file")

	if judul == "" || penulis == "" || jenisFile == "" {
		http.Error(w, "Metadata tidak lengkap", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File tidak ditemukan", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	storedName := uuid.New().String() + ext
	filePath := "uploads/" + storedName

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	id := uuid.New()

	query := `
		INSERT INTO documents (id, judul, penulis, jenis_file, file_path, status)
		VALUES ($1, $2, $3, $4, $5, 'draft')
	`

	_, err = config.DB.Exec(context.Background(), query,
		id, judul, penulis, jenisFile, filePath)

	if err != nil {
		http.Error(w, "Gagal menyimpan metadata", http.StatusInternalServerError)
		return
	}

	r.ParseMultipartForm(50 << 20)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"id": "%s",
		"judul": "%s",
		"penulis": "%s",
		"jenis_file": "%s",
		"status": "draft"
	}`, id, judul, penulis, jenisFile)
}

// GDriveProxyHandler proxy gambar dari Google Drive agar bisa di-embed di frontend
// GET /api/gdrive-proxy/{fileId}
func GDriveProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/api/gdrive-proxy/")
	if fileID == "" {
		http.Error(w, "File ID tidak valid", http.StatusBadRequest)
		return
	}

	// Validasi bahwa fileID terlihat seperti Google Drive ID
	if !utils.IsGDriveID(fileID) {
		http.Error(w, "File ID tidak valid", http.StatusBadRequest)
		return
	}

	body, mimeType, err := utils.GetGDriveFileContent(fileID)
	if err != nil {
		http.Error(w, "Gagal mengambil file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache 24 jam
	io.Copy(w, body)
}
