package database

type Identifiable interface {
	EntityIdentifier() int64
}

type Board struct {
	Identifier       int64   `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Pinned           int     `json:"pinned"`
	PinnedAt         *string `json:"pinned_at"`
	CreatedAt        string  `json:"created_at"`
	TimeSpentSeconds int64   `json:"time_spent_seconds"`
	Lists            []*List `json:"lists"`
}

type BoardSummary struct {
	Identifier       int64   `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Pinned           int     `json:"pinned"`
	PinnedAt         *string `json:"pinned_at"`
	CreatedAt        string  `json:"created_at"`
	TimeSpentSeconds int64   `json:"time_spent_seconds"`
}

type List struct {
	Identifier int64   `json:"id"`
	Name       string  `json:"name"`
	Cards      []*Card `json:"cards"`
}

type Card struct {
	Identifier      int64         `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Color           string        `json:"color"`
	Tags            []string      `json:"tags"`
	CoverIdentifier *int64        `json:"cover_id"`
	CreatedAt       string        `json:"created_at"`
	Checklists      []*Checklist  `json:"checklists"`
	Attachments     []*Attachment `json:"attachments"`
	ListIdentifier  int64         `json:"list_id,omitempty"`
}

type Checklist struct {
	Identifier int64            `json:"id"`
	Title      string           `json:"title"`
	Items      []*ChecklistItem `json:"items"`
}

type ChecklistItem struct {
	Identifier int64  `json:"id"`
	Text       string `json:"text"`
	Done       bool   `json:"done"`
}

type Attachment struct {
	Identifier int64  `json:"id"`
	Name       string `json:"name"`
	File       string `json:"file"`
	Kind       string `json:"kind"`
	Size       int64  `json:"size"`
	CreatedAt  string `json:"created_at"`
}

func (board *Board) EntityIdentifier() int64 {
	return board.Identifier
}

func (list *List) EntityIdentifier() int64 {
	return list.Identifier
}

func (card *Card) EntityIdentifier() int64 {
	return card.Identifier
}

func (checklist *Checklist) EntityIdentifier() int64 {
	return checklist.Identifier
}

func (item *ChecklistItem) EntityIdentifier() int64 {
	return item.Identifier
}

func (attachment *Attachment) EntityIdentifier() int64 {
	return attachment.Identifier
}

func (board *Board) Summary() BoardSummary {
	return BoardSummary{
		Identifier:       board.Identifier,
		Name:             board.Name,
		Description:      board.Description,
		Pinned:           board.Pinned,
		PinnedAt:         board.PinnedAt,
		CreatedAt:        board.CreatedAt,
		TimeSpentSeconds: board.TimeSpentSeconds,
	}
}
