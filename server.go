package main

import (
	"bufio"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type RigConfig struct {
	Coinbase string
}

// Returns a list of lines of a text file
func readWords(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}
	return words, scanner.Err()
}

// Sample n word from list (with replaceament)
func sample(list []string, n int) []string {
	m := len(list)
	r := make([]string, n)
	for i := 0; i < n; i++ {
		r[i] = list[rand.Intn(m)]
	}
	return r
}

const baseUrl = "https://ethercubi.lol/"
const idLen = 3

// Create an id for a new rig config
func createId(alphabet []string) string {
	return strings.Join(sample(alphabet, idLen), "-")
}

func cubiHandler() http.Handler {
	db := make(map[string]*RigConfig)
	words, err := readWords("wordlist")
	if err != nil {
		log.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			cid := strings.Trim(r.URL.Path, "/")
			_, ok := db[cid]
			if !ok {
				http.NotFound(w, r)
				return
			}

			// w.Write([]byte(c.Coinbase))
			w.Write([]byte("The Ether must flow !"))
		case "POST":
			var id string
			for {
				id = createId(words)
				if _, ok := db[id]; !ok {
					break
				}
			}

			log.Println(id)
		default:
			http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	http.Handle("/", cubiHandler())

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}
