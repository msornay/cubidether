package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Rig struct {
	Coinbase string
}

// Rig configuration are stored along a timestamp so they expire after some
// time
type RigEntry struct {
	Rig      *Rig
	Creation time.Time
}

const expirationMinutes = 15

func (r *RigEntry) expired() bool {
	if time.Since(r.Creation) > expirationMinutes*time.Minute {
		return true
	}
	return false
}

type RigDb struct {
	rw   sync.RWMutex
	rigs map[string]*RigEntry
}

func NewRigDb() *RigDb {
	return &RigDb{
		rigs: make(map[string]*RigEntry),
	}
}

func (db *RigDb) Set(id string, r *Rig) {
	db.rw.Lock()
	defer db.rw.Unlock()

	db.rigs[id] = &RigEntry{
		Rig:      r,
		Creation: time.Now(),
	}
}

func (db *RigDb) Get(id string) (*Rig, bool) {
	db.rw.RLock()
	defer db.rw.RUnlock()
	v, ok := db.rigs[id]
	if !ok || v.expired() {
		return nil, false
	}
	return v.Rig, true
}

func (db *RigDb) cleanup() {
	db.rw.Lock()
	defer db.rw.Unlock()

	for k, v := range db.rigs {
		if v.expired() {
			delete(db.rigs, k)
		}
	}
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

// Sample n strings from population
func sample(population []string, n int) ([]string, error) {
	m := len(population)
	if m < n {
		return nil, errors.New("sample: size larger than population")
	}

	result := make([]string, n)
	set := make(map[int]struct{}) // to remember what wa spicked

	for i := 0; i < n; i++ {
		var j int
		for {
			j = rand.Intn(m)
			if _, ok := set[j]; !ok {
				break
			}
		}
		set[j] = struct{}{}
		result[i] = population[j]
	}

	return result, nil
}

const baseUrl = "https://ethercubi.lol/"
const idLen = 3

// Create an id for a new rig config
func createId(words []string) string {
	ws, err := sample(words, idLen)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Join(ws, "-")
}

var addrRegexp = regexp.MustCompile("0x[0123456789abcdefABCDEF]{40}")

// Check if addr is a valid ethereum address
func validAddress(addr string) bool {
	return addrRegexp.MatchString(addr)
}

func cubiHandler() http.Handler {
	db := NewRigDb()

	words, err := readWords("wordlist")
	if err != nil {
		log.Fatal(err)
	}

	installTemplate, err := template.ParseFiles("install_rig.sh")
	if err != nil {
		log.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			id := strings.Trim(r.URL.Path, "/")
			rig, ok := db.Get(id)
			if !ok {
				http.NotFound(w, r)
				return
			}

			installTemplate.Execute(w, rig)

			w.Write([]byte("The Ether must flow !"))

		case "POST":
			if r.Body == nil {
				http.Error(w, "empty request body", http.StatusBadRequest)
				return
			}

			decoder := json.NewDecoder(r.Body)
			var rig Rig
			if err := decoder.Decode(&rig); err != nil && err != io.EOF {
				http.Error(w, "cannot decode request body", http.StatusBadRequest)
				return
			}

			if !validAddress(rig.Coinbase) {
				http.Error(w, "invalid coinbase address", http.StatusBadRequest)
				return
			}

			var id string
			for {
				id = createId(words)
				if _, ok := db.Get(id); !ok {
					break
				}
			}

			// Reply a 201 Created with the ressource id in JSON
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(struct{ RigId string }{id})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	http.Handle("/", cubiHandler())

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}
