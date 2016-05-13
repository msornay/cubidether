package main

import (
	"bufio"
	"log"
	"net/http"
	"strings"
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
	return lines, scanner.Err()

}

func cubiHandler() http.Handler {

	db := make(map[string]*RigConfig)
	words, err := readWords("wordlist")
	if err != nil()

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
			log.Println(re)
		default:
			http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {
	http.Handle("/", cubiHandler())

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}
