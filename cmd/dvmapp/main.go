package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	humanize "github.com/dustin/go-humanize"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"image/color"
	_ "image/jpeg"

	_ "github.com/mattn/go-sqlite3"
)

var (
	puzzle *Puzzle

	mu sync.Mutex
	db *sql.DB
)

var (
	listen  = flag.String("listen", "0.0.0.0:8080", "hostport to listen on")
	dbpath  = flag.String("db", "data.db", "path to database")
	logfile = flag.String("log", "", "path to logfile")
)

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

// Story is a minimal text.
type Story struct {
	Identifier      int
	ImageIdentifier string
	Text            string
	Created         time.Time
}

// RandomImageWithStory returns the identifier if an image, that has a story
// associated with it.
func (p *Puzzle) RandomImageWithStory() (string, error) {
	rows, err := db.Query("SELECT imageid FROM text")
	if err != nil {
		return "", fmt.Errorf("sql failed: %s", err)
	}

	var imageIdentifiers []string
	var iid string

	for rows.Next() {
		err = rows.Scan(&iid)
		if err != nil {
			return "", fmt.Errorf("sql failed: %s", err)
		}
		imageIdentifiers = append(imageIdentifiers, iid)
	}

	if len(imageIdentifiers) == 0 {
		return p.RandomIdentifier(), nil
	}
	return imageIdentifiers[rand.Intn(len(imageIdentifiers)-1)], nil
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
	return fmt.Sprintf("%s%s%s", as[rand.Intn(len(as)-1)], ps[rand.Intn(len(ps)-1)], ls[rand.Intn(len(ls)-1)])
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

// AboutHandler render info.
func AboutHandler(w http.ResponseWriter, r *http.Request) {

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"ago":   humanize.Time,
		"clip": func(s string) string {
			if len(s) > 50 {
				return fmt.Sprintf("%s ...", s[:50])
			}
			return s
		},
	}

	t, err := template.New("about.html").Funcs(funcMap).ParseFiles("templates/about.html")
	if t == nil {
		log.Printf("template is nil: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var data = struct {
		RandomIdentifier      string
		RandomVideoIdentifier string
	}{
		RandomIdentifier:      puzzle.RandomIdentifier(),
		RandomVideoIdentifier: puzzle.RandomVideoIdentifier(),
	}
	// Cache an image, as poster for certain devices.
	if err := puzzle.CombineImages(data.RandomVideoIdentifier); err != nil {
		log.Printf("cannot combine images for %s: %s", data.RandomVideoIdentifier, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// HomeHandler render homepage.
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"ago":   humanize.Time,
		"clip": func(s string) string {
			if len(s) > 50 {
				return fmt.Sprintf("%s ...", s[:50])
			}
			return s
		},
	}

	t, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if t == nil {
		log.Printf("template is nil: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, imageid, text, created FROM text ORDER BY created ASC LIMIT 100")
	if err != nil {
		log.Printf("sql failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var stories []Story

	var id int
	var imageIdentifier string
	var text string
	var created time.Time

	for rows.Next() {
		err = rows.Scan(&id, &imageIdentifier, &text, &created)
		if err != nil {
			log.Printf("sql failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stories = append(stories, Story{
			Identifier:      id,
			ImageIdentifier: imageIdentifier,
			Text:            text,
			Created:         created,
		})
	}

	riws, err := puzzle.RandomImageWithStory()
	if err != nil {
		log.Printf("sql failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var data = struct {
		RandomIdentifier      string
		RandomImageWithStory  string
		RandomVideoIdentifier string
		Stories               []Story
	}{
		RandomIdentifier:      puzzle.RandomIdentifier(),
		RandomVideoIdentifier: puzzle.RandomVideoIdentifier(),
		Stories:               stories,
		RandomImageWithStory:  riws,
	}
	// Cache an image, as poster for certain devices.
	if err := puzzle.CombineImages(data.RandomVideoIdentifier); err != nil {
		log.Printf("cannot combine images for %s: %s", data.RandomVideoIdentifier, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ReadHandler renders a page with stories to read.
func ReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	rid := vars["rid"]

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"ago":   humanize.Time,
		"clip": func(s string) string {
			if len(s) > 50 {
				return fmt.Sprintf("%s ...", s[:50])
			}
			return s
		},
	}

	t, err := template.New("read.html").Funcs(funcMap).ParseFiles("templates/read.html")
	if t == nil {
		log.Printf("template failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	it, err := puzzle.ResolveImages(rid)
	if err != nil {
		log.Printf("cannot resolve images for identifier %s: %s", rid, err)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := puzzle.CombineImages(rid); err != nil {
		log.Printf("cannot combine images for %s: %s", rid, err)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rows, err := db.Query("SELECT id, text, created FROM text where imageid = ? ORDER BY created DESC", rid)
	if err != nil {
		log.Printf("sql failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var stories []Story

	var identifier int
	var text string
	var created time.Time

	for rows.Next() {
		err = rows.Scan(&identifier, &text, &created)
		if err != nil {
			log.Printf("sql failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stories = append(stories, Story{
			Identifier: identifier,
			Text:       text,
			Created:    created,
		})
	}

	var data = struct {
		RandomIdentifier string
		ImageTriplet     *ImageTriplet
		Stories          []Story
	}{
		RandomIdentifier: rid,
		ImageTriplet:     it,
		Stories:          stories,
	}
	if err := t.Execute(w, data); err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// StoryHandler renders a single stories to read.
func StoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	identifierString := vars["id"]
	identifier, err := strconv.Atoi(identifierString)
	if err != nil {
		log.Printf("cannot resolve story for identifier %s", identifierString)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"ago":   humanize.Time,
		"clip": func(s string) string {
			if len(s) > 50 {
				return fmt.Sprintf("%s ...", s[:50])
			}
			return s
		},
	}

	t, err := template.New("story.html").Funcs(funcMap).ParseFiles("templates/story.html")
	if t == nil {
		log.Printf("template failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT imageid, text, created FROM text where id = ? LIMIT 1", identifier)
	if err != nil {
		log.Printf("sql failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var stories []Story

	var imageIdentifier string
	var text string
	var created time.Time

	for rows.Next() {
		err = rows.Scan(&imageIdentifier, &text, &created)
		if err != nil {
			log.Printf("sql failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stories = append(stories, Story{
			Identifier:      identifier,
			ImageIdentifier: imageIdentifier,
			Text:            text,
			Created:         created,
		})
	}

	if len(stories) == 0 {
		log.Printf("cannot resolve story for identifier %s", identifierString)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	it, err := puzzle.ResolveImages(imageIdentifier)
	if err != nil {
		log.Printf("cannot resolve images for identifier %s: %s", imageIdentifier, err)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := puzzle.CombineImages(imageIdentifier); err != nil {
		log.Printf("cannot combine images for %s: %s", imageIdentifier, err)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data = struct {
		RandomIdentifier string
		ImageTriplet     *ImageTriplet
		Story            Story
	}{
		RandomIdentifier: imageIdentifier,
		ImageTriplet:     it,
		Story:            stories[0],
	}
	if err := t.Execute(w, data); err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// WriteHandler renders a page to write a story.
func WriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rid := vars["rid"]
	log.Printf("write: %v", rid)

	if r.Method == "POST" {
		mu.Lock()
		defer mu.Unlock()

		r.ParseForm()       // Parse url parameters passed, then parse the response packet for the POST body (request body) attention: If you do not call ParseForm method, the following data can not be obtained form
		fmt.Println(r.Form) // print information on server side.

		story := strings.TrimSpace(r.Form.Get("story"))

		if len(story) > 20000 {
			story = story[:20000]
		}

		time.Sleep(2 * time.Second)

		if story == "" {
			log.Println("no content")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Cleanup.
		// r.Form["language"]

		// Store: story, date, ip, language (ger, niedersorb, obersorb,)

		// `id` INTEGER PRIMARY KEY AUTOINCREMENT,
		// `imageid` TEXT NOT NULL,
		// `text` TEXT NOT NULL,
		// `language` TEXT NOT NULL,
		// `ip` TEXT NOT NULL,
		// `flagged` INTEGER NOT NULL,
		// `created` DATE DEFAULT (datetime('now', 'localtime'))

		stmt, err := db.Prepare(`INSERT INTO text (imageid, text, language, ip, flagged) values (?, ?, ?, ?, ?)`)
		if err != nil {
			log.Printf("prepared statement failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var lang = "ger"
		language := r.Form.Get("language")
		log.Println(language)

		result, err := stmt.Exec(rid, story, lang, r.RemoteAddr, 0)
		if err != nil {
			log.Printf("sql failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		lid, err := result.LastInsertId()
		if err != nil {
			log.Printf("sql last id failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("inserted: %d", lid)
		http.Redirect(w, r, fmt.Sprintf("/r/%s", rid), http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("templates/write.html")
	if t == nil {
		log.Printf("template failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	it, err := puzzle.ResolveImages(rid)
	if err != nil {
		log.Printf("cannot resolve images for identifier %s: %s", rid, err)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := puzzle.CombineImages(rid); err != nil {
		log.Printf("cannot combine images for %s: %s", rid, err)
		io.WriteString(w, http.StatusText(http.StatusNotFound))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data = struct {
		RandomIdentifier string
		ImageTriplet     *ImageTriplet
	}{
		RandomIdentifier: rid,
		ImageTriplet:     it,
	}
	if err := t.Execute(w, data); err != nil {
		log.Printf("template err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mittagsfrau.de"+r.RequestURI, http.StatusMovedPermanently)
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().Unix())

	if _, err := os.Stat(*dbpath); os.IsNotExist(err) {
		log.Fatal("create a database first, with: make data.db ")
	}
	var err error

	db, err = sql.Open("sqlite3", *dbpath)
	if err != nil {
		log.Fatal(err)
	}

	var loggingWriter = os.Stdout

	if *logfile != "" {
		file, err := os.OpenFile(*logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		loggingWriter = file
		defer file.Close()
	}

	puzzle, err = NewPuzzle()
	if err != nil {
		log.Printf("puzzle err: %s", err)
	}
	log.Printf("%d images, %d combinations, %d animations",
		puzzle.Size(), puzzle.Combinations(), len(puzzle.Videos))

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/robots.txt", 302)
	})
	r.HandleFunc("/r/{rid}", ReadHandler)
	r.HandleFunc("/w/{rid}", WriteHandler)
	r.HandleFunc("/s/{id}", StoryHandler)
	r.HandleFunc("/about", AboutHandler)
	r.HandleFunc("/", HomeHandler)
	http.Handle("/", r)

	loggedRouter := handlers.LoggingHandler(loggingWriter, r)

	go func() {
		// log.Fatal(http.ListenAndServe(*listen, loggedRouter))
		log.Fatal(http.ListenAndServe(*listen, http.HandlerFunc(redirectHandler)))
	}()
	log.Fatal(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/mittagsfrau.de/fullchain.pem", "/etc/letsencrypt/live/mittagsfrau.de/privkey.pem", loggedRouter))
}
