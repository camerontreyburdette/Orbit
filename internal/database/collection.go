package database

func ClampIndex(index int, length int) int {
	if index < 0 {
		return 0
	}
	if index > length {
		return length
	}
	return index
}

func IndexOfIdentifier[Item Identifiable](items []Item, identifier int64) int {
	for index, item := range items {
		if item.EntityIdentifier() == identifier {
			return index
		}
	}
	return -1
}

func RemoveAtIndex[Item any](items []Item, index int) []Item {
	if index < 0 || index >= len(items) {
		return items
	}
	return append(items[:index], items[index+1:]...)
}

func RemoveByIdentifier[Item Identifiable](items []Item, identifier int64) []Item {
	return RemoveAtIndex(items, IndexOfIdentifier(items, identifier))
}

func InsertAtIndex[Item any](items []Item, index int, item Item) []Item {
	index = ClampIndex(index, len(items))
	items = append(items, item)
	copy(items[index+1:], items[index:])
	items[index] = item
	return items
}

func InsertAfterIdentifier[Item Identifiable](items []Item, anchorIdentifier int64, item Item) []Item {
	anchorIndex := IndexOfIdentifier(items, anchorIdentifier)
	if anchorIndex < 0 {
		return append(items, item)
	}
	return InsertAtIndex(items, anchorIndex+1, item)
}

func MoveToIndex[Item Identifiable](items []Item, item Item, newIndex int) []Item {
	items = RemoveByIdentifier(items, item.EntityIdentifier())
	return InsertAtIndex(items, newIndex, item)
}

func InsertAllAtIndex[Item any](items []Item, index int, inserted []Item) []Item {
	index = ClampIndex(index, len(items))
	return append(items[:index], append(inserted, items[index:]...)...)
}

func PartitionByIdentifiers[Item Identifiable](items []Item, identifiers map[int64]struct{}) ([]Item, []Item) {
	selected := make([]Item, 0, len(identifiers))
	remaining := make([]Item, 0, len(items))
	for _, item := range items {
		if _, isSelected := identifiers[item.EntityIdentifier()]; isSelected {
			selected = append(selected, item)
		} else {
			remaining = append(remaining, item)
		}
	}
	return selected, remaining
}

func IdentifierSet(identifiers []int64) map[int64]struct{} {
	set := make(map[int64]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		set[identifier] = struct{}{}
	}
	return set
}
