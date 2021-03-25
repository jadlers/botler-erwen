package bot

import (
	"context"
	"fmt"

	"github.com/google/go-github/v33/github"
)

func (b *Bot) getProject(name string) (*github.Project, error) {
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

	b.log.Debugf("Found project we're looking for: %+v\n", project)
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
		b.log.Debugf("\t#%d: %s (id: %d)\n", i, *column.Name, column.ID)
	}

	return projectColumns, nil
}

func (b *Bot) getProjectColumn(project *github.Project, column string) (*github.ProjectColumn, error) {
	allColumns, _, err := b.gh.Projects.ListProjectColumns(context.Background(), project.GetID(), &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, col := range allColumns {
		if *col.Name == column {
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
	}
	return cards, nil
}
