package bot

import (
	"strconv"
	"strings"
)

func (b *Bot) IssueIDFromURL(url string) int {
	parts := strings.Split(url, "/")
	id, _ := strconv.Atoi(parts[len(parts)-1])
	return id
}
