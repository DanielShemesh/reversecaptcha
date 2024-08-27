package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/danielshemesh/reversecaptcha/artists"
	"github.com/danielshemesh/reversecaptcha/jsonextractor"
	"github.com/danielshemesh/reversecaptcha/llm"
)

// AlbumInfo represents the structure of album information
type AlbumInfo struct {
	Title string `json:"title"`
	Day   int    `json:"day"`
	Month int    `json:"month"`
	Year  int    `json:"year"`
}

// Test represents the structure of a test
type Test struct {
	Description string `json:"description"`
	ArtistName  string `json:"artistName"`
}

// AlbumSongsWithTestRequest represents the structure of the request for the album songs with test endpoint
type AlbumSongsWithTestRequest struct {
	AlbumName   string `json:"albumName"`
	ImageBase64 string `json:"imageBase64"`
}

type JSONData struct {
	Score       int    `json:"score"`
	Description string `json:"description"`
}

var (
	tests        = make(map[string]Test)
	mu           sync.RWMutex
	descriptions []string
)

// albumsHandler handles the /albums endpoint
func albumsHandler(w http.ResponseWriter, r *http.Request) {
	artistName := r.URL.Query().Get("artistName")
	if strings.TrimSpace(artistName) == "" {
		http.Error(w, "Missing artistName query parameter", http.StatusBadRequest)
		return
	}

	albums, err := getAlbumsByArtist(artistName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching albums: %v", err), http.StatusInternalServerError)
		return
	}

	sortedAlbums := sortAlbumsByDate(albums)
	albumInfo := extractAlbumInfo(sortedAlbums)

	testID, testDescription, err := generateTest(artistName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating test: %v", err), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "testID",
		Value: testID,
		Path:  "/",
	})

	response := map[string]interface{}{
		"albums":      albumInfo,
		"description": testDescription,
	}

	respondWithJSON(w, response)
}

// getAlbumsByArtist fetches albums for a given artist
func getAlbumsByArtist(artistName string) ([]artists.ReleaseGroup, error) {
	artistID, err := artists.SearchArtist(artistName)
	if err != nil {
		return nil, fmt.Errorf("error searching for artist: %w", err)
	}

	allAlbums, err := artists.GetAlbums(artistID)
	if err != nil {
		return nil, fmt.Errorf("error fetching albums: %w", err)
	}

	var filteredAlbums []artists.ReleaseGroup
	for _, album := range allAlbums {
		if len(album.SecondaryTypes) == 0 {
			filteredAlbums = append(filteredAlbums, album)
		}
	}

	return filteredAlbums, nil
}

// sortAlbumsByDate sorts the albums by release date
func sortAlbumsByDate(albums []artists.ReleaseGroup) []artists.ReleaseGroup {
	sort.Slice(albums, func(i, j int) bool {
		dateI, errI := time.Parse("2006-01-02", albums[i].FirstReleaseDate)
		dateJ, errJ := time.Parse("2006-01-02", albums[j].FirstReleaseDate)
		if errI != nil || errJ != nil {
			return false
		}
		return dateI.Before(dateJ)
	})
	return albums
}

// extractAlbumInfo extracts relevant information from albums
func extractAlbumInfo(albums []artists.ReleaseGroup) []AlbumInfo {
	var result []AlbumInfo
	for _, album := range albums {
		date, err := time.Parse("2006-01-02", album.FirstReleaseDate)
		if err != nil {
			log.Printf("Error parsing date for album %s: %v", album.Title, err)
			continue
		}
		result = append(result, AlbumInfo{
			Title: album.Title,
			Day:   date.Day(),
			Month: int(date.Month()),
			Year:  date.Year(),
		})
	}
	return result
}

// generateTest creates a new test and returns its ID and description
func generateTest(artistName string) (string, string, error) {
	mu.Lock()
	defer mu.Unlock()

	if len(descriptions) == 0 {
		return "", "", fmt.Errorf("no descriptions available")
	}

	testID := fmt.Sprintf("test-%d", rand.Intn(100000))
	testDescription := getRandomDescription()
	tests[testID] = Test{
		Description: testDescription,
		ArtistName:  artistName,
	}

	return testID, testDescription, nil
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

// albumSongsHandler handles the /album-songs endpoint
func albumSongsHandler(w http.ResponseWriter, r *http.Request) {
	artistName := r.URL.Query().Get("artistName")
	albumName := r.URL.Query().Get("albumName")

	if strings.TrimSpace(artistName) == "" || strings.TrimSpace(albumName) == "" {
		http.Error(w, "Missing artistName or albumName query parameter", http.StatusBadRequest)
		return
	}

	albums, err := getAlbumsByArtist(artistName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching albums: %v", err), http.StatusInternalServerError)
		return
	}

	targetAlbum, found := findAlbum(albums, albumName)
	if !found {
		http.Error(w, "Album not found", http.StatusNotFound)
		return
	}

	tracks, err := artists.GetAlbumSongs(targetAlbum)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching album songs: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, tracks)
}

// findAlbum searches for an album by name in a slice of albums
func findAlbum(albums []artists.ReleaseGroup, albumName string) (artists.ReleaseGroup, bool) {
	for _, album := range albums {
		if strings.EqualFold(album.Title, albumName) {
			return album, true
		}
	}
	return artists.ReleaseGroup{}, false
}

// loadDescriptions loads test descriptions from a JSON file
func loadDescriptions(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading descriptions file: %w", err)
	}

	err = json.Unmarshal(data, &descriptions)
	if err != nil {
		return fmt.Errorf("error unmarshaling descriptions: %w", err)
	}

	if len(descriptions) == 0 {
		return fmt.Errorf("no descriptions found in file")
	}

	return nil
}

// getRandomDescription returns a random description from the loaded descriptions
func getRandomDescription() string {
	return descriptions[rand.Intn(len(descriptions))]
}

// verifyTest checks if the user passed the test
func verifyTest(testID string, imageBase64 string) (bool, error) {
	// return true, nil
	mu.Lock()
	defer mu.Unlock()
	test := tests[testID]
	description := test.Description

	answer, err := llm.AnalyzeImage(imageBase64, description)
	if err != nil {
		return false, nil
	}

	var data JSONData
	err = jsonextractor.UnmarshalJSONData(answer, &data)
	if err != nil {
		return false, nil
	}

	passed := data.Score > 2
	return passed, nil
}

// albumSongsWithTestHandler handles the new endpoint for album songs with test verification
func albumSongsWithTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		AlbumName   string `json:"albumName"`
		ImageBase64 string `json:"imageBase64"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// log.Printf("Received request: %+v", request)

	if request.AlbumName == "" {
		http.Error(w, "Album name is required", http.StatusBadRequest)
		return
	}

	if request.ImageBase64 == "" {
		http.Error(w, "Image data is required", http.StatusBadRequest)
		return
	}

	testID, err := r.Cookie("testID")
	if err != nil {
		http.Error(w, "Test ID not found", http.StatusBadRequest)
		return
	}

	passed, err := verifyTest(testID.Value, request.ImageBase64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error verifying test: %v", err), http.StatusInternalServerError)
		return
	}

	if !passed {
		http.Error(w, "Test failed. Please try again.", http.StatusUnauthorized)
		return
	}

	mu.RLock()
	test, exists := tests[testID.Value]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}

	albums, err := getAlbumsByArtist(test.ArtistName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching albums: %v", err), http.StatusInternalServerError)
		return
	}

	targetAlbum, found := findAlbum(albums, request.AlbumName)
	if !found {
		http.Error(w, "Album not found", http.StatusNotFound)
		return
	}

	tracks, err := artists.GetAlbumSongs(targetAlbum)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching album songs: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, tracks)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend.html")
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Instead of using "*", we'll echo the Origin header if present
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback to localhost if no Origin header
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {
	if err := loadDescriptions("descriptions.json"); err != nil {
		log.Fatalf("Error loading descriptions: %v", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/albums", enableCORS(albumsHandler))
	http.HandleFunc("/album-songs", enableCORS(albumSongsHandler))
	http.HandleFunc("/album-songs-with-test", enableCORS(albumSongsWithTestHandler))

	port := ":8080"
	log.Printf("Server is running on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
