package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Card struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Position    int      `json:"position"`
	Images      []string `json:"images"`
}

var cards []Card
var nextID = 1

func main() {
	// Раздача статики
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Раздача изображений через /images/
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./static/cards"))))

	// Обработка API
	http.HandleFunc("/cards/", cardsHandler)
	http.HandleFunc("/cards", cardsHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func cardsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/cards")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cards)

		case http.MethodPost:
			var newCard Card
			err := json.NewDecoder(r.Body).Decode(&newCard)
			if err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}

			newCard.ID = nextID
			nextID++
			cards = append(cards, newCard)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(newCard)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	path = strings.TrimPrefix(path, "/")
	if strings.HasSuffix(path, "images") {
		idStr := strings.TrimSuffix(path, "/images")
		id, err := strconv.Atoi(strings.TrimSuffix(idStr, "/"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		var imgs []string
		err = json.NewDecoder(r.Body).Decode(&imgs)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		go func() {
			err := saveImages(id, imgs)
			if err != nil {
				log.Printf("Error saving images for card %d: %v", id, err)
			}
		}()

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Image processing started for card %d\n", id)
		return
	}

	idStr := strings.TrimSuffix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		found := false
		for i, card := range cards {
			if card.ID == id {
				cards = append(cards[:i], cards[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "Card with ID %d deleted", id)

	case http.MethodPut:
		var found *Card
		for i := range cards {
			if cards[i].ID == id {
				found = &cards[i]
				break
			}
		}
		if found == nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}

		var updateData struct {
			Position int `json:"position"`
		}
		err := json.NewDecoder(r.Body).Decode(&updateData)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		found.Position = updateData.Position
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(found)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func saveImages(cardID int, urls []string) error {
	dir := fmt.Sprintf("static/cards/%d", cardID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	savedPaths := []string{}

	for i, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Failed to download image %s: %v", url, err)
			continue
		}
		defer resp.Body.Close()

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Printf("Failed to decode image %s: %v", url, err)
			continue
		}

		fileName := fmt.Sprintf("img%d.jpg", i+1)
		filePath := filepath.Join(dir, fileName)

		outFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Failed to create file %s: %v", filePath, err)
			continue
		}

		opts := jpeg.Options{Quality: 80}
		err = jpeg.Encode(outFile, img, &opts)
		outFile.Close()
		if err != nil {
			log.Printf("Failed to encode and save image %s: %v", filePath, err)
			continue
		}

		log.Printf("Saved image %s", filePath)

		savedPaths = append(savedPaths, fmt.Sprintf("/images/%d/%s", cardID, fileName))
	}

	for i := range cards {
		if cards[i].ID == cardID {
			cards[i].Images = savedPaths
			break
		}
	}

	return nil
}
