package bot

import (
	"context"

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

// MoveCard moves the card with given ID to the column of the targetState
func (b *Bot) MoveCard(targetState *SyncState, cardID int64) error {
	_, err := b.gh.Projects.MoveProjectCard(
		context.Background(),
		cardID,
		&github.ProjectCardMoveOptions{Position: "bottom", ColumnID: targetState.ProjectColumn.GetID()})
	return err
}

// CreateCard creates a new project card in the column of the targetState where
// the content of the card links to the issue
func (b *Bot) CreateCard(targetState *SyncState, issue *github.Issue) (*github.ProjectCard, error) {
	card, _, err := b.gh.Projects.CreateProjectCard(context.Background(),
		targetState.ProjectColumn.GetID(),
		&github.ProjectCardOptions{ContentID: issue.GetID(), ContentType: "Issue"})

	if err != nil {
		return nil, err
	}

	b.cachedCards[card.GetID()] = card
	return card, nil
}
