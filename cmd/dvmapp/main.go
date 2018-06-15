package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var puzzle *Puzzle

// Puzzle game allows to retrieve a random combination of images.
type Puzzle struct {
	Artifacts  []string
	People     []string
	Landscapes []string
}

// Size returns the total number of images.
func (p *Puzzle) Size() int {
	return len(p.Artifacts) + len(p.People) + len(p.Landscapes)
}

// Combinations returns the number of possible combinations.
func (p *Puzzle) Combinations() int {
	return len(p.Artifacts) * len(p.People) * len(p.Landscapes)
}

// RandomIdentifier returns a random string pointing to a combination of three images.
func (p *Puzzle) RandomIdentifier() string {
	var as, ps, ls []string
	for _, fn := range p.Artifacts {
		base := path.Base(fn)
		as = append(as, strings.Replace(base, ".jpg", "", -1))
	}
	for _, fn := range p.People {
		base := path.Base(fn)
		ps = append(ps, strings.Replace(base, ".jpg", "", -1))
	}
	for _, fn := range p.Artifacts {
		base := path.Base(fn)
		ls = append(ls, strings.Replace(base, ".jpg", "", -1))
	}
	return fmt.Sprintf("%s%s%s", as[rand.Intn(len(as))], ps[rand.Intn(len(ps))], ls[rand.Intn(len(ls))])
}

// CreateRandomImage creates a random image from three elements and stores it
// under a directory for generated files.
func (p *Puzzle) CreateRandomImage() (string, error) {
	// If exists, just return the file.
	// Combine, resize and stack horizontally.
	// Save.
	return "", nil
}

// dirFilenames returns all filenames in a given directory.
func dirFilenames(dir string) (result []string, err error) {
	var files []os.FileInfo
	files, err = ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, f := range files {
		result = append(result, filepath.Join(dir, f.Name()))
	}
	return
}

// NewPuzzle finds images in places.
func NewPuzzle() (*Puzzle, error) {
	var puzzle = &Puzzle{}
	var err error
	if puzzle.Artifacts, err = dirFilenames("static/images/a"); err != nil {
		return nil, err
	}
	if puzzle.People, err = dirFilenames("static/images/p"); err != nil {
		return nil, err
	}
	if puzzle.Landscapes, err = dirFilenames("static/images/l"); err != nil {
		return nil, err
	}
	return puzzle, nil
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(puzzle.RandomIdentifier())
	t, err := template.ParseFiles("templates/index.html")
	if t == nil {
		log.Fatal("template failed")
	}
	if err != nil {
		log.Fatal(err)
	}
	if err := t.Execute(w, nil); err != nil {
		log.Fatal(err)
	}
}

// ReadHandler renders a page with stories to read.
func ReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v\n", vars["rid"])
}

// WriteHandler renders a page to write a story.
func WriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v\n", vars["rid"])
}

func main() {
	rand.Seed(time.Now().Unix())

	var err error
	puzzle, err = NewPuzzle()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d images, %d combinations", puzzle.Size(), puzzle.Combinations())

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	r.HandleFunc("/r/{rid}", ReadHandler)
	r.HandleFunc("/w/{rid}", WriteHandler)
	r.HandleFunc("/", HomeHandler)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
