package api

import (
	"fmt"
	"strings"

	"orbit/internal/database"
)

func (handler *Handler) CreateCard(listIdentifier int64, title string) (response, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("Card title cannot be empty")
	}

	board, list := handler.store.FindList(listIdentifier)
	if list == nil {
		return nil, fmt.Errorf("List not found")
	}

	card := &database.Card{
		Identifier:  handler.store.NewIdentifier(),
		Title:       title,
		Tags:        make([]string, 0),
		CreatedAt:   database.FormatTimestampNow(),
		Checklists:  make([]*database.Checklist, 0),
		Attachments: make([]*database.Attachment, 0),
	}

	list.Cards = append(list.Cards, card)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}

	return response{"id": card.Identifier}, nil
}

func (handler *Handler) UpdateCard(cardIdentifier int64, fields map[string]interface{}) (response, error) {
	updates, hasUpdates, parseError := parseCardUpdates(fields)
	if parseError != nil {
		return nil, parseError
	}
	if !hasUpdates {
		return okResponse(), nil
	}

	board, _, card := handler.store.FindCard(cardIdentifier)
	if card == nil {
		return nil, fmt.Errorf("Card not found")
	}

	if applyError := updates.applyTo(card); applyError != nil {
		return nil, applyError
	}
	if updates.title != nil {
		handler.synchronizeCardDirectory(board, card)
	}

	return handler.saveBoardOrFail(board)
}

func (handler *Handler) DeleteCard(cardIdentifier int64) (response, error) {
	board, list, card := handler.store.FindCard(cardIdentifier)
	if card == nil {
		return nil, fmt.Errorf("Card not found")
	}

	list.Cards = database.RemoveByIdentifier(list.Cards, cardIdentifier)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}
	handler.removeCardAttachmentFolders(board, cardAttachmentFolders([]*database.Card{card}))

	return okResponse(), nil
}

func (handler *Handler) MoveCard(cardIdentifier int64, listIdentifier int64, newIndex int) (response, error) {
	sourceBoard, sourceList, card := handler.store.FindCard(cardIdentifier)
	if card == nil {
		return nil, fmt.Errorf("Card not found")
	}

	targetBoard, targetList := handler.store.FindList(listIdentifier)
	if targetList == nil {
		return nil, fmt.Errorf("List not found")
	}

	sourceList.Cards = database.RemoveByIdentifier(sourceList.Cards, cardIdentifier)
	targetList.Cards = database.InsertAtIndex(targetList.Cards, newIndex, card)

	isCrossBoardMove := targetBoard.Identifier != sourceBoard.Identifier
	if isCrossBoardMove {
		handler.moveCardDirectory(sourceBoard, targetBoard, card)
	}

	_ = handler.store.SaveBoard(sourceBoard)
	if isCrossBoardMove {
		_ = handler.store.SaveBoard(targetBoard)
	}

	return okResponse(), nil
}

func (handler *Handler) MoveCards(cardIdentifiers []int64, targetListIdentifier int64, newIndex int) (response, error) {
	if len(cardIdentifiers) == 0 {
		return okResponse(), nil
	}

	targetBoard, targetList := handler.store.FindList(targetListIdentifier)
	if targetList == nil {
		return nil, fmt.Errorf("list not found")
	}

	affectedBoards := map[int64]*database.Board{targetBoard.Identifier: targetBoard}
	cardsToMove := make([]*database.Card, 0, len(cardIdentifiers))

	for _, identifier := range cardIdentifiers {
		sourceBoard, sourceList, card := handler.store.FindCard(identifier)
		if card == nil || sourceList == nil {
			continue
		}
		affectedBoards[sourceBoard.Identifier] = sourceBoard
		cardsToMove = append(cardsToMove, card)
		sourceList.Cards = database.RemoveByIdentifier(sourceList.Cards, identifier)

		if targetBoard.Identifier != sourceBoard.Identifier {
			handler.moveCardDirectory(sourceBoard, targetBoard, card)
		}
	}

	targetList.Cards = database.InsertAllAtIndex(targetList.Cards, newIndex, cardsToMove)
	if saveError := handler.saveBoards(affectedBoards); saveError != nil {
		return nil, saveError
	}

	return okResponse(), nil
}

func (handler *Handler) BatchUpdateCards(cardIdentifiers []int64, fields map[string]interface{}) (response, error) {
	if len(cardIdentifiers) == 0 {
		return okResponse(), nil
	}

	updates := parseBatchCardUpdates(fields)
	affectedBoards := make(map[int64]*database.Board)

	for _, identifier := range cardIdentifiers {
		board, _, card := handler.store.FindCard(identifier)
		if card == nil {
			continue
		}
		affectedBoards[board.Identifier] = board
		updates.applyTo(card)
	}

	if saveError := handler.saveBoards(affectedBoards); saveError != nil {
		return nil, saveError
	}
	return okResponse(), nil
}

func (handler *Handler) BatchDeleteCards(cardIdentifiers []int64) (response, error) {
	if len(cardIdentifiers) == 0 {
		return okResponse(), nil
	}

	affectedBoards := make(map[int64]*database.Board)
	foldersToRemove := make([]string, 0, len(cardIdentifiers))

	for _, identifier := range cardIdentifiers {
		board, list, card := handler.store.FindCard(identifier)
		if card == nil || list == nil {
			continue
		}
		affectedBoards[board.Identifier] = board
		list.Cards = database.RemoveByIdentifier(list.Cards, identifier)
		foldersToRemove = append(foldersToRemove, joinDirectoryPaths(handler.store.AttachmentDirectory(board), cardAttachmentFolders([]*database.Card{card}))...)
	}

	if saveError := handler.saveBoards(affectedBoards); saveError != nil {
		return nil, saveError
	}
	removeDirectories(foldersToRemove)

	return okResponse(), nil
}

func (handler *Handler) DuplicateCards(cardIdentifiers []int64) (response, error) {
	if len(cardIdentifiers) == 0 {
		return okResponse(), nil
	}

	affectedBoards := make(map[int64]*database.Board)
	createdCardIdentifiers := make([]int64, 0, len(cardIdentifiers))

	for _, identifier := range cardIdentifiers {
		board, list, originalCard := handler.store.FindCard(identifier)
		if originalCard == nil || list == nil {
			continue
		}
		affectedBoards[board.Identifier] = board

		clonedCard := handler.cloneCard(board, originalCard, originalCard.Title+" (Copy)")
		list.Cards = database.InsertAfterIdentifier(list.Cards, identifier, clonedCard)
		createdCardIdentifiers = append(createdCardIdentifiers, clonedCard.Identifier)
	}

	if saveError := handler.saveBoards(affectedBoards); saveError != nil {
		return nil, saveError
	}
	return okResponseWithCreatedIdentifiers(createdCardIdentifiers), nil
}
