package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v33/github"
	"github.com/joho/godotenv"
	webhook "gopkg.in/go-playground/webhooks.v5/github"
)

const (
	path        = "/webhooks"
	repoOwner   = "jadlers"
	repoName    = "webhook-testing-TMP"
	projectName = "Suggestions overview"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ghClient := createGithubClient()

	rl, _, err := ghClient.RateLimits(context.Background())
	if err != nil {
		log.Panicln(err)
		os.Exit(1)
	}
	logDebugf("Rate limits: %d/%d remaining", rl.Core.Remaining, rl.Core.Limit)

	hook, err := webhook.New(webhook.Options.Secret(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Fatalf("Error setting up webhook: %+v\n", err)
	}

	project := getProject(ghClient)
	// columns := getProjectColumns(ghClient, project)
	// for _, column := range columns {
	// 	getProjectCards(ghClient, column)
	// }

	logDebugln("Looking for card for issue 1")
	foundCard, err := FindIssueProjectCard(ghClient, project, 1)
	if err != nil {
		log.Fatalln(err)
	} else {
		logDebugf("\tFound card %+v\n", foundCard)
	}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, webhook.IssuesEvent, webhook.LabelEvent, webhook.ProjectCardEvent)
		if err != nil {
			if err == webhook.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
			}
		}

		switch payload.(type) {
		case webhook.IssuesPayload:
			issue := payload.(webhook.IssuesPayload)
			handleIssuesEvent(issue)

		case webhook.ProjectCardPayload:
			projectCard := payload.(webhook.ProjectCardPayload)
			handleProjectCardEvent(projectCard)
		}
	})
	http.ListenAndServe(":3000", nil)
}

func createGithubClient() *github.Client {
	appIdEnv, exist := os.LookupEnv("GITHUB_APP_IDENTIFIER")
	if !exist {
		logInfoln("'GITHUB_APP_IDENTIFIER' not set")
	} else {
		logDebugf("Using 'app_id': %s", appIdEnv)
	}
	appID, err := strconv.ParseInt(appIdEnv, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}

	installationIDEnv, exist := os.LookupEnv("GITHUB_INSTALLATION_IDENTIFIER")
	if !exist {
		logInfoln("'GITHUB_INSTALLATION_IDENTIFIER' not set")
	} else {
		logDebugf("Using 'installation_id': %s", appIdEnv)
	}
	installationID, err := strconv.ParseInt(installationIDEnv, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}

	privateKeyFile, exist := os.LookupEnv("GITHUB_PRIVATE_KEY")
	if !exist {
		logInfoln("'PRIVATE_KEY_FILE' not set")
	} else {
		logDebugf("Using private key file: %s\n", privateKeyFile)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyFile)
	if err != nil {
		log.Fatalf("Error creating transport layer %+v\n", err)
	}

	return github.NewClient(&http.Client{Transport: itr})
}

func getProject(gh *github.Client) *github.Project {
	projects, _, err := gh.Repositories.ListProjects(context.Background(), repoOwner, repoName, &github.ProjectListOptions{State: "open"})
	if err != nil {
		log.Fatalln("Could not get repo", err)
	}

	logDebugln("Listing projects in the repo:")
	var project *github.Project
	for i, p := range projects {
		logDebugf("\t#%d: %s (id: %d)\n", i, *p.Name, p.ID)
		if *p.Name == projectName {
			project = p
			break
		}
	}

	if project == nil {
		logInfof("Could not find project named %s\n", projectName)
	} else {
		logDebugln("Found project we're looking for:", project)
	}

	return project
}

func getProjectColumns(gh *github.Client, project *github.Project) []*github.ProjectColumn {
	projectColumns, _, err := gh.Projects.ListProjectColumns(context.Background(), *project.ID, &github.ListOptions{})
	if err != nil {
		log.Fatal("Could not get project columns:", err)
	}
	logDebugf("Listing columns in '%s'\n", *project.Name)
	for i, column := range projectColumns {
		logDebugf("\t#%d: %s (id: %d)\n", i, *column.Name, column.ID)
	}

	return projectColumns
}

func getProjectCards(gh *github.Client, column *github.ProjectColumn) []*github.ProjectCard {
	cards, _, err := gh.Projects.ListProjectCards(context.Background(), *column.ID, &github.ProjectCardListOptions{})
	if err != nil {
		log.Fatalf("Could not fetch cards for column '%s':\n%+v\n", *column.Name, err)
	}

	logDebugf("Listing cards in '%s'\n", *column.Name)
	for i, card := range cards {
		logDebugf("\t#%d: ContentURL %+v\n", i, *card.ContentURL)
	}
	return cards
}

func FindIssueProjectCard(gh *github.Client, project *github.Project, issueNum int) (projectCard *github.ProjectCard, err error) {
	issueURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", repoOwner, repoName, issueNum)
	logDebugf("Looking for card with URL: %s\n", issueURL)
	columns := getProjectColumns(gh, project)
	for _, column := range columns {
		cards := getProjectCards(gh, column)
		for _, card := range cards {
			if issueURL == *card.ContentURL {
				return card, nil
			}
		}
	}

	return nil, fmt.Errorf("No project card found for issue %d", issueNum)
}

func handleIssuesEvent(issue webhook.IssuesPayload) {
	if issue.Action != "labeled" && issue.Action != "unlabeled" {
		logDebugln("Ignoring issue event:", issue.Action)
		return
	}

	logDebugln("New IssueEvent:")
	logDebugf("\tAction: %+v\n", issue.Action)
	logDebugf("\tLabel: %+v\n", issue.Label.Name)
}

func handleProjectCardEvent(projectCard webhook.ProjectCardPayload) {
	logDebugln("New ProjectCardEvent:")
	logDebugf("\t%+v\n", projectCard.Action)
	logDebugf("\t%+v\n", projectCard.ProjectCard)
}
