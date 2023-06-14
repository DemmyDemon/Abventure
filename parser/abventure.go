package parser

import (
	"fmt"
	"html"
	"io"
	"time"

	"github.com/demmydemon/abventure/hash"
	"github.com/demmydemon/abventure/inventory"
)

type AbventureCell struct {
	Name  string
	Label string `json:",omitempty"`
	Lines []AbventureLine
}

type AbventureLine struct {
	RequireItems []string `json:",omitempty"`
	ForbidItems  []string `json:",omitempty"`
	GiveItem     string   `json:",omitempty"`
	TakeItem     string   `json:",omitempty"`
	LinksTo      string   `json:",omitempty"`
	Text         string   `json:",omitempty"`
}

func (line *AbventureLine) Tick(abv *Abventure, inv *inventory.Inventory) string {

	if !inv.HasAll(line.RequireItems) {
		return "" // One or more missing items
	}
	if inv.HasAny(line.ForbidItems) {
		return "" // One or more required items missing
	}
	if line.GiveItem != "" && !inv.Add(line.GiveItem) {
		return "" // Failed to give the item, so we already have it
	}
	if line.TakeItem != "" && !inv.Remove(line.TakeItem) {
		return "" // Faled to take item, so we didn't have it
	}
	if line.LinksTo != "" {
		link := hash.Single(line.LinksTo)
		targetCell, exists := abv.Cells[link]
		text := ""
		if line.Text != "" {
			text = line.Text
		}
		if !exists {
			if text == "" {
				text = "[[BROKEN LINK]]"
			}
			return fmt.Sprintf("<a class=\"broken\" href=\"./%s%d\">%s</a>", link, inv.GetState(), html.EscapeString(text))
		}
		if text == "" {

			text = targetCell.Name

			if targetCell.Label != "" {
				text = targetCell.Label
			}
		}
		return fmt.Sprintf("<a href=\"./%s%d\">%s</a>", link, inv.GetState(), html.EscapeString(text))
	}
	return line.Text

}

type Abventure struct {
	Title     string
	Inventory *inventory.Inventory
	Cells     map[string]AbventureCell
	ParseTime *time.Time
}

func (abv *Abventure) out(w io.Writer, format string, a ...any) error {
	_, err := w.Write([]byte(fmt.Sprintf(format, a...)))
	return err
}

func (abv *Abventure) TickCell(w io.Writer, cellHash string, inven uint64) error {
	if cellHash == "" {
		cellHash = hash.PrecalcStart
	}
	cell, ok := abv.Cells[cellHash]
	if !ok {
		return abv.out(w, "<h2>No such cell %s</h2>\n", cellHash)
	}

	err := abv.out(w, "\n<!-- cell %s: %q, holding %d -->\n", cell.Name, cell.Label, inven)
	if err != nil {
		return fmt.Errorf("write cell comment: %w", err)
	}

	label := cell.Label
	if label == "" {
		label = cell.Name
	}
	err = abv.out(w, "<article>\n    <h2>%s</h2>\n", label)
	if err != nil {
		return fmt.Errorf("write cell name: %w", err)
	}

	inv := inventory.FromExisting(abv.Inventory)
	inv.SetState(inven)

	// Keep track if the last line was blank, so we don't double up on blank lines.
	wasBlank := false

	for num, ln := range cell.Lines {

		lineText := ln.Tick(abv, inv)

		if lineText == "" {
			if wasBlank {
				continue // Skip *second* blank line
			}
			wasBlank = true
		} else {
			wasBlank = false
		}
		if lineText != "" {
			err = abv.out(w, "    <p>%s</p>\n", lineText)
			if err != nil {
				return fmt.Errorf("write cell line %d: %w", num, err)
			}
		}
	}

	err = abv.out(w, "</article>\n<ul id=\"inventory\">\n")
	if err != nil {
		return fmt.Errorf("write inventory start: %w", err)
	}

	for _, itemDescription := range inv.Contents() {
		err = abv.out(w, "  <li>%s</li>\n", html.EscapeString(itemDescription))
		if err != nil {
			return fmt.Errorf("write inventory item: %w", err)
		}
	}

	err = abv.out(w, "</ul>\n")
	if err != nil {
		return fmt.Errorf("write incentory end: %w", err)
	}
	return nil
}
