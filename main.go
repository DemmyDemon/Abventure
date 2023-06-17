package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/demmydemon/abventure/listing"
	"github.com/demmydemon/abventure/parser"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var isNumber = regexp.MustCompile(`^[0-9]+$`)

//go:embed etc/*
var embedded embed.FS

type Listing struct {
	Name      string
	File      string
	Time      time.Time
	Abventure *parser.Abventure
}

func htmlBegin(title string) []byte {
	return []byte(`<!DOCTYPE html>
<html>
	<head>
		<title>` + title + `</title>
		<link rel="stylesheet" href="/etc/style.css">
	</head>
	<body>
`)
}
func htmlEnd() []byte {
	return []byte(`
	</body>
</html>
`)
}

func parseAbventure(w http.ResponseWriter, name string) {
	w.Header().Add("Content-Type", "text/plain")
	abv, err := parser.ParseFile(filepath.Clean("abventures/"+name+".abv"), true)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	abvJSON, err := json.MarshalIndent(abv, "", "  ")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(abvJSON)
}

func onAbventure(w http.ResponseWriter, name string, cell string, stuff uint64, idx *listing.Index) {
	w.Header().Add("Content-Type", "text/html")

	lst, exist := idx.Get(name)
	if !exist {
		_, err := w.Write([]byte("I'm sorry, but the abventure has derailed entirely!"))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	abv, err := lst.GetAbventure()
	if err != nil {
		fmt.Println(err)
		_, err = w.Write([]byte("Something very bad happened while loading your abventure!"))
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	_, err = w.Write(htmlBegin(abv.Title + " - Abventure"))
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Write([]byte("\n" + `<p><a class="back" href="/">&larr; Return to abventure selection</a></p>`))

	//abv.Inventory.Verbose = true

	abv.TickCell(w, cell, stuff)

	_, err = w.Write(htmlEnd())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func dumpFile(w http.ResponseWriter, r *http.Request, filename string) {
	w.Header().Add("content-type", "text/plain")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("[%s] %s: %s\n", r.RemoteAddr, filename, err)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[%s] %s: %s\n", r.RemoteAddr, filename, err)
		return
	}
	w.Write(data)
}

func main() {

	idx := listing.NewIndex("abventures/")

	dumperEnabled := os.Getenv("ABVDUMPER") != ""
	port := os.Getenv("ABVPORT")
	if port == "" {
		port = "8187"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	if dumperEnabled {
		fmt.Println("WARNING: ABV DUMPER IS ENABLED")
		r.Get("/{abventure:[a-zA-Z0-9_-]+}.json", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "abventure")
			parseAbventure(w, name)
		})
		r.Get("/{abventure:[a-zA-Z0-9_-]+}.abv", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "abventure")
			dumpFile(w, r, filepath.Clean("abventures/"+name+".abv"))
		})
	}

	r.Get("/{abventure:[a-zA-Z0-9_-]+}/", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "abventure")
		onAbventure(w, name, "", 0, idx)

	})
	r.Get("/{abventure:[a-zA-Z0-9_-]+}/{cell:[a-f0-9]{8,}}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "abventure")
		rawCell := chi.URLParam(r, "cell")

		cell, stuff := rawCell[:8], rawCell[8:]
		if !isNumber.MatchString(stuff) {
			stuff = "0"
		}

		invState, err := strconv.ParseUint(stuff, 10, 64)
		if err != nil {
			w.Write([]byte(`Something weird about that inventory!`))
			return
		}

		fmt.Printf("[%s] abventure: %s, cell: %s, stuff: %d\n", r.RemoteAddr, name, cell, invState)

		onAbventure(w, name, cell, invState, idx)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		Listings(w, r, idx)
	})

	r.Handle("/etc/*", http.FileServer(http.FS(embedded)))

	fmt.Println("Will listen on port", port)
	panic(http.ListenAndServe(":"+port, r))
}

func Listings(w http.ResponseWriter, r *http.Request, list *listing.Index) error {
	_, err := w.Write(htmlBegin("Have an Abventure!"))
	if err != nil {
		return fmt.Errorf("listing write error: %w", err)
	}

	w.Write([]byte("<h2>Have an Abventure!</h2>"))

	err = list.Write(w)
	if err != nil {
		return fmt.Errorf("listing handler: %w", err)
	}

	_, err = w.Write(htmlEnd())
	if err != nil {
		return fmt.Errorf("listing write error: %w", err)
	}

	return nil
}
