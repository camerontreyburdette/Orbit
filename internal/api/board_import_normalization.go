package api

import (
	"os"
	"path/filepath"
	"strings"

	"orbit/internal/database"
)

const (
	untitledBoardName = "Untitled board"
	untitledCardTitle = "Untitled card"
)

func withoutNilEntries[Item any](items []*Item) []*Item {
	compacted := make([]*Item, 0, len(items))
	for _, item := range items {
		if item != nil {
			compacted = append(compacted, item)
		}
	}
	return compacted
}

func normalizeImportedBoard(board *database.Board) {
	board.Name = strings.TrimSpace(board.Name)
	if board.Name == "" {
		board.Name = untitledBoardName
	}
	board.Description = strings.TrimSpace(board.Description)
	if board.CreatedAt == "" {
		board.CreatedAt = database.FormatTimestampNow()
	}
	if board.TimeSpentSeconds < 0 {
		board.TimeSpentSeconds = 0
	}

	board.Lists = withoutNilEntries(board.Lists)
	for _, list := range board.Lists {
		normalizeImportedList(list)
	}
}

func normalizeImportedList(list *database.List) {
	list.Name = strings.TrimSpace(list.Name)
	list.Cards = withoutNilEntries(list.Cards)
	for _, card := range list.Cards {
		normalizeImportedCard(card)
	}
}

func normalizeImportedCard(card *database.Card) {
	card.Title = strings.TrimSpace(card.Title)
	if card.Title == "" {
		card.Title = untitledCardTitle
	}
	if card.CreatedAt == "" {
		card.CreatedAt = database.FormatTimestampNow()
	}
	if card.Tags == nil {
		card.Tags = make([]string, 0)
	}
	card.ListIdentifier = 0
	card.Checklists = withoutNilEntries(card.Checklists)
	for _, checklist := range card.Checklists {
		checklist.Items = withoutNilEntries(checklist.Items)
	}
	card.Attachments = withoutNilEntries(card.Attachments)
}

func attachmentFileExists(attachmentDirectory string, attachment *database.Attachment) bool {
	if attachment.File == "" {
		return false
	}
	_, statError := os.Stat(filepath.Join(attachmentDirectory, filepath.FromSlash(attachment.File)))
	return statError == nil
}

func pruneMissingCardAttachments(card *database.Card, attachmentDirectory string) {
	presentAttachments := make([]*database.Attachment, 0, len(card.Attachments))
	hasCover := false
	for _, attachment := range card.Attachments {
		if !attachmentFileExists(attachmentDirectory, attachment) {
			continue
		}
		presentAttachments = append(presentAttachments, attachment)
		if card.CoverIdentifier != nil && *card.CoverIdentifier == attachment.Identifier {
			hasCover = true
		}
	}
	card.Attachments = presentAttachments
	if !hasCover {
		card.CoverIdentifier = nil
	}
}

func pruneMissingAttachments(board *database.Board, attachmentDirectory string) {
	for _, list := range board.Lists {
		for _, card := range list.Cards {
			pruneMissingCardAttachments(card, attachmentDirectory)
		}
	}
}
