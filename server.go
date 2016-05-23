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

// rig configuration are stored along a timestamp so they expire after some
// time
type RigEntry struct {
	Rig      *Rig
	Creation time.Time
}

// check if an entry has expired given a TTL
func (r *RigEntry) expired(ttl time.Duration) bool {
	if time.Since(r.Creation) > ttl {
		return true
	}
	return false
}

// in-memory map of the rigs configurations
type RigDb struct {
	sync.RWMutex
	rigs map[string]*RigEntry
	ttl  time.Duration
}

// Create a new db where entries will expire after TTL
func NewRigDb(ttl time.Duration) *RigDb {
	return &RigDb{
		rigs: make(map[string]*RigEntry),
		ttl:  ttl,
	}
}

// add/override an entry
func (db *RigDb) Set(id string, r *Rig) {
	db.Lock()
	defer db.Unlock()

	db.rigs[id] = &RigEntry{
		Rig:      r,
		Creation: time.Now(),
	}
}

// retrieve an entry
func (db *RigDb) Get(id string) (*Rig, bool) {
	db.RLock()
	defer db.RUnlock()
	v, ok := db.rigs[id]
	if !ok || v.expired(db.ttl) {
		return nil, false
	}
	return v.Rig, true
}

// delete expired entries
func (db *RigDb) cleanup() {
	db.Lock()
	defer db.Unlock()

	for k, v := range db.rigs {
		if v.expired(db.ttl) {
			delete(db.rigs, k)
		}
	}
}

// starts a Ticker to cleanup db regulary
func startCleanupTask(db *RigDb, t time.Duration) chan<- struct{} {
	ticker := time.NewTicker(t)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				db.cleanup()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

// returns a list of lines of a text file
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

// sample n strings from population
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

// create an id for a new rig config
func createId(words []string, n int) string {
	ws, err := sample(words, n)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Join(ws, "-")
}

var addrRegexp = regexp.MustCompile("0x[0123456789abcdefABCDEF]{40}")

// check if addr is a valid ethereum address
func validAddress(addr string) bool {
	return addrRegexp.MatchString(addr)
}

// creates the HTTP handler
func cubiHandler(db *RigDb, installFile string, idLen int) http.Handler {
	words, err := readWords("wordlist")
	if err != nil {
		log.Fatal(err)
	}

	installTemplate, err := template.ParseFiles(installFile)
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
				id = createId(words, idLen)
				if _, ok := db.Get(id); !ok {
					break
				}
			}

			db.Set(id, &rig)

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

	db := NewRigDb(15 * time.Minute)

	http.Handle("/", cubiHandler(db, "install_rig.sh", 3))

	stop := startCleanupTask(db, time.Minute)

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)

	close(stop)
}
