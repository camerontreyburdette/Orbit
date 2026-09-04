package api

import (
	"fmt"
	"strings"

	"orbit/internal/database"
)

func (handler *Handler) CreateList(boardIdentifier int64, name string) (response, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Untitled list"
	}

	board := handler.store.Board(boardIdentifier)
	if board == nil {
		return nil, fmt.Errorf("Board not found")
	}

	newList := &database.List{
		Identifier: handler.store.NewIdentifier(),
		Name:       name,
		Cards:      make([]*database.Card, 0),
	}

	board.Lists = append(board.Lists, newList)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}

	return response{"id": newList.Identifier}, nil
}

func (handler *Handler) RenameList(listIdentifier int64, name string) (response, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("List name cannot be empty")
	}

	board, list := handler.store.FindList(listIdentifier)
	if list == nil {
		return nil, fmt.Errorf("List not found")
	}

	list.Name = name
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) DeleteList(listIdentifier int64) (response, error) {
	board, list := handler.store.FindList(listIdentifier)
	if list == nil {
		return nil, fmt.Errorf("List not found")
	}

	attachmentFolders := cardAttachmentFolders(list.Cards)
	board.Lists = database.RemoveByIdentifier(board.Lists, listIdentifier)

	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}
	handler.removeCardAttachmentFolders(board, attachmentFolders)

	return okResponse(), nil
}

func (handler *Handler) MoveList(listIdentifier int64, newIndex int) (response, error) {
	board, list := handler.store.FindList(listIdentifier)
	if list == nil {
		return nil, fmt.Errorf("List not found")
	}

	board.Lists = database.MoveToIndex(board.Lists, list, newIndex)
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) MoveLists(listIdentifiers []int64, newIndex int) (response, error) {
	if len(listIdentifiers) == 0 {
		return okResponse(), nil
	}

	board, firstList := handler.store.FindList(listIdentifiers[0])
	if firstList == nil {
		return nil, fmt.Errorf("list not found")
	}

	listsToMove, remainingLists := database.PartitionByIdentifiers(board.Lists, database.IdentifierSet(listIdentifiers))
	board.Lists = database.InsertAllAtIndex(remainingLists, newIndex, listsToMove)
	return handler.saveBoardOrFail(board)
}

func (handler *Handler) DuplicateLists(listIdentifiers []int64) (response, error) {
	if len(listIdentifiers) == 0 {
		return okResponse(), nil
	}

	affectedBoards := make(map[int64]*database.Board)
	createdListIdentifiers := make([]int64, 0, len(listIdentifiers))

	for _, identifier := range listIdentifiers {
		board, originalList := handler.store.FindList(identifier)
		if originalList == nil {
			continue
		}
		affectedBoards[board.Identifier] = board

		clonedList := handler.cloneList(board, originalList)
		board.Lists = database.InsertAfterIdentifier(board.Lists, identifier, clonedList)
		createdListIdentifiers = append(createdListIdentifiers, clonedList.Identifier)
	}

	if saveError := handler.saveBoards(affectedBoards); saveError != nil {
		return nil, saveError
	}
	return okResponseWithCreatedIdentifiers(createdListIdentifiers), nil
}

func (handler *Handler) BatchDeleteLists(listIdentifiers []int64) (response, error) {
	if len(listIdentifiers) == 0 {
		return okResponse(), nil
	}

	targetIdentifiers := database.IdentifierSet(listIdentifiers)
	affectedBoards := make(map[int64]*database.Board)
	attachmentFoldersToDelete := make([]string, 0)

	for _, identifier := range listIdentifiers {
		board, list := handler.store.FindList(identifier)
		if list == nil {
			continue
		}
		affectedBoards[board.Identifier] = board
		attachmentFoldersToDelete = append(attachmentFoldersToDelete, joinDirectoryPaths(handler.store.AttachmentDirectory(board), cardAttachmentFolders(list.Cards))...)
	}

	for _, affectedBoard := range affectedBoards {
		_, remainingLists := database.PartitionByIdentifiers(affectedBoard.Lists, targetIdentifiers)
		affectedBoard.Lists = remainingLists
	}
	if saveError := handler.saveBoards(affectedBoards); saveError != nil {
		return nil, saveError
	}

	removeDirectories(attachmentFoldersToDelete)
	return okResponse(), nil
}
