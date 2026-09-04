package api

import "orbit/internal/database"

func (handler *Handler) cloneChecklists(originalChecklists []*database.Checklist) []*database.Checklist {
	clonedChecklists := make([]*database.Checklist, 0, len(originalChecklists))
	for _, originalChecklist := range originalChecklists {
		clonedItems := make([]*database.ChecklistItem, 0, len(originalChecklist.Items))
		for _, originalItem := range originalChecklist.Items {
			clonedItems = append(clonedItems, &database.ChecklistItem{
				Identifier: handler.store.NewIdentifier(),
				Text:       originalItem.Text,
				Done:       originalItem.Done,
			})
		}
		clonedChecklists = append(clonedChecklists, &database.Checklist{
			Identifier: handler.store.NewIdentifier(),
			Title:      originalChecklist.Title,
			Items:      clonedItems,
		})
	}
	return clonedChecklists
}

func (handler *Handler) cloneCard(board *database.Board, originalCard *database.Card, clonedTitle string) *database.Card {
	clonedTags := make([]string, len(originalCard.Tags))
	copy(clonedTags, originalCard.Tags)

	clonedCard := &database.Card{
		Identifier:  handler.store.NewIdentifier(),
		Title:       clonedTitle,
		Description: originalCard.Description,
		Color:       originalCard.Color,
		Tags:        clonedTags,
		CreatedAt:   database.FormatTimestampNow(),
		Checklists:  handler.cloneChecklists(originalCard.Checklists),
		Attachments: make([]*database.Attachment, 0),
	}

	handler.copyCardAttachments(board, originalCard, clonedCard)
	return clonedCard
}

func (handler *Handler) cloneList(board *database.Board, originalList *database.List) *database.List {
	clonedCards := make([]*database.Card, 0, len(originalList.Cards))
	for _, originalCard := range originalList.Cards {
		clonedCards = append(clonedCards, handler.cloneCard(board, originalCard, originalCard.Title))
	}
	return &database.List{
		Identifier: handler.store.NewIdentifier(),
		Name:       originalList.Name + " (Copy)",
		Cards:      clonedCards,
	}
}
