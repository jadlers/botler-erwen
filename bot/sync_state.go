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

// GetCorrectLabels returns the correct labels according to the targetState. It
// also returns if any changes to the current labels where made. All labels
// which are not SyncState labels are ignored.
func (b *Bot) GetCorrectLabels(targetState *SyncState, issue *github.Issue) ([]string, bool) {
	changed := false
	issueLabels := map[string]bool{} // Label name is key, val is true if it should be kept
	for _, label := range issue.Labels {
		issueLabels[label.GetName()] = true
	}

	// Find set of all state labels
	invalidLabels := map[string]bool{}
	for _, ss := range b.SyncStates {
		for _, label := range ss.Labels {
			invalidLabels[label.GetName()] = true
		}
	}

	// Add state labels to issueLabels and remove from invalid labels
	for _, validLabel := range targetState.Labels {
		labelName := validLabel.GetName()
		delete(invalidLabels, labelName)
		if _, exist := issueLabels[labelName]; !exist {
			changed = true
			issueLabels[labelName] = true
			b.log.Debugf("Adding missing label: %s\n", labelName)
		}
	}

	// Remove invalid labels
	for labelName := range issueLabels {
		if _, exist := invalidLabels[labelName]; exist {
			changed = true
			issueLabels[labelName] = false
			b.log.Debugf("Removing invalid label: %s\n", labelName)
		}
	}

	finalLabels := []string{}
	for labelName, include := range issueLabels {
		if include {
			finalLabels = append(finalLabels, labelName)
		}
	}

	b.log.Debugf("Final list of labels: %+v\n", finalLabels)

	return finalLabels, changed
}

// SetIssueLabels sets the labels on the issue so the end result is is having
// the **only** labels provided
func (b *Bot) SetIssueLabels(issue *github.Issue, labels []string) error {
	_, _, err := b.gh.Issues.ReplaceLabelsForIssue(context.Background(), b.conf.Owner, b.conf.Repository, issue.GetNumber(), labels)
	return err
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

	// Move the newly created card to the bottom in order to preserve oldest on top
	b.MoveCard(targetState, card.GetID())

	b.cachedCards[card.GetID()] = card
	return card, nil
}
