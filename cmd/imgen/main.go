// Generate combined images.
package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageTriplet groups three images for a puzzle.
type ImageTriplet struct {
	Artifact  string
	People    string
	Landscape string
}

// Puzzle game allows to retrieve a random combination of images.
type Puzzle struct {
	Artifacts  []string
	People     []string
	Landscapes []string
	Videos     []string
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

// ResolveImages returns a list of image paths given an identifier.
func (p *Puzzle) ResolveImages(identifier string) (*ImageTriplet, error) {
	if len(identifier) != 6 {
		return nil, fmt.Errorf("six digit identifier expected")
	}
	// Add exception for a/02, p/02, a/25 - first story.
	if identifier == "020225" {
		return &ImageTriplet{
			Artifact:  fmt.Sprintf("/static/images/a/%s.jpg", identifier[:2]),
			People:    fmt.Sprintf("/static/images/p/%s.jpg", identifier[2:4]),
			Landscape: fmt.Sprintf("/static/images/a/%s.jpg", identifier[4:6]),
		}, nil
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
	resizeHeight := 300

	filename := fmt.Sprintf("./static/cache/%s.jpg", identifier)
	if _, err := os.Stat(filename); err == nil {
		return nil
	}

	it, err := p.ResolveImages(identifier)
	if err != nil {
		return err
	}

	// Resize artifact.
	artifact, err := imaging.Open(path.Join(".", it.Artifact))
	if err != nil {
		return err
	}
	artifact = imaging.Resize(artifact, 0, resizeHeight, imaging.Lanczos)

	// Resize people.
	people, err := imaging.Open(path.Join(".", it.People))
	if err != nil {
		return err
	}
	people = imaging.Resize(people, 0, resizeHeight, imaging.Lanczos)

	// Resize landscape.
	landscape, err := imaging.Open(path.Join(".", it.Landscape))
	if err != nil {
		return err
	}
	landscape = imaging.Resize(landscape, 0, resizeHeight, imaging.Lanczos)

	// Join.
	dst := imaging.New(960, 300, color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, artifact, image.Pt(0, 0))
	dst = imaging.Paste(dst, people, image.Pt(320, 0))
	dst = imaging.Paste(dst, landscape, image.Pt(640, 0))

	// Save.
	if err := os.MkdirAll("./static/cache", 0777); err != nil {
		return err
	}
	err = imaging.Save(dst, filename)
	if err != nil {
		return err
	}
	log.Printf("cached combined image at %s", filename)
	return err
}

func main() {
	puzzle, err := NewPuzzle()
	if err != nil {
		log.Fatal(err)
	}

	// 25024 combinations.
	for i := 0; i < len(puzzle.Artifacts); i++ {
		for j := 0; j < len(puzzle.People); j++ {
			for k := 0; k < len(puzzle.Landscapes); k++ {
				identifier := fmt.Sprintf("%02d%02d%02d", i, j, k)
				if err := puzzle.CombineImages(identifier); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

}
