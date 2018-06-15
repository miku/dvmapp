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

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/gorilla/mux"

	_ "image/jpeg"
)

var puzzle *Puzzle

// Puzzle game allows to retrieve a random combination of images.
type Puzzle struct {
	Artifacts  []string
	People     []string
	Landscapes []string
	Videos     []string
}

// ImageTriplet groups three images for a puzzle.
type ImageTriplet struct {
	Artifact  string
	People    string
	Landscape string
}

// ResolveImages returns a list of image paths given an identifier.
func (p *Puzzle) ResolveImages(identifier string) (*ImageTriplet, error) {
	if len(identifier) != 6 {
		return nil, fmt.Errorf("six digit identifier expected")
	}
	if err := p.CombineImages(identifier); err != nil {
		log.Fatal(err)
	}
	// TODO: Test if identifier is valid.
	return &ImageTriplet{
		Artifact:  fmt.Sprintf("/static/images/a/%s.jpg", identifier[:2]),
		People:    fmt.Sprintf("/static/images/p/%s.jpg", identifier[2:4]),
		Landscape: fmt.Sprintf("/static/images/l/%s.jpg", identifier[4:6]),
	}, nil
}

// CombineImages takes three images and concatenates them into a single one.
func (p *Puzzle) CombineImages(identifier string) error {
	it, err := p.ResolveImages(identifier)
	if err != nil {
		return err
	}
	log.Println(it.Artifact)
	img, err := imgio.Open(it.Artifact)
	if err != nil {
		return err
	}
	w, h := img.Bounds().Max.X, img.Bounds().Max.Y
	ratio := float64(w) / float64(h)
	rh := 300
	rw := int(ratio * float64(rh))
	log.Println(w, h, rh, rw)
	resized := transform.Resize(img, rw, rh, transform.Linear)
	return imgio.Save("xxx.png", resized, imgio.JPEGEncoder(100))
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

// RandomVideoIdentifier returns an existing video identifier.
func (p *Puzzle) RandomVideoIdentifier() string {
	return p.Videos[rand.Intn(len(p.Videos))]
}

// CreateRandomImage creates a random image from three elements and stores it
// under a directory for generated files.
func (p *Puzzle) CreateRandomImage() (string, error) {
	// If exists, just return the file.
	// Combine, resize and stack horizontally.
	// Save.
	return "", nil
}

// readDir returns all filenames in a given directory.
func readDir(dir string) (result []string, err error) {
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
	if puzzle.Artifacts, err = readDir("static/images/a"); err != nil {
		return nil, err
	}
	if puzzle.People, err = readDir("static/images/p"); err != nil {
		return nil, err
	}
	if puzzle.Landscapes, err = readDir("static/images/l"); err != nil {
		return nil, err
	}
	files, err := readDir("static/videos")
	if err != nil {
		return nil, err
	}
	// Store only the ids, like 051314 and so on for a single file type.
	var vids []string
	for _, v := range files {
		if strings.HasSuffix(v, ".webm") {
			base := strings.Replace(path.Base(v), ".webm", "", -1)
			vids = append(vids, strings.Replace(base, "dvm-", "", -1))
		}
	}
	puzzle.Videos = vids
	return puzzle, nil
}

// HomeHandler render homepage.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if t == nil {
		log.Fatal("template failed")
	}
	if err != nil {
		log.Fatal(err)
	}
	var data = struct {
		RandomIdentifier      string
		RandomVideoIdentifier string
	}{
		RandomIdentifier:      puzzle.RandomIdentifier(),
		RandomVideoIdentifier: puzzle.RandomVideoIdentifier(),
	}
	if err := t.Execute(w, data); err != nil {
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
	log.Printf("write: %s", r.Method)

	vars := mux.Vars(r)
	rid := vars["rid"]
	log.Printf("write form for: %v", rid)

	t, err := template.ParseFiles("templates/write.html")
	if t == nil {
		log.Fatalf("template failed: %s", err)
	}
	if err != nil {
		log.Fatal(err)
	}
	it, err := puzzle.ResolveImages(rid)
	if err != nil {
		log.Fatal(err)
	}
	var data = struct {
		RandomIdentifier string
		ImageTriplet     *ImageTriplet
	}{
		RandomIdentifier: rid,
		ImageTriplet:     it,
	}
	if err := t.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	var err error
	puzzle, err = NewPuzzle()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d images, %d combinations, %d animations",
		puzzle.Size(), puzzle.Combinations(), len(puzzle.Videos))

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	r.HandleFunc("/r/{rid}", ReadHandler)
	r.HandleFunc("/w/{rid}", WriteHandler)
	r.HandleFunc("/", HomeHandler)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
