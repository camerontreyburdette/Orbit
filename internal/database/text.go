package database

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const maximumSafeNameCharacters = 60

var (
	boldRegularExpression        = regexp.MustCompile(`\*\*(.+?)\*\*`)
	underlineRegularExpression   = regexp.MustCompile(`(^|[^\w])_([^_]+)_([^\w]|$)`)
	strikeRegularExpression      = regexp.MustCompile(`(^|[^\w-])-([^\s-]|[^\s-][^-]*[^\s-])-([^\w-]|$)`)
	italicRegularExpression      = regexp.MustCompile(`\*([^*]+)\*`)
	colorMarkupRegularExpression = regexp.MustCompile(`(?i)\[(?:red|orange|yellow|green|teal|blue|purple|pink)\s*:\s*([^\]]*)\]`)
	colorPrefixRegularExpression = regexp.MustCompile(`(?i)\[(?:red|orange|yellow|green|teal|blue|purple|pink)\s*:\s*`)
	illegalFileCharacters        = regexp.MustCompile(`[\\/:*?"<>|]`)
	multipleSpaces               = regexp.MustCompile(`\s+`)
	formattingCharacters         = regexp.MustCompile(`[*_\]]`)
)

var reservedDeviceNames = map[string]struct{}{
	"con": {}, "prn": {}, "aux": {}, "nul": {},
	"com1": {}, "com2": {}, "com3": {}, "com4": {}, "com5": {}, "com6": {}, "com7": {}, "com8": {}, "com9": {},
	"lpt1": {}, "lpt2": {}, "lpt3": {}, "lpt4": {}, "lpt5": {}, "lpt6": {}, "lpt7": {}, "lpt8": {}, "lpt9": {},
}

func FormatTimestampNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func TruncateRunes(text string, maximumCharacters int) string {
	characterRunes := []rune(text)
	if len(characterRunes) <= maximumCharacters {
		return text
	}
	return string(characterRunes[:maximumCharacters])
}

func StripMarkup(text string) string {
	result := boldRegularExpression.ReplaceAllString(text, "$1")
	result = underlineRegularExpression.ReplaceAllString(result, "$1$2$3")
	result = strikeRegularExpression.ReplaceAllString(result, "$1$2$3")
	result = italicRegularExpression.ReplaceAllString(result, "$1")
	return colorMarkupRegularExpression.ReplaceAllString(result, "$1")
}

func PlainName(text string) string {
	strippedColor := colorPrefixRegularExpression.ReplaceAllString(text, "")
	cleanText := formattingCharacters.ReplaceAllString(strippedColor, "")
	return strings.ToLower(cleanText)
}

func SanitizeSafeName(rawText string) string {
	name := StripMarkup(rawText)
	name = illegalFileCharacters.ReplaceAllString(name, "")
	name = multipleSpaces.ReplaceAllString(name, " ")
	name = strings.Trim(name, " .")
	name = TruncateRunes(name, maximumSafeNameCharacters)
	return strings.Trim(name, " .")
}

func uniqueDirectoryName(rawName string, fallbackPrefix string, identifier int64, takenDirectoryNames map[string]struct{}) string {
	name := SanitizeSafeName(rawName)
	if _, isReserved := reservedDeviceNames[strings.ToLower(name)]; name == "" || isReserved {
		name = fmt.Sprintf("%s_%d", fallbackPrefix, identifier)
	}
	if _, isTaken := takenDirectoryNames[strings.ToLower(name)]; isTaken {
		name = fmt.Sprintf("%s_%d", name, identifier)
	}
	return name
}

func BoardDirectoryName(board *Board, takenDirectoryNames map[string]struct{}) string {
	return uniqueDirectoryName(board.Name, "board", board.Identifier, takenDirectoryNames)
}

func CardDirectoryName(card *Card, takenDirectoryNames map[string]struct{}) string {
	return uniqueDirectoryName(card.Title, "card", card.Identifier, takenDirectoryNames)
}

func CardAttachmentFolder(card *Card) string {
	if len(card.Attachments) == 0 {
		return ""
	}
	folder, _, _ := strings.Cut(card.Attachments[0].File, "/")
	return folder
}
