package bot

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v33/github"
	"github.com/jadlers/botler-erwen/configuration"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	gh   *github.Client
	conf *configuration.Conf
	log  *logrus.Logger

	SyncStates []*SyncState

	// To reduce the number of requests sent
	cachedProjects map[string]*github.Project
	cachedColumns  map[string]*github.ProjectColumn
	cachedCards    map[int64]*github.ProjectCard
}

func New(conf *configuration.Conf) *Bot {
	bot := &Bot{
		log:            conf.Log,
		conf:           conf,
		cachedProjects: map[string]*github.Project{},
		cachedColumns:  map[string]*github.ProjectColumn{},
		cachedCards:    map[int64]*github.ProjectCard{},
	}
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, conf.GitHubAppID, conf.GitHubInstallationID, conf.GitHubAppPrivateKey)
	if err != nil {
		fmt.Printf("Could not initialise transport layer: %v\n", err)
		os.Exit(1)
	}

	bot.gh = github.NewClient(&http.Client{Transport: itr})

	return bot
}

// AddSyncState finds the project column and labels used to define a SyncState
// and adds it to the stored slice of SyncStates. If any part is missing
// nothing is added to the SyncState slice and an error is returned.
func (b *Bot) AddSyncState(name, projectName, columnName string, labels []string) error {
	ss := &SyncState{Name: name}

	// Find project
	project, err := b.getProject(projectName)
	if err != nil {
		b.log.Warnf("Could not find project named '%s', are you sure it exists?\n", projectName)
		return err
	}

	// Find column
	column, err := b.getProjectColumn(project, columnName)
	if err != nil {
		b.log.Warn(err)
		return err
	}
	ss.ProjectColumn = column

	// Find labels
	for _, labelName := range labels {
		label, err := b.getLabel(labelName)
		if err != nil {
			b.log.Warnf("Could not find the label '%s', does it exist in the repository?\n", labelName)
			return err
		}
		ss.Labels = append(ss.Labels, label)
	}

	// Store the filled out SyncState
	b.SyncStates = append(b.SyncStates, ss)
	b.log.WithField("Name", name).Infoln("Added new SyncState")
	return nil
}

// setupSyncStates is a temporary solution for setting up states which are
// represented by GitHub Projects and labels on issues.
//
// TODO: This should be read from some kind of file. Maybe using viper?
func (b *Bot) SetupSyncStates() {
	requiredLabels := []string{"Suggestion"} // Required for all issues in this state group
	stateNames := [5]string{"Pending", "In Consideration", "Accepted", "Rejected", "Added"}
	for _, stateName := range stateNames {
		labels := append(requiredLabels, stateName)
		b.AddSyncState(stateName, "Suggestions overview", stateName, labels)
	}
}

func (b *Bot) ConnectionStatus() (bool, error) {
	zen, _, err := b.gh.Zen(context.Background())
	if err != nil {
		return false, fmt.Errorf("Can not connect to GitHub. %s\n", err)
	}
	b.log.Debugf("Got zen for testing connectivity: '%s'\n", zen)
	return true, nil
}

func (b *Bot) RateLimitStatus() *github.Rate {
	rl, _, err := b.gh.RateLimits(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	return rl.Core
}

func (b *Bot) GetIssue(id int) (*github.Issue, error) {
	issue, _, err := b.gh.Issues.Get(context.Background(), b.conf.Owner, b.conf.Repository, id)
	if err != nil {
		b.log.Warnf("Could not get issue with ID=%d\n", id)
		return nil, fmt.Errorf("could not get issue with id=%d\n", id)
	}

	return issue, nil
}
