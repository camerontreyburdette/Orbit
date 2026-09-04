package api

import (
	"fmt"
	"regexp"
	"strings"

	"orbit/internal/database"
)

var hexadecimalColorRegularExpression = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type cardUpdates struct {
	title           *string
	description     *string
	color           *string
	tags            []string
	hasTags         bool
	coverIdentifier *int64
	hasCover        bool
}

func normalizeTagText(rawTag string) string {
	cleaned := strings.TrimPrefix(strings.TrimSpace(rawTag), "#")
	return database.TruncateRunes(cleaned, maximumTagCharacters)
}

func containsTagIgnoringCase(tags []string, candidate string) bool {
	for _, existingTag := range tags {
		if strings.EqualFold(existingTag, candidate) {
			return true
		}
	}
	return false
}

func parseTagList(rawTags interface{}) ([]string, error) {
	switch value := rawTags.(type) {
	case []interface{}:
		tags := make([]string, 0, len(value))
		for _, tagItem := range value {
			tags = append(tags, fmt.Sprintf("%v", tagItem))
		}
		return tags, nil
	case []string:
		return value, nil
	}
	return nil, fmt.Errorf("Tags must be a list")
}

func sanitizeTagList(rawTags []string) []string {
	cleanTags := make([]string, 0, len(rawTags))
	seenTags := make(map[string]struct{}, len(rawTags))
	for _, rawTag := range rawTags {
		cleaned := normalizeTagText(rawTag)
		if cleaned == "" {
			continue
		}
		lowercaseTag := strings.ToLower(cleaned)
		if _, isSeen := seenTags[lowercaseTag]; isSeen {
			continue
		}
		seenTags[lowercaseTag] = struct{}{}
		cleanTags = append(cleanTags, cleaned)
	}
	if len(cleanTags) > maximumTagsCount {
		return cleanTags[:maximumTagsCount]
	}
	return cleanTags
}

func parseCoverIdentifier(rawCoverIdentifier interface{}) *int64 {
	var parsedCoverIdentifier int64
	switch value := rawCoverIdentifier.(type) {
	case float64:
		parsedCoverIdentifier = int64(value)
	case int64:
		parsedCoverIdentifier = value
	}
	if parsedCoverIdentifier == 0 {
		return nil
	}
	return &parsedCoverIdentifier
}

func parseCardUpdates(fields map[string]interface{}) (cardUpdates, bool, error) {
	var updates cardUpdates
	hasUpdates := false

	if rawTitle, hasTitle := fields["title"]; hasTitle {
		hasUpdates = true
		title := strings.TrimSpace(fmt.Sprintf("%v", rawTitle))
		if title == "" {
			return updates, hasUpdates, fmt.Errorf("Card title cannot be empty")
		}
		updates.title = &title
	}

	if rawDescription, hasDescription := fields["description"]; hasDescription {
		hasUpdates = true
		description := fmt.Sprintf("%v", rawDescription)
		updates.description = &description
	}

	if rawColor, hasColor := fields["color"]; hasColor {
		hasUpdates = true
		color := strings.TrimSpace(fmt.Sprintf("%v", rawColor))
		if color != "" && !hexadecimalColorRegularExpression.MatchString(color) {
			return updates, hasUpdates, fmt.Errorf("Invalid color")
		}
		updates.color = &color
	}

	if rawTags, hasTags := fields["tags"]; hasTags {
		hasUpdates = true
		tags, tagsError := parseTagList(rawTags)
		if tagsError != nil {
			return updates, hasUpdates, tagsError
		}
		updates.tags = sanitizeTagList(tags)
		updates.hasTags = true
	}

	if rawCoverIdentifier, hasCoverIdentifier := fields["cover_id"]; hasCoverIdentifier {
		hasUpdates = true
		updates.hasCover = true
		if rawCoverIdentifier != nil {
			updates.coverIdentifier = parseCoverIdentifier(rawCoverIdentifier)
		}
	}

	return updates, hasUpdates, nil
}

func findImageAttachment(card *database.Card, attachmentIdentifier int64) *database.Attachment {
	index := database.IndexOfIdentifier(card.Attachments, attachmentIdentifier)
	if index < 0 || card.Attachments[index].Kind != "image" {
		return nil
	}
	return card.Attachments[index]
}

func (updates cardUpdates) applyTo(card *database.Card) error {
	if updates.hasCover {
		if updates.coverIdentifier != nil && findImageAttachment(card, *updates.coverIdentifier) == nil {
			return fmt.Errorf("Cover must be an image attached to this card")
		}
		card.CoverIdentifier = updates.coverIdentifier
	}
	if updates.title != nil {
		card.Title = *updates.title
	}
	if updates.description != nil {
		card.Description = *updates.description
	}
	if updates.color != nil {
		card.Color = *updates.color
	}
	if updates.hasTags {
		card.Tags = updates.tags
	}
	return nil
}

type batchCardUpdates struct {
	color        string
	hasColor     bool
	addTag       string
	hasAddTag    bool
	removeTag    string
	hasRemoveTag bool
}

func parseBatchCardUpdates(fields map[string]interface{}) batchCardUpdates {
	var updates batchCardUpdates
	updates.color, updates.hasColor = fields["color"].(string)

	if addTagValue, hasAddTag := fields["add_tag"].(string); hasAddTag {
		updates.hasAddTag = true
		updates.addTag = strings.TrimPrefix(strings.TrimSpace(addTagValue), "#")
	}
	if removeTagValue, hasRemoveTag := fields["remove_tag"].(string); hasRemoveTag {
		updates.hasRemoveTag = true
		updates.removeTag = strings.TrimPrefix(strings.TrimSpace(removeTagValue), "#")
	}
	return updates
}

func (updates batchCardUpdates) applyTo(card *database.Card) {
	if updates.hasColor && (updates.color == "" || hexadecimalColorRegularExpression.MatchString(updates.color)) {
		card.Color = updates.color
	}
	if updates.hasRemoveTag && updates.removeTag != "" {
		card.Tags = removeTagIgnoringCase(card.Tags, updates.removeTag)
	}
	if updates.hasAddTag && updates.addTag != "" && !containsTagIgnoringCase(card.Tags, updates.addTag) && len(card.Tags) < maximumTagsCount {
		card.Tags = append(card.Tags, database.TruncateRunes(updates.addTag, maximumTagCharacters))
	}
}

func removeTagIgnoringCase(tags []string, tagToRemove string) []string {
	filteredTags := make([]string, 0, len(tags))
	for _, existingTag := range tags {
		if !strings.EqualFold(existingTag, tagToRemove) {
			filteredTags = append(filteredTags, existingTag)
		}
	}
	return filteredTags
}
