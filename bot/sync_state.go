package bot

import (
	"github.com/google/go-github/v33/github"
)

type SyncState struct {
	Name          string
	ProjectColumn *github.ProjectColumn
	Labels        []*github.Label
}

// InState checks if the set of provided labels include **all** of the labels
// in the SyncState.
func (ss *SyncState) InState(labels []string) bool {
	for _, requiredLabel := range ss.Labels {
		found := false
		for _, label := range labels {
			if *requiredLabel.Name == label {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

// Correct column checks to see if the issue is in the correct column.
func (b *Bot) IsInCorrectColumn(ss *SyncState, issueURL string) bool {
	columnCards, err := b.getProjectCards(ss.ProjectColumn)
	if err != nil {
		b.log.Warnf("Could not get project cards from column '%s'\n", *ss.ProjectColumn.Name)
		return false
	}

	b.log.Debugf("Looking for card with contentURL='%s'\n", issueURL)
	for _, card := range columnCards {
		if *card.ContentURL == issueURL {
			return true
		}
	}
	return false
}

func (b *Bot) MoveCard(ss *SyncState, cardID int64) error {
	b.log.Fatal("UNIMPLEMENTED")
	return nil
}
