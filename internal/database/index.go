package database

type ListLocation struct {
	Board *Board
	List  *List
}

type CardLocation struct {
	Board *Board
	List  *List
	Card  *Card
}

type AttachmentLocation struct {
	Board      *Board
	List       *List
	Card       *Card
	Attachment *Attachment
}

type ChecklistLocation struct {
	Board     *Board
	Card      *Card
	Checklist *Checklist
}

type ChecklistItemLocation struct {
	Board     *Board
	Card      *Card
	Checklist *Checklist
	Item      *ChecklistItem
}

type boardIndex struct {
	lists          map[int64]ListLocation
	cards          map[int64]CardLocation
	attachments    map[int64]AttachmentLocation
	checklists     map[int64]ChecklistLocation
	checklistItems map[int64]ChecklistItemLocation
	maximum        int64
}

func buildBoardIndex(board *Board) *boardIndex {
	index := &boardIndex{
		lists:          make(map[int64]ListLocation, len(board.Lists)),
		cards:          make(map[int64]CardLocation),
		attachments:    make(map[int64]AttachmentLocation),
		checklists:     make(map[int64]ChecklistLocation),
		checklistItems: make(map[int64]ChecklistItemLocation),
		maximum:        board.Identifier,
	}
	for _, list := range board.Lists {
		index.indexList(board, list)
	}
	return index
}

func (index *boardIndex) trackMaximum(identifier int64) {
	if identifier > index.maximum {
		index.maximum = identifier
	}
}

func (index *boardIndex) indexList(board *Board, list *List) {
	index.lists[list.Identifier] = ListLocation{Board: board, List: list}
	index.trackMaximum(list.Identifier)
	for _, card := range list.Cards {
		index.indexCard(board, list, card)
	}
}

func (index *boardIndex) indexCard(board *Board, list *List, card *Card) {
	index.cards[card.Identifier] = CardLocation{Board: board, List: list, Card: card}
	index.trackMaximum(card.Identifier)
	for _, attachment := range card.Attachments {
		index.attachments[attachment.Identifier] = AttachmentLocation{Board: board, List: list, Card: card, Attachment: attachment}
		index.trackMaximum(attachment.Identifier)
	}
	for _, checklist := range card.Checklists {
		index.indexChecklist(board, card, checklist)
	}
}

func (index *boardIndex) indexChecklist(board *Board, card *Card, checklist *Checklist) {
	index.checklists[checklist.Identifier] = ChecklistLocation{Board: board, Card: card, Checklist: checklist}
	index.trackMaximum(checklist.Identifier)
	for _, item := range checklist.Items {
		index.checklistItems[item.Identifier] = ChecklistItemLocation{Board: board, Card: card, Checklist: checklist, Item: item}
		index.trackMaximum(item.Identifier)
	}
}
