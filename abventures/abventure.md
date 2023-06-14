# Abventure file documentation

This is *hypothetical documentation*, as the software it documents has not been written yet. Expect changes.

## Introduction

The Abventure file format, in .abv files, is intended for small text adventures that are easy for humans to write manually.
An attempt is made to keep the format very easy to understand, with very few rules. First of all, this is to aid in the writing of the file, but it is also in an effort to make it easy to implement software to read the file.

## Basic setup

An abventure file's first line is assumed to be the abventure's title, and is treated as a completely literal string, no matter what it contains.

It is *highly recommended* to add all item definitions to the top of the file, though technically they can occur anywhere.  Keep in mind that how the items are stored in an inventory might break if you add items to the middle of the list, so if you need to grow it after people started playing, do so at the end.

The abventure *must* contain a cell named `Start`, that is used when no other cell name is given.

## Instruction glyphs

Lines can contain an arbituary number of spaces at the start. These must be ignored by software.
An instruction is always a full line. There are no in-line instruction, such as a destination being just part of a line.
A line can contain more than one instruction glyph, but they are not considered different instructions. They are taken together.
A line with no instructions glyph is just text intended for the player to read.
Instructions must come at the start of a line. Any instructions encountered later will be considered literal text insteead.
Empty lines should be considered "on purpose" and be displayed, but never more than one empty line at a time.

- `:` → Cell definition
- `>` → Cell destination
- `%` → Item definition
- `?` → Item check
- `!` → Inveted item check
- `&` → Item added
- `@` → Item removed
- `#` → Skip this line

For the item checks, adding more than one to an instruction means all the conditions must be met for the line to be displayed. This means you can check if somone has a sword *and* a shield, but not if they have a sword *or* an axe. To do *or* logic, use two separate lines.

### : → Cell definition

- *Must* contain a single-word canonical name of the cell being defined.
- *May* contain an additional title text to be displayed in place of the name.
- *May not* be preceded by any item checks.

A cell definiton *begins* at the cell definition glyph, and only ends when the next cell definition starts. Any instructions or lines of text between the cell definition and the next cell definition is considered part of this cell. Obviously, the cell definition also ends if the file ends.

Examples:
```
:Start This is where the adventure begins!
:Village3 The village of Summerfell
```

### > → Cell destination

- *Must* contain a valid single-word canonical name of the cell to go to when activated.
- *May* contain a text to be used in place of the destination cell's title.
- *May* be preceded by one or more item checks.

This is an optionally conditional link to another cell.

Examples:
```
>Village3 Go North
?Map >Village4 Go through the woods
```

### % → Item definition

- *Must* contain a valid single-word canonical name for the item.
- *May* contain a text describing the item.
- *May not* be preceded by any item checks.

Keep in mind that items don't have to be actual, physical items. They can just as easily be emotional states, titles or accolades. It is recommended that the descriptive text is a sentence that makes sense on it's own, for example if your inventory is displayed as a list.

**IMPORTANT:** The way an inventory is stored, using a single UINT64, means that an abventure can define a maximum of 64 items.

Examples:
```
%Map You have a map. It is labeled "spoom" in large, friendly letters.
%Torch You hold a lit torch bright enough to light up your immediate area.
%Bravery You have the heart of a lion. Not a cowardly one, either!
```

### ? → Item check

- *Must* contain a valid single-word canonical item name.
- *Must* be followed by either text to be displayed, or additional instructions.

Examples:
```
?Map You can see this location is marked with a large X on your map. The word "Cellar" is scribbled next to the mark.
?Map ?Torch >Cellar Enter the cellar.
```

### ! → Inverted item check

- *Must* contain a valid single-word canonical item name.
- *Must* be followed by either text to be displayed, or additional instructions.

Examples:
```
!Map !Compass You have no idea where you are.
!Torch It is pitch black.
!Sword You are likely eaten by a Grue.
```

### & → Item added

- *Must* contain a valid single-word canonical item name.
- *May* be followed by text to be displayed.
- *Must not* be followed by any additional instructions.

Note that this is a conditional, meaning that if you already have the item, it is *not* added, and any following text is not displayed.

Examples:
```
&Map You find a map!
&Torch A lit torch is mounted on a wall. You take it down to bring with you.
```

### @ → Item removed

- *Must* contain a valid single-word canonical item name.
- *May* be followed by text to be displayed.
- *Must not* be followed by additional instructions.

Note that this is a conditional, meaning that if you do not have the item, it can't be removed, and any following text is not displayed.

Examples:
```
@Torch Your torch burns out.
@Sword In the darkness, you bump into a table, and drop your sword.
```

### # → Skip this line

- *May* contain whatever you want. Software *must* always ignore lines from the glyph onwards.

This is provided as a way to enter comments into the abventure file that will not affect the displayed abventure in any way.

Examples:
```
# ITEM LIST #
%Wagon You're pulling along a small, red wagon. # This was used in a previous version. Only here to not break inventories.
# Next time, maybe I should develop the abventure a little more before releasing it. Oh well, whatever.
```
