package api

import "orbit/internal/database"

type boardView struct {
	Board database.BoardSummary `json:"board"`
	Lists []listView            `json:"lists"`
}

type listView struct {
	Identifier int64           `json:"id"`
	Name       string          `json:"name"`
	Cards      []database.Card `json:"cards"`
}

func buildBoardView(board *database.Board) boardView {
	lists := make([]listView, 0, len(board.Lists))
	for _, list := range board.Lists {
		lists = append(lists, buildListView(list))
	}
	return boardView{Board: board.Summary(), Lists: lists}
}

func buildListView(list *database.List) listView {
	cards := make([]database.Card, 0, len(list.Cards))
	for _, card := range list.Cards {
		cards = append(cards, buildCardView(card, list.Identifier))
	}
	return listView{Identifier: list.Identifier, Name: list.Name, Cards: cards}
}

func buildCardView(card *database.Card, listIdentifier int64) database.Card {
	cardView := *card
	cardView.ListIdentifier = listIdentifier
	if cardView.Attachments == nil {
		cardView.Attachments = make([]*database.Attachment, 0)
	}
	return cardView
}
