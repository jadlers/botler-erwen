package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
)

func (b *Bot) getProject(name string) (*github.Project, error) {
	if project, ok := b.cachedProjects[name]; ok {
		b.log.WithField("project", name).Debugln("Using cached project")
		return project, nil
	}

	projects, _, err := b.gh.Repositories.ListProjects(context.Background(),
		b.conf.Owner,
		b.conf.Repository,
		&github.ProjectListOptions{State: "open"})

	if err != nil {
		return nil, err
	}

	b.log.Debugln("Listing projects in the repo:")
	var project *github.Project
	for i, p := range projects {
		b.log.Debugf("\t#%d: %s (id: %d)\n", i, *p.Name, p.ID)
		if *p.Name == name {
			project = p
			break
		}
	}

	if project == nil {
		return nil, fmt.Errorf("Could not find project named %s\n", name)
	}

	b.log.WithFields(logrus.Fields{"name": name, "ID": *project.ID}).Debugf("Found project we're looking for.\n")
	b.cachedProjects[name] = project
	return project, nil
}

func (b *Bot) getProjectColumns(project *github.Project) ([]*github.ProjectColumn, error) {
	projectColumns, _, err := b.gh.Projects.ListProjectColumns(context.Background(),
		*project.ID,
		&github.ListOptions{})

	if err != nil {
		return nil, err
	}

	b.log.Debugf("Listing columns in '%s'\n", *project.Name)
	for i, column := range projectColumns {
		b.cachedColumns[column.GetName()] = column
		b.log.Debugf("\t#%d: %s (id: %d)\n", i, *column.Name, column.ID)
	}

	return projectColumns, nil
}

func (b *Bot) getProjectColumn(project *github.Project, column string) (*github.ProjectColumn, error) {
	if column, ok := b.cachedColumns[column]; ok {
		b.log.WithField("column", column).Debugln("Using cached column")
		return column, nil
	}
	allColumns, _, err := b.gh.Projects.ListProjectColumns(context.Background(), project.GetID(), &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, col := range allColumns {
		if *col.Name == column {
			b.cachedColumns[column] = col
			return col, nil
		}
	}
	return nil, fmt.Errorf("Could not find column named '%s' in project '%s'\n", column, *project.Name)
}

func (b *Bot) getProjectCards(column *github.ProjectColumn) ([]*github.ProjectCard, error) {
	cards, _, err := b.gh.Projects.ListProjectCards(context.Background(),
		*column.ID,
		&github.ProjectCardListOptions{})

	if err != nil {
		return nil, err
	}

	b.log.Debugf("Listing cards in '%s'\n", *column.Name)
	for i, card := range cards {
		b.log.Debugf("\t#%d: ContentURL %+v\n", i, *card.ContentURL)
		b.cachedCards[*card.ID] = card
	}
	return cards, nil
}

// GetGardColumn takes a project card and finds the column it's in
func (b *Bot) GetCardColumn(card *github.ProjectCard) (*github.ProjectColumn, error) {
	if card.ColumnID == nil {
		return nil, fmt.Errorf("No columnID given on *github.ProjectCard")
	}

	column, _, err := b.gh.Projects.GetProjectColumn(context.Background(), card.GetColumnID())
	return column, err
}

func IssueNumberFromURL(url string) int {
	parts := strings.Split(url, "/")
	id, _ := strconv.Atoi(parts[len(parts)-1])
	return id
}

func (b *Bot) FindIssueProjectCard(issueNumber int) (*github.ProjectCard, error) {
	// First go through cache and see if we've cached the card.
	b.log.Debugf("Looking for card for issue with id: %d\n", issueNumber)
	for _, card := range b.cachedCards {
		b.log.Debugf("Potential card: %s\n", card.GetColumnURL())
		if IssueNumberFromURL(*card.ContentURL) == issueNumber {
			return card, nil
		}
	}

	// TODO: This needs to be rethought
	project, _ := b.getProject("Suggestions overview")
	columns, _ := b.getProjectColumns(project)
	for _, column := range columns {
		cards, _ := b.getProjectCards(column)
		for _, card := range cards {
			b.log.Debugf("Currently checking %s\n", card.GetContentURL())
			if b.IssueIDFromURL(card.GetContentURL()) == issueNumber {
				return card, nil
			}
		}
	}

	// Otherwise we'll need to go throgh all project columns

	return nil, fmt.Errorf("no card found for issue with ID='%d'\n", issueNumber)
}
