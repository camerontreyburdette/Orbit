package api

import (
	"fmt"
	"strings"

	"orbit/internal/database"
)

func normalizeChecklistTitle(title string) string {
	return database.TruncateRunes(strings.TrimSpace(title), maximumTitleCharacters)
}

func normalizeChecklistItemText(text string) (string, error) {
	text = database.TruncateRunes(strings.TrimSpace(text), maximumItemCharacters)
	if text == "" {
		return "", fmt.Errorf("Item text cannot be empty")
	}
	return text, nil
}

func (handler *Handler) AddChecklist(cardIdentifier int64, title string) (response, error) {
	title = normalizeChecklistTitle(title)
	if title == "" {
		title = "Checklist"
	}

	board, _, card := handler.store.FindCard(cardIdentifier)
	if card == nil {
		return nil, fmt.Errorf("Card not found")
	}

	checklist := &database.Checklist{
		Identifier: handler.store.NewIdentifier(),
		Title:      title,
		Items:      make([]*database.ChecklistItem, 0),
	}

	card.Checklists = database.InsertAtIndex(card.Checklists, 0, checklist)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}

	return response{"checklist": checklist}, nil
}

func (handler *Handler) RenameChecklist(checklistIdentifier int64, title string) (response, error) {
	title = normalizeChecklistTitle(title)
	if title == "" {
		return nil, fmt.Errorf("Checklist title cannot be empty")
	}

	board, _, checklist := handler.store.FindChecklist(checklistIdentifier)
	if checklist == nil {
		return nil, fmt.Errorf("Checklist not found")
	}

	checklist.Title = title
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) MoveChecklist(checklistIdentifier int64, newIndex int) (response, error) {
	board, card, checklist := handler.store.FindChecklist(checklistIdentifier)
	if checklist == nil {
		return nil, fmt.Errorf("Checklist not found")
	}

	card.Checklists = database.MoveToIndex(card.Checklists, checklist, newIndex)
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) DeleteChecklist(checklistIdentifier int64) (response, error) {
	board, card, checklist := handler.store.FindChecklist(checklistIdentifier)
	if checklist == nil {
		return nil, fmt.Errorf("Checklist not found")
	}

	card.Checklists = database.RemoveByIdentifier(card.Checklists, checklistIdentifier)
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) AddChecklistItem(checklistIdentifier int64, text string) (response, error) {
	text, textError := normalizeChecklistItemText(text)
	if textError != nil {
		return nil, textError
	}

	board, _, checklist := handler.store.FindChecklist(checklistIdentifier)
	if checklist == nil {
		return nil, fmt.Errorf("Checklist not found")
	}

	item := &database.ChecklistItem{
		Identifier: handler.store.NewIdentifier(),
		Text:       text,
	}

	checklist.Items = append(checklist.Items, item)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}

	return response{"item": item}, nil
}

func (handler *Handler) ToggleChecklistItem(itemIdentifier int64, done bool) (response, error) {
	board, _, _, item := handler.store.FindChecklistItem(itemIdentifier)
	if item == nil {
		return nil, fmt.Errorf("Item not found")
	}

	item.Done = done
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) EditChecklistItem(itemIdentifier int64, text string) (response, error) {
	text, textError := normalizeChecklistItemText(text)
	if textError != nil {
		return nil, textError
	}

	board, _, _, item := handler.store.FindChecklistItem(itemIdentifier)
	if item == nil {
		return nil, fmt.Errorf("Item not found")
	}

	item.Text = text
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) DeleteChecklistItem(itemIdentifier int64) (response, error) {
	board, _, checklist, item := handler.store.FindChecklistItem(itemIdentifier)
	if item == nil {
		return nil, fmt.Errorf("Item not found")
	}

	checklist.Items = database.RemoveByIdentifier(checklist.Items, itemIdentifier)
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) MoveChecklistItem(itemIdentifier int64, toChecklistIdentifier int64, newIndex int) (response, error) {
	board, sourceCard, sourceChecklist, item := handler.store.FindChecklistItem(itemIdentifier)
	if item == nil {
		return nil, fmt.Errorf("Item not found")
	}

	_, targetCard, targetChecklist := handler.store.FindChecklist(toChecklistIdentifier)
	if targetChecklist == nil {
		return nil, fmt.Errorf("Checklist not found")
	}

	if targetCard != sourceCard {
		return nil, fmt.Errorf("Items can only move within the same card")
	}

	sourceChecklist.Items = database.RemoveByIdentifier(sourceChecklist.Items, itemIdentifier)
	targetChecklist.Items = database.InsertAtIndex(targetChecklist.Items, newIndex, item)
	return handler.saveBoardOrFail(board)
}
