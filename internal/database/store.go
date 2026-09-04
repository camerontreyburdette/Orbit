package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

const attachmentsDirectoryName = "attachments"

type Store struct {
	directory      string
	mutex          sync.RWMutex
	boards         []*Board
	boardsByID     map[int64]*Board
	directories    map[int64]string
	indexes        map[int64]*boardIndex
	nextIdentifier int64
}

func NewStore(boardsDirectory string) (*Store, error) {
	if createError := os.MkdirAll(boardsDirectory, 0750); createError != nil {
		return nil, fmt.Errorf("failed to create boards directory: %w", createError)
	}

	loaded, loadError := loadBoardsFromDirectory(boardsDirectory)
	if loadError != nil {
		return nil, loadError
	}

	store := &Store{
		directory:   boardsDirectory,
		boards:      make([]*Board, 0, len(loaded)),
		boardsByID:  make(map[int64]*Board, len(loaded)),
		directories: make(map[int64]string, len(loaded)),
		indexes:     make(map[int64]*boardIndex, len(loaded)),
	}

	var maximumIdentifier int64
	for _, entry := range loaded {
		store.boards = append(store.boards, entry.board)
		store.boardsByID[entry.board.Identifier] = entry.board
		store.directories[entry.board.Identifier] = entry.directoryName
		index := buildBoardIndex(entry.board)
		store.indexes[entry.board.Identifier] = index
		if index.maximum > maximumIdentifier {
			maximumIdentifier = index.maximum
		}
	}
	store.nextIdentifier = maximumIdentifier + 1
	if persistedNextIdentifier := readIdentifierCounter(boardsDirectory); persistedNextIdentifier > store.nextIdentifier {
		store.nextIdentifier = persistedNextIdentifier
	}

	return store, nil
}

func (store *Store) NewIdentifier() int64 {
	return atomic.AddInt64(&store.nextIdentifier, 1) - 1
}

func (store *Store) Directory() string {
	return store.directory
}

func (store *Store) RemoveAllBoards() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.boards = make([]*Board, 0)
	store.boardsByID = make(map[int64]*Board)
	store.directories = make(map[int64]string)
	store.indexes = make(map[int64]*boardIndex)

	if removeError := os.RemoveAll(store.directory); removeError != nil {
		return fmt.Errorf("failed to remove boards directory: %w", removeError)
	}
	if createError := os.MkdirAll(store.directory, 0750); createError != nil {
		return fmt.Errorf("failed to recreate boards directory: %w", createError)
	}

	atomic.StoreInt64(&store.nextIdentifier, 1)
	return nil
}

func (store *Store) Boards() []*Board {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	copied := make([]*Board, len(store.boards))
	copy(copied, store.boards)
	return copied
}

func (store *Store) AddBoard(board *Board) error {
	store.mutex.Lock()
	store.boards = append(store.boards, board)
	store.boardsByID[board.Identifier] = board
	store.indexes[board.Identifier] = buildBoardIndex(board)
	store.mutex.Unlock()
	return store.SaveBoard(board)
}

func (store *Store) BoardDirectory(board *Board) string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	return filepath.Join(store.directory, store.directories[board.Identifier])
}

func (store *Store) AttachmentDirectory(board *Board) string {
	return filepath.Join(store.BoardDirectory(board), attachmentsDirectoryName)
}

func (store *Store) takenDirectoryNamesExcluding(boardIdentifier int64) map[string]struct{} {
	takenDirectoryNames := make(map[string]struct{}, len(store.directories))
	ownDirectoryName := store.directories[boardIdentifier]
	for directoryName := range listDirectoryNames(store.directory) {
		if !strings.EqualFold(directoryName, ownDirectoryName) {
			takenDirectoryNames[strings.ToLower(directoryName)] = struct{}{}
		}
	}
	for currentBoardIdentifier, directoryName := range store.directories {
		if currentBoardIdentifier != boardIdentifier {
			takenDirectoryNames[strings.ToLower(directoryName)] = struct{}{}
		}
	}
	return takenDirectoryNames
}

func (store *Store) resolveBoardDirectoryName(board *Board) string {
	targetDirectoryName := BoardDirectoryName(board, store.takenDirectoryNamesExcluding(board.Identifier))
	previousDirectoryName := store.directories[board.Identifier]
	if previousDirectoryName == "" || strings.EqualFold(previousDirectoryName, targetDirectoryName) {
		return targetDirectoryName
	}

	previousFullPath := filepath.Join(store.directory, previousDirectoryName)
	targetFullPath := filepath.Join(store.directory, targetDirectoryName)
	if renameError := os.Rename(previousFullPath, targetFullPath); renameError != nil {
		return previousDirectoryName
	}
	return targetDirectoryName
}

func (store *Store) SaveBoard(board *Board) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.indexes[board.Identifier] = buildBoardIndex(board)

	targetDirectoryName := store.resolveBoardDirectoryName(board)
	boardFullPath := filepath.Join(store.directory, targetDirectoryName)
	if mkdirError := os.MkdirAll(boardFullPath, 0750); mkdirError != nil {
		return fmt.Errorf("failed to create board directory: %w", mkdirError)
	}

	if writeError := AtomicWriteJSON(filepath.Join(boardFullPath, boardFileName), board); writeError != nil {
		return fmt.Errorf("failed to save board file: %w", writeError)
	}

	store.directories[board.Identifier] = targetDirectoryName
	_ = writeIdentifierCounter(store.directory, atomic.LoadInt64(&store.nextIdentifier))
	return nil
}

func (store *Store) RemoveBoard(board *Board) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.boards = RemoveByIdentifier(store.boards, board.Identifier)
	delete(store.boardsByID, board.Identifier)
	delete(store.indexes, board.Identifier)

	directoryName, exists := store.directories[board.Identifier]
	delete(store.directories, board.Identifier)

	if exists && directoryName != "" {
		if removeError := os.RemoveAll(filepath.Join(store.directory, directoryName)); removeError != nil {
			return fmt.Errorf("failed to delete board directory: %w", removeError)
		}
	}

	return nil
}

func (store *Store) Board(boardIdentifier int64) *Board {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	return store.boardsByID[boardIdentifier]
}

func (store *Store) FindList(listIdentifier int64) (*Board, *List) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, board := range store.boards {
		if location, exists := store.indexes[board.Identifier].lists[listIdentifier]; exists {
			return location.Board, location.List
		}
	}
	return nil, nil
}

func (store *Store) FindCard(cardIdentifier int64) (*Board, *List, *Card) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, board := range store.boards {
		if location, exists := store.indexes[board.Identifier].cards[cardIdentifier]; exists {
			return location.Board, location.List, location.Card
		}
	}
	return nil, nil, nil
}

func (store *Store) FindAttachment(attachmentIdentifier int64) (*Board, *List, *Card, *Attachment) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, board := range store.boards {
		if location, exists := store.indexes[board.Identifier].attachments[attachmentIdentifier]; exists {
			return location.Board, location.List, location.Card, location.Attachment
		}
	}
	return nil, nil, nil, nil
}

func (store *Store) FindChecklist(checklistIdentifier int64) (*Board, *Card, *Checklist) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, board := range store.boards {
		if location, exists := store.indexes[board.Identifier].checklists[checklistIdentifier]; exists {
			return location.Board, location.Card, location.Checklist
		}
	}
	return nil, nil, nil
}

func (store *Store) FindChecklistItem(itemIdentifier int64) (*Board, *Card, *Checklist, *ChecklistItem) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, board := range store.boards {
		if location, exists := store.indexes[board.Identifier].checklistItems[itemIdentifier]; exists {
			return location.Board, location.Card, location.Checklist, location.Item
		}
	}
	return nil, nil, nil, nil
}
