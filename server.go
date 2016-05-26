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

type TTLMapEntry struct {
     Creation time.Time
     Value interface{}
}

// check if an entry has expired given a TTL
func (e *TTLMapEntry) expired(ttl time.Duration) bool {
	if time.Since(e.Creation) > ttl {
		return true
	}
	return false
}

type TTLMap struct {
     mu sync.RWMutex
     ttl  time.Duration
     items    map[string]*TTLMapEntry
}

func NewTTLMap(ttl time.Duration) *TTLMap {
	return &TTLMap{
		items: make(map[string]*TTLMapEntry),
		ttl:  ttl,
	}
}

// add/override an entry
func (m *TTLMap) Set(k string, v interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[k] = &TTLMapEntry{
		Creation: time.Now(),
		Value:      v,
	}
}

// retrieve an entry
func (m *TTLMap) Get(k string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.items[k]
	if !ok || v.expired(m.ttl) {
		return nil, false
	}
	return v.Value, true
}

// delete expired entries
func (m *TTLMap) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.items {
		if v.expired(m.ttl) {
			delete(m.items, k)
		}
	}
}

// starts a Ticker to cleanup db regulary
func startCleanupTask(m *TTLMap, t time.Duration) chan<- struct{} {
	ticker := time.NewTicker(t)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				m.cleanup()
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


type Rig struct {
     Coinbase string
}


// creates the HTTP handler
func cubiHandler(rigs *TTLMap, installFile string, idLen int) http.Handler {
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
			rig, ok := rigs.Get(id)
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
				if _, ok := rigs.Get(id); !ok {
					break
				}
			}

			rigs.Set(id, &rig)

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

	m := NewTTLMap(15 * time.Minute)

	http.Handle("/", cubiHandler(m, "install_rig.sh", 3))

	stop := startCleanupTask(m, time.Minute)

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)

	close(stop)
}
