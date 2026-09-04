package api

import "orbit/internal/database"

func (handler *Handler) assignFreshBoardIdentifiers(board *database.Board) {
	board.Identifier = handler.store.NewIdentifier()
	for _, list := range board.Lists {
		list.Identifier = handler.store.NewIdentifier()
		for _, card := range list.Cards {
			handler.assignFreshCardIdentifiers(card)
		}
	}
}

func (handler *Handler) assignFreshCardIdentifiers(card *database.Card) {
	card.Identifier = handler.store.NewIdentifier()
	for _, checklist := range card.Checklists {
		checklist.Identifier = handler.store.NewIdentifier()
		for _, item := range checklist.Items {
			item.Identifier = handler.store.NewIdentifier()
		}
	}
	card.CoverIdentifier = handler.assignFreshAttachmentIdentifiers(card.Attachments, card.CoverIdentifier)
}

func (handler *Handler) assignFreshAttachmentIdentifiers(attachments []*database.Attachment, previousCoverIdentifier *int64) *int64 {
	var freshCoverIdentifier *int64
	for _, attachment := range attachments {
		previousIdentifier := attachment.Identifier
		attachment.Identifier = handler.store.NewIdentifier()
		if previousCoverIdentifier != nil && *previousCoverIdentifier == previousIdentifier {
			coverIdentifier := attachment.Identifier
			freshCoverIdentifier = &coverIdentifier
		}
	}
	return freshCoverIdentifier
}
