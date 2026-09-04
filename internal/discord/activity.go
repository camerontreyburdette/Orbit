package discord

import (
	"fmt"
	"strings"
)

const maximumActivityTextLength = 128

type Context struct {
	View       string `json:"view"`
	BoardName  string `json:"board_name"`
	CardTitle  string `json:"card_title"`
	IsEditing  bool   `json:"editing"`
	ListCount  int    `json:"lists"`
	CardCount  int    `json:"cards"`
	BoardCount int    `json:"boards"`
}

type Timestamps struct {
	Start int64 `json:"start,omitempty"`
}

type Assets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
}

type Activity struct {
	Details    string      `json:"details,omitempty"`
	State      string      `json:"state,omitempty"`
	Timestamps *Timestamps `json:"timestamps,omitempty"`
	Assets     *Assets     `json:"assets,omitempty"`
}

func truncateActivityText(text string) string {
	if len(text) > maximumActivityTextLength {
		return text[:maximumActivityTextLength]
	}
	return text
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func homeState(context Context) string {
	if context.BoardCount <= 0 {
		return ""
	}
	return fmt.Sprintf("%d %s", context.BoardCount, pluralize(context.BoardCount, "board", "boards"))
}

func boardState(context Context) string {
	cleanCardTitle := strings.TrimSpace(context.CardTitle)
	if cleanCardTitle != "" {
		actionVerb := "Viewing"
		if context.IsEditing {
			actionVerb = "Editing"
		}
		return truncateActivityText(fmt.Sprintf("%s card: %s", actionVerb, cleanCardTitle))
	}
	return fmt.Sprintf(
		"%d %s · %d %s",
		context.ListCount, pluralize(context.ListCount, "list", "lists"),
		context.CardCount, pluralize(context.CardCount, "card", "cards"),
	)
}

func describeContext(context Context) (string, string) {
	cleanBoardName := strings.TrimSpace(context.BoardName)
	if context.View == "board" && cleanBoardName != "" {
		return truncateActivityText("Board: " + cleanBoardName), boardState(context)
	}
	return "Browsing boards", homeState(context)
}

func (client *Client) BuildActivity(context Context) Activity {
	details, state := describeContext(context)
	return Activity{
		Details:    details,
		State:      state,
		Timestamps: &Timestamps{Start: client.sessionStart},
		Assets:     &Assets{LargeImage: "orbit", LargeText: "Orbit"},
	}
}
