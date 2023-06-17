package listing

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/demmydemon/abventure/parser"
)

type Listing struct {
	FileName  string
	Title     string
	FileTime  time.Time
	Abventure *parser.Abventure
}

func (li *Listing) GetAbventure() (*parser.Abventure, error) {
	if li.Abventure == nil {
		abv, err := parser.ParseFile(li.FileName, false)
		if err != nil {
			return nil, err
		}
		li.Abventure = &abv
	}
	return li.Abventure, nil
}

type Index struct {
	path     string
	mutex    sync.RWMutex
	listings map[string]*Listing
}

func NewIndex(path string) *Index {
	idx := Index{
		path:     path,
		listings: make(map[string]*Listing),
	}
	idx.Refresh()
	return &idx
}

func (idx *Index) Refresh() error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	files, err := os.ReadDir(idx.path)
	if err != nil {
		return fmt.Errorf("reading file list: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue // As we're not being recursive at all!
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".abv") {
			continue // As we only care about Abventure files
		}
		shortName := strings.TrimSuffix(name, ".abv")

		info, err := file.Info()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// Not sure how we'd get here, considering it's working off a file list, but okay.
				delete(idx.listings, shortName)
				continue
			}
			return fmt.Errorf("reading file info for %s: %w", name, err)
		}

		if lst, ok := idx.listings[shortName]; ok {
			if info.ModTime().After(lst.FileTime) {
				lst.Abventure = nil // That is, the file has changed, so should be re-read.
				lst.Title = readFirstLine(lst.FileName)
			}
		} else {
			idx.listings[shortName] = &Listing{
				FileName:  idx.path + file.Name(),
				Title:     readFirstLine(idx.path + file.Name()),
				FileTime:  info.ModTime(),
				Abventure: nil, // This is lazy-loaded later.
			}
		}
	}
	return nil
}

func (idx *Index) Write(w io.Writer) error {
	for shortname, listing := range idx.listings {
		_, err := w.Write([]byte(`<a href="` + shortname + `/">` + listing.Title + "</a><br>\n"))
		if err != nil {
			return fmt.Errorf("listing output: %w", err)
		}
	}
	return nil
}

func (idx *Index) Get(shortName string) (*Listing, bool) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	if lst, ok := idx.listings[shortName]; ok {
		return lst, true
	}
	return &Listing{}, false
}

func readFirstLine(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return err.Error()
	}
	out := []byte{}
	buffer := make([]byte, 1)
	for {
		rd, err := file.Read(buffer)
		if err != nil {
			return err.Error()
		}
		if buffer[0] == '\n' {
			break
		}
		out = append(out, buffer[0:rd]...)
	}
	return string(out)
}
