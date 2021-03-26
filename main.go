package main

import (
	"net/http"

	"github.com/google/go-github/v33/github"
	"github.com/jadlers/botler-erwen/bot"
	"github.com/jadlers/botler-erwen/configuration"
	"github.com/sirupsen/logrus"
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

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, []byte(conf.GitHubWebhookSecret))
		if err != nil {
			log.Infoln("Ignoring unvalidated request")
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			log.Errorln(err)
		}

		switch event := event.(type) {
		case *github.IssuesEvent:
			if selfTriggered(event.Sender) {
				log.Debugln("Ignoring 'IssuesEvent' since triggered by this bot")
				return
			}

			var labels []string
			for _, label := range event.Issue.Labels {
				labels = append(labels, label.GetName())
			}

			issueID := erwen.IssueIDFromURL(event.Issue.GetURL())
			eventLog := log.WithFields(logrus.Fields{
				"action":  event.GetAction(),
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
				eventLog.Infoln("No SyncState matches labels on issue")
				return
			} else if len(matchedStates) > 1 {
				eventLog.Warnln("Issue matches multiple SyncStates")
				return
			}

			matchedState := matchedStates[0]
			eventLog.Debugf("Found single matching state: %s\n", matchedState.Name)

			// Check if it's in the correct project column
			if erwen.IsInCorrectColumn(matchedState, event.Issue.GetURL()) {
				eventLog.Debugln("Issue is already in correct synced state")
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
				ghIssue, _ := erwen.GetIssue(issueID)
				if _, err := erwen.CreateCard(matchedState, ghIssue); err != nil {
					eventLog.Warnf("Error creating card", err)
				} else {
					eventLog.Infoln("Created card.")
				}
			}

		case *github.ProjectCardEvent:
			if selfTriggered(event.Sender) {
				log.Debugln("Ignoring 'ProjectCardEvent' since triggered by this bot")
				return
			}

			projectCard := event
			eventLog := log.WithField("action", projectCard.GetAction())
			eventLog.Infoln("New ProjectCardEvent")
			eventLog.Errorln("NOT IMPLEMENTED")

		default:
			log.WithField("eventType", github.WebHookType(r)).Debugln("New unhandled event")
		}
	})

	http.ListenAndServe(":3000", nil)
}

// selfTriggered returns true if the passed user is this bot
func selfTriggered(user *github.User) bool {
	if user.GetType() == "Bot" && user.GetLogin() == "botler-erwen[bot]" {
		return true
	}
	return false
}
