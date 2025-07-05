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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Card API is running!")
	})

	http.HandleFunc("/cards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
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

		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cards)

		default:
			http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/cards/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/cards/")

		// Обработка загрузки изображений по пути /cards/{id}/images
		if strings.HasSuffix(path, "/images") {
			idStr := strings.TrimSuffix(path, "/images")
			id, err := strconv.Atoi(idStr)
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

			// Запускаем обработку в горутине
			go func() {
				err := saveImages(id, imgs)
				if err != nil {
					log.Printf("Error saving images for card %d: %v", id, err)
				}
			}()

			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, "Image processing started for card %d\n", id)
			fmt.Println("Current working dir:", filepath.Dir(os.Args[0]))
			return
		}

		// Работа с карточками по ID (удаление, редактирование)
		id, err := strconv.Atoi(path)
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
	})

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// saveImages скачивает, сжимает и сохраняет изображения локально в папку cards/{id}/
func saveImages(cardID int, urls []string) error {
	dir := fmt.Sprintf("cards/%d", cardID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

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

		// Сжимаем и сохраняем как JPEG с качеством 80
		filePath := filepath.Join(dir, fmt.Sprintf("img%d.jpg", i+1))
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
	}

	return nil
}
