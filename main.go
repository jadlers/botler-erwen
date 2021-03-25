package main

import (
	"net/http"
	"os"

	"github.com/jadlers/botler-erwen/bot"
	"github.com/jadlers/botler-erwen/configuration"
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
			log.WithField("action", issue.Action).Infoln("Got new IssuesEvent")

		case webhook.ProjectCardPayload:
			projectCard := payload.(webhook.ProjectCardPayload)
			log.WithField("action", projectCard.Action).Infoln("Got new ProjectCardEvent")
		}
	})

	if conf.Environment == configuration.Production {
		http.ListenAndServe(":3000", nil)
	}
}
