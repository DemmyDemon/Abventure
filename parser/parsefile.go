package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/demmydemon/abventure/hash"
	"github.com/demmydemon/abventure/inventory"
)

var (
	ReInstructionGlyph = regexp.MustCompile(`^([\:\>\%\?\!\&\@])(\w+)\s*(.*)$`)
	ReInstructionWord  = regexp.MustCompile(`^([\:\>\%\?\!\&\@])([\w-_]+)$`)
	ReComment          = regexp.MustCompile(`#.*$`)
)

func Trim(txt string) string {
	txt = strings.Trim(txt, " \t\n\v\f\r\u0085\u00A0") // TODO: Still missing some very high value ones. See bufio.isSpace()
	txt = ReComment.ReplaceAllString(txt, "")
	return txt
}

type ParserState struct {
	Abventure   Abventure
	currentCell AbventureCell
	currentLine int
	Verbose     bool
}

func NewParserState(verbose bool) ParserState {
	return ParserState{
		Abventure: Abventure{
			Cells:     make(map[string]AbventureCell),
			Inventory: inventory.New(),
		},
		currentCell: AbventureCell{},
		currentLine: 0,
		Verbose:     verbose,
	}
}

func ParseFile(filename string, verbose bool) (Abventure, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Abventure{}, fmt.Errorf("load abventure: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// scanner.Split(bufio.ScanLines) // This is the default behaviour

	state := NewParserState(verbose)

	for scanner.Scan() {
		err := state.ParseLine(scanner.Text())
		state.currentLine++
		if err != nil {
			return state.Abventure, err
		}
	}

	state.CloseCell() // Because we have to close the last cell

	now := time.Now()
	state.Abventure.ParseTime = &now

	return state.Abventure, nil
}

func (state *ParserState) bark(format string, a ...any) {
	if state.Verbose {
		format = fmt.Sprintf("[%s:%s:%d] %s\n", state.Abventure.Title, state.currentCell.Name, state.currentLine, format)
		fmt.Printf(format, a...)
	}
}

func (state *ParserState) ParseLine(line string) error {
	line = Trim(line)

	if line == "" {
		state.bark("Empty line skipped over")
		return nil
	}

	if state.Abventure.Title == "" {
		state.bark("Title found: %s", line)
		state.Abventure.Title = line
		return nil
	}

	cellLine := AbventureLine{}

	words := strings.Split(line, " ")
lineParse:
	for i, word := range words {
		found := ReInstructionWord.FindStringSubmatch(word)
		if found == nil { // Done with instructions, apparently!
			cellLine.Text = Trim(strings.Join(words[i:], " "))
			state.bark("Line text: %q", cellLine.Text)
			break
		}
		switch found[1] {
		case ":": // New cell
			state.CloseCell()
			state.NewCell(found[2])
			if len(words) > i {
				state.currentCell.Label = Trim(strings.Join(words[i+1:], " "))
			}
			state.bark("New cell %s labeled %q", found[2], state.currentCell.Label)
			return nil // Don't save this line
		case ">": // Destination
			state.bark("Destination: %s", found[2])
			cellLine.LinksTo = found[2]
		case "%": // Item definition¨
			description := ""
			if len(words) > i {
				description = Trim(strings.Join(words[i+1:], " "))
			}
			state.bark("Item definition: %s: %q", found[2], description)
			state.Abventure.Inventory.Define(found[2], description)
			return nil // Don't save this line
		case "?": // Item check
			state.bark("Item check: %s", found[2])
			cellLine.RequireItems = append(cellLine.RequireItems, found[2])
		case "!": // Inverted item check
			state.bark("Inverted item check: %s", found[2])
			cellLine.ForbidItems = append(cellLine.ForbidItems, found[2])
		case "&": // Give item
			text := ""
			if len(words) > i {
				text = Trim(strings.Join(words[i+1:], " "))
			}
			state.bark("Give item: %s %q", found[2], text)
			cellLine.GiveItem = found[2]
			cellLine.Text = text
			break lineParse
		case "@": // Take item
			text := ""
			if len(words) > i {
				text = Trim(strings.Join(words[i+1:], " "))
			}
			state.bark("Take item: %s %q", found[2], text)
			cellLine.TakeItem = found[2]
			cellLine.Text = text
			break lineParse
		default:
			return errors.New("unexpected glyph " + found[0])
		}
	}

	state.currentCell.Lines = append(state.currentCell.Lines, cellLine)

	return nil
}

func (state *ParserState) CloseCell() {
	state.bark("Closing active cell")
	if state.currentCell.Name == "" {
		return // Because this isn't a real cell, it's a zero value
	}
	key := hash.Single(state.currentCell.Name)
	state.Abventure.Cells[key] = state.currentCell
}

func (state *ParserState) NewCell(name string) {
	state.bark("Initializing cell: %s", name)
	state.currentCell = AbventureCell{
		Name:  name,
		Lines: []AbventureLine{},
	}
}
