package inventory_test

import (
	"testing"

	"github.com/demmydemon/abventure/inventory"
)

func TestDefine(t *testing.T) {
	name := "Item Name"
	desc := "Descriptive text"
	desc2 := "Other Descriptive Test"
	inv := inventory.New()
	inv.Define(name, desc)
	if inv.Describe(name) != desc {
		t.Errorf("Stored item description mismatch: Expected %q, got %q", desc, inv.Describe(name))
	}
	item, exist := inv.Lookup(name)
	if !exist {
		t.Error("Defined item could not be looked up")
	}
	if item.Description != desc {
		t.Errorf("Defined item did not retain description: Expected %q, got %q", desc, item.Description)
	}
	if item.ID != 1 {
		t.Errorf("Defined item did not get expected ID: Expected %d, got %d", 1, item.ID)
	}

	inv.Define(name, desc2)
	if inv.Describe(name) != desc2 {
		t.Errorf("Second definition description mismatch: Expected %q, got %q", desc2, inv.Describe(name))
	}

	item2, exist := inv.Lookup(name)
	if !exist {
		t.Error("Item went missing during test?!")
	}
	if item == item2 {
		t.Errorf("Unexpected reference retention of previous item definition when redefining")
	}
	if item.ID != item2.ID {
		t.Errorf("Redefined item has wrong ID: Expected %d, got %d", item.ID, item2.ID)
	}
}

func TestInventoryState(t *testing.T) {
	inv := inventory.New()
	inv.Define("One", "One Description")
	inv.Define("Two", "Two Description")
	if inv.GetState() != 0 {
		t.Errorf("Empty inventory returns wrong state: Expected 0, got %d", inv.GetState())
	}

	inv.Add("One")
	if inv.GetState() != 1 {
		t.Errorf("Single item inventory has wrong state: Expected 1, got %d", inv.GetState())
	}

	inv.Add("Two")
	if inv.GetState() != 3 {
		t.Errorf("Two item inventory has wrong state: Expected 1, got %d", inv.GetState())
	}

	inv.Remove("One")
	if inv.GetState() != 2 {
		t.Errorf("Inventory state wrong after removing One item: Expected 2, got %d", inv.GetState())
	}
}
