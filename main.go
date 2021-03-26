package main

import (
	"net/http"
	"os"

	"github.com/jadlers/botler-erwen/bot"
	"github.com/jadlers/botler-erwen/configuration"
	"github.com/sirupsen/logrus"
	webhook "gopkg.in/go-playground/webhooks.v5/github"
)

func main() {
	conf := configuration.Init()
	log := conf.Log

	erwen := bot.New(conf)
	if conf.Environment == configuration.Production {
		// Don't fetch from GitHub while developing
		if zen, err := erwen.ConnectionStatus(); err != nil {
			log.Warn(err)
		} else {
			log.Debugf("GitHub connection works, got zen: %s\n", zen)
		}
	}

	erwen.SetupSyncStates()

	hook, err := webhook.New(webhook.Options.Secret(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Errorf("Could not set up webhook: %v\n", err)
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, webhook.IssuesEvent, webhook.LabelEvent, webhook.ProjectCardEvent)
		if err != nil {
			if err == webhook.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
			}
		}

		switch payload.(type) {
		case webhook.IssuesPayload:
			issue := payload.(webhook.IssuesPayload)
			var labels []string
			for _, label := range issue.Issue.Labels {
				labels = append(labels, label.Name)
			}
			issueID := erwen.IssueIDFromURL(issue.Issue.URL)
			eventLog := log.WithFields(logrus.Fields{
				"action":  issue.Action,
				"issueID": issueID,
				"labels":  labels,
			})
			eventLog.Infoln("New issue event")

			// Check if the labels on the issue match labels of any SyncState
			var matchedStates []*bot.SyncState
			for _, syncState := range erwen.SyncStates {
				if syncState.InState(labels) {
					matchedStates = append(matchedStates, syncState)
					break
				}
			}

			// See that there's only one match
			if len(matchedStates) == 0 {
				eventLog.Debugln("No SyncState matches labels on issue")
				return
			} else if len(matchedStates) > 1 {
				eventLog.Warnln("The issue matches multiple SyncStates")
				return
			}

			matchedState := matchedStates[0]
			eventLog.Debugf("Found single matching state: %s\n", matchedState.Name)

			// Check if it's in the correct project column
			if erwen.IsInCorrectColumn(matchedState, issue.Issue.URL) {
				eventLog.Debugln("Issue is in the synced state")
				return
			}

			// If not: Find it and move it or create it.
			if card, err := erwen.FindIssueProjectCard(issueID); err == nil {
				if err := erwen.MoveCard(matchedState, card.GetID()); err != nil {
					eventLog.Errorf("Error moving card", err)
				} else {
					eventLog.Infof("Moved card to column: %s\n", matchedState.ProjectColumn.GetName())
				}
			} else {
				// Create the card
				log.WithField("issueID", issueID).Fatalln("Could not find a card, creating one.")
			}

		case webhook.ProjectCardPayload:
			projectCard := payload.(webhook.ProjectCardPayload)
			log.WithField("action", projectCard.Action).Infoln("New ProjectCardEvent")

		case webhook.LabelPayload:
			label := payload.(webhook.LabelPayload)
			log.WithField("action", label.Action).Infoln("New LabelEvent")
		}
	})

	http.ListenAndServe(":3000", nil)
}
