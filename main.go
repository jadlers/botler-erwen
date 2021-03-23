package main

import (
	"context"
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
	path = "/webhooks"
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
	logDebugln("Rate limits:", rl)

	hook, err := webhook.New(webhook.Options.Secret(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Fatalf("Error setting up webhook: %+v\n", err)
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

func handleIssuesEvent(issue webhook.IssuesPayload) {
	if issue.Action != "labeled" && issue.Action != "unlabeled" {
		logDebugln("Ignoring issue event:", issue.Action)
		return
	}

	logDebugln("New IssueEvent:")
	logDebugf("  Action: %+v\n", issue.Action)
	logDebugf("  Label: %+v\n", issue.Label.Name)
}

func handleProjectCardEvent(projectCard webhook.ProjectCardPayload) {
	logDebugln("New ProjectCardEvent:")
	logDebugf("  %+v\n", projectCard.Action)
	logDebugf("  %+v\n", projectCard.ProjectCard)
}
