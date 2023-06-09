// Package invetory deals with inventory mangement
package inventory

import (
	"fmt"
	"sort"
)

// Item holds the description and ID of items
type Item struct {
	ID          uint64
	Description string
}

// Inventory holds the inventory state and the item descriptions
type Inventory struct {
	state   uint64
	Items   map[string]Item
	Verbose bool `json:"-"`
}

// New creates an empty inventory with no items described
func New() *Inventory {
	empty := make(map[string]Item)
	return &Inventory{
		state: 0,
		Items: empty,
	}
}

// FromExisting creates an empty inventory containing the item definition from the inventory passed to it.
func FromExisting(inv *Inventory) *Inventory {
	return &Inventory{
		state:   0,
		Items:   inv.Items,
		Verbose: inv.Verbose,
	}
}

func (inv *Inventory) bark(format string, a ...any) {
	if inv.Verbose {
		fmt.Printf(format, a...)
	}
}

// Define stores an item description under the given name, making up an ID for it.
// If the item is already defined, it just updates the description.
func (inv *Inventory) Define(name string, desciption string) {

	// If it already exists, we just update the description.
	if old, exists := inv.Items[name]; exists {
		old.Description = desciption
		inv.Items[name] = old
		inv.bark("%s updated\n", name)
		return
	}

	slot := len(inv.Items)
	id := uint64(1 << slot)
	item := Item{
		ID:          id,
		Description: desciption,
	}
	inv.Items[name] = item
	inv.bark("%s defined (%02d)\n", name, id)
}

func (inv *Inventory) Describe(name string) string {
	item, exist := inv.Lookup(name)
	if exist {
		return item.Description
	}
	inv.bark("Can't describe %s: No such item!\n", name)
	return ""
}

// Contents returns descriptions of all the items currently held in this inventory, in the order they were first defined.
func (inv *Inventory) Contents() []string {

	// First we get just the items we actually have in the inventory
	items := make([]Item, 0, len(inv.Items))
	for _, item := range inv.Items {
		if inv.HasItem(item) {
			items = append(items, item)
		}
	}

	// Then we sort them by ID
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	// And finally collect their descriptions
	descriptions := make([]string, 0, len(items))
	for _, item := range items {
		if item.Description != "" {
			descriptions = append(descriptions, item.Description)
		}
	}

	return descriptions
}

// DebugTable dumps the current state of the inventory to STDOUT for debugging purposes.
func (inv *Inventory) DebugTable() {
	fmt.Println(" Has | ID | Name       | Description")
	fmt.Println("-----+----+------------+------------")
	for name, item := range inv.Items {
		has := " "
		if inv.Has(item.ID) {
			has = "*"
		}
		fmt.Printf("  %s  | %02d | %-10s | %s\n", has, item.ID, name, item.Description)
	}
	fmt.Println("-----+----+------------+------------")
	fmt.Printf("State: %d\n", inv.state)
}

// Lookup takes an item name and returns the Item struct for it, and a bool indicating if it exists or not.
func (inv *Inventory) Lookup(name string) (Item, bool) {
	item, ok := inv.Items[name]
	inv.bark("Lookup of %s: %v\n", name, ok)
	return item, ok
}

// SetState sets the inventory state, doing no checks for validity what so ever.
func (inv *Inventory) SetState(state uint64) {
	inv.state = state
}

// GetState returns the current state of what is in the inventory.
func (inv *Inventory) GetState() uint64 {
	return inv.state
}

// Has returns if the inventory state contains the given item ID
func (inv *Inventory) Has(itemID uint64) bool {
	return inv.state&itemID != 0
}

// HasItem returns if the inventory state contains the given Item
func (inv *Inventory) HasItem(item Item) bool {
	return inv.state&item.ID != 0
}

// HasAny returns true if any of the given names matches a held item.
func (inv *Inventory) HasAny(names []string) bool {

	// If no items are specified, by definition we don't have any of them.
	if len(names) == 0 {
		return false
	}

	for _, candidate := range names {
		item, exist := inv.Lookup(candidate)
		if !exist {
			// If it's a non-existant item, we can't possibly have it in our inventory!
			continue
		}
		if inv.HasItem(item) {
			return true
		}
	}
	return false
}

// HasAll returns true if all of the given items match a held item.
func (inv *Inventory) HasAll(names []string) bool {

	// We always have all of the non-items. An empty set is a set we possess all members of, by definition.
	if len(names) == 0 {
		return true
	}
	for _, candidate := range names {
		item, exist := inv.Lookup(candidate)
		if !exist {
			// We can't possibly have a non-existant item in our inventory!
			return false
		}
		if !inv.HasItem(item) {
			return false
		}
	}
	return true
}

// Add adds the named item to the inventory, returning if the operation was successful.
// Note that "item was misspelled" and "item was already in there" both return false.
func (inv *Inventory) Add(name string) bool {
	item, exist := inv.Lookup(name)
	if !exist {
		inv.bark("Could not add item %q: Not defined!\n", name)
		return false
	}

	if inv.HasItem(item) {
		inv.bark("Could not add item: %q: Already possessed!\n", name)
		return false
	}

	inv.bark("Added item %q\n", name)
	inv.AddItem(item)
	return true
}

// AddItem uncritically adds the given item to the inventory with no checks what so ever.
func (inv *Inventory) AddItem(item Item) {
	inv.state = inv.state | item.ID
}

// Remove removes the named item from the inventory, returning if the operation was successful.
// Note that "item was misspelled" and "item was not in there" both return false.
func (inv *Inventory) Remove(name string) bool {
	item, exist := inv.Lookup(name)
	if !exist {
		inv.bark("Could not remove item %q: Not defined!\n", name)
		return false
	}
	if inv.HasItem(item) {
		inv.bark("Removed item %q\n", name)
		inv.RemoveItem(item)
		return true
	}
	inv.bark("Could not remove item %q: Not possessed!\n", name)
	return false
}

// RemoveItem removes the given item from the inventory with no checks what so ever.
func (inv *Inventory) RemoveItem(item Item) {
	inv.state = inv.state &^ item.ID
}
