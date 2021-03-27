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

			issueNumber := event.Issue.GetNumber()
			eventLog := log.WithFields(logrus.Fields{
				"action":   event.GetAction(),
				"issueNum": issueNumber,
				"labels":   labels,
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
			if card, err := erwen.FindIssueProjectCard(issueNumber); err == nil {
				if err := erwen.MoveCard(matchedState, card.GetID()); err != nil {
					eventLog.Errorf("Error moving card", err)
				} else {
					eventLog.Infof("Moved card to column: %s\n", matchedState.ProjectColumn.GetName())
				}
			} else {
				// Create the card
				ghIssue, _ := erwen.GetIssue(issueNumber)
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

			column, err := erwen.GetCardColumn(event.ProjectCard)
			if err != nil {
				eventLog.Errorln(err)
				return
			}

			columnName := column.GetName()
			eventLog.Debugf("Looking for column name '%s'\n", columnName)

			// Get the SyncState based on the column
			var newState *bot.SyncState
			for _, ss := range erwen.SyncStates {
				if columnName == ss.Name {
					newState = ss
					break
				}
			}

			if newState == nil {
				eventLog.Infoln("Project card not in a synced column")
				return
			}

			// Check if the required state labels are set
			issueNumber := bot.IssueNumberFromURL(event.ProjectCard.GetContentURL())
			issue, err := erwen.GetIssue(issueNumber)
			if err != nil {
				eventLog.Errorf("Could not find issue for Project Card with content URL: %s\n", event.ProjectCard.GetColumnURL())
				return
			}

			labels, changed := erwen.GetCorrectLabels(newState, issue)
			if !changed {
				eventLog.Infoln("Project card already has correct labels")
				return
			}

			// 3. If not set them
			if err := erwen.SetIssueLabels(issue, labels); err != nil {
				eventLog.Errorf("Failed to set labels on issue: %v\n", err)
			}

			eventLog.Infoln("Labels for issue updated")

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
