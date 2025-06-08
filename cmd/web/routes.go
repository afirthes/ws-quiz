package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/afirthes/ws-quiz/internal/handlers"
	"github.com/afirthes/ws-quiz/template"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Claim struct {
	ID      string          `json:"id"`
	Details json.RawMessage `json:"details"`
}

type Image struct {
	ID          int    `json:"id"`
	ClaimID     string `json:"claim_id"`
	Filename    string `json:"filename"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func routes(rh *handlers.RestHandlers, wsh *handlers.WsHandlers) http.Handler {
	r := chi.NewRouter()

	r.Get("/", rh.LoadImageMain)
	r.Get("/ws", wsh.WsEndpoint)
	r.Post("/upload", uploadHandler)

	r.Get("/c/{id}", resultPage)

	r.Route("/claims", func(r chi.Router) {
		r.Post("/", createClaimHandler)
		r.Get("/{id}", getClaimHandler)
		r.Put("/{id}", updateClaimHandler)
		r.Post("/{id}/process", ProcessClaim)

		r.Route("/{id}/images", func(r chi.Router) {
			r.Get("/", getClaimImagesHandler)
			r.Post("/", addClaimImageHandler)
			r.Delete("/", deleteAllClaimImages)
		})

	})

	// Статические файлы
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.Handle("/static/*", fileServer)

	return r
}

func ProcessClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	if claimID == "" {
		http.Error(w, "missing claim id", http.StatusBadRequest)
		return
	}

	// 1. Парсим список файлов из тела запроса
	var payload struct {
		Files []string `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(payload.Files) == 0 {
		http.Error(w, "no files provided", http.StatusBadRequest)
		return
	}

	// 2. Формируем JSON для Python-сервиса
	pythonReq := map[string]interface{}{
		"files": payload.Files,
	}
	body, _ := json.Marshal(pythonReq)

	// 3. Отправляем в Python-сервис
	req, err := http.NewRequest("POST", "http://localhost:8000/check_car", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Starting processing of claim %s\n", claimID)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to process claim '%s': %v", claimID, err)
		http.Error(w, "python service error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	log.Printf("Ended processing of claim %s\n", claimID)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 207 {
		http.Error(w, "processing failed", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteAllClaimImages(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	if claimID == "" {
		http.Error(w, "claimID is required", http.StatusBadRequest)
		return
	}

	// Удалить изображения с диска
	dirPath := fmt.Sprintf("./uploads/%s", claimID)
	if err := os.RemoveAll(dirPath); err != nil {
		http.Error(w, "failed to delete image files", http.StatusInternalServerError)
		return
	}

	// Удалить записи из базы (если хранятся)
	if _, err := db.Exec(`DELETE FROM claim_images WHERE claim_id = $1`, claimID); err != nil {
		http.Error(w, "failed to delete DB records", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	claimID := r.URL.Query().Get("claim_id")
	if claimID == "" {
		http.Error(w, "Missing claim_id", http.StatusBadRequest)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Can't parse multipart", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	for _, fileHeader := range files {
		src, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Can't open file", http.StatusInternalServerError)
			return
		}
		defer src.Close()

		// расширение
		ext := filepath.Ext(fileHeader.Filename)
		ext = strings.ToLower(ext)
		if ext == "" {
			ext = ".bin"
		}

		// рандомное имя
		random := make([]byte, 16)
		rand.Read(random)
		randomName := hex.EncodeToString(random) + ext

		// сохраняем файл
		savePath := filepath.Join("static/uploads", randomName)
		dst, err := os.Create(savePath)
		if err != nil {
			http.Error(w, "Can't save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, src)

		// сохраняем в БД
		_, err = db.Exec(`
			INSERT INTO claim_images (claim_id, filename, description, type)
			VALUES ($1, $2, '', 'photo')
		`, claimID, randomName)
		if err != nil {
			http.Error(w, "Can't insert image record", http.StatusInternalServerError)
			return
		}
	}

	w.Write([]byte("OK"))
}

func createClaimHandler(w http.ResponseWriter, r *http.Request) {
	id := ksuid.New().String()
	var c Claim
	if err := json.NewDecoder(r.Body).Decode(&c.Details); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`INSERT INTO claims (id, details) VALUES ($1, $2)`, id, c.Details)
	if err != nil {
		http.Error(w, "failed to insert", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func getClaimHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c Claim
	err := db.QueryRow(`SELECT id, details FROM claims WHERE id = $1`, id).Scan(&c.ID, &c.Details)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "query error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(c)
}

func updateClaimHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c Claim
	if err := json.NewDecoder(r.Body).Decode(&c.Details); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	_, err := db.Exec(`UPDATE claims SET details = $1 WHERE id = $2`, c.Details, id)
	if err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getClaimImagesHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rows, err := db.Query(`SELECT id, claim_id, filename, description, type FROM claim_images WHERE claim_id = $1`, id)
	if err != nil {
		http.Error(w, "query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.ID, &img.ClaimID, &img.Filename, &img.Description, &img.Type); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		images = append(images, img)
	}

	// всегда JSON
	w.Header().Set("Content-Type", "application/json")

	// если nil — вернуть пустой массив []
	if images == nil {
		images = []Image{}
	}

	json.NewEncoder(w).Encode(images)
}

func addClaimImageHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var img Image
	img.ClaimID = id
	if err := json.NewDecoder(r.Body).Decode(&img); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`INSERT INTO claim_images (claim_id, filename, description, type) VALUES ($1, $2, $3, $4)`,
		img.ClaimID, img.Filename, img.Description, img.Type)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func resultPage(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")

	// Получаем заявку
	var claim template.Claim
	err := db.QueryRow(`SELECT id, details FROM claims WHERE id = $1`, claimID).Scan(&claim.ID, &claim.Details)
	if err != nil {
		http.Error(w, "Claim not found", http.StatusNotFound)
		return
	}

	// Получаем изображения
	rows, err := db.Query(`SELECT id, claim_id, filename, description, type FROM claim_images WHERE claim_id = $1`, claimID)
	if err != nil {
		http.Error(w, "Failed to load images", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var images []template.Image
	for rows.Next() {
		var img template.Image
		if err := rows.Scan(&img.ID, &img.ClaimID, &img.Filename, &img.Description, &img.Type); err != nil {
			http.Error(w, "Failed to parse image row", http.StatusInternalServerError)
			return
		}
		images = append(images, img)
	}

	// Отрисовываем templ-шаблон
	template.Result(claim, images).Render(r.Context(), w)
}
