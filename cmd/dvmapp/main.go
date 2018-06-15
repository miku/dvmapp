//go:generate statik -src=../../public
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

	_ "github.com/miku/dvmapp/cmd/dvmapp/statik"
	"github.com/rakyll/statik/fs"
)

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
	if puzzle.Artifacts, err = dirFilenames("media/images/a"); err != nil {
		return nil, err
	}
	if puzzle.People, err = dirFilenames("media/images/p"); err != nil {
		return nil, err
	}
	if puzzle.Landscapes, err = dirFilenames("media/images/l"); err != nil {
		return nil, err
	}
	return puzzle, nil
}

func main() {
	rand.Seed(time.Now().Unix())

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	puzzle, err := NewPuzzle()
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(puzzle)
	log.Println(puzzle.Size())
	log.Println(puzzle.Combinations())

	fs := http.FileServer(http.Dir("media"))
	http.Handle("/media/", http.StripPrefix("/media/", fs))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(statikFS)))

	// /w/{id}  write new story
	// /r       read all
	// /r/{id}  read a specific image story
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.ListenAndServe("0.0.0.0:8080", nil)
}
