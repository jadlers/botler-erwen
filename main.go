package main

import (
	"fmt"

	"net/http"

	"gopkg.in/go-playground/webhooks.v5/github"
)

const (
	path = "/webhooks"
)

func main() {
	hook, _ := github.New()

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.IssuesEvent, github.LabelEvent, github.ProjectCardEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
			}
		}

		switch payload.(type) {
		case github.IssuesPayload:
			issue := payload.(github.IssuesPayload)

			if issue.Action != "labeled" && issue.Action != "unlabeled" {
				fmt.Println("Ignoring issue event:", issue.Action)
				return
			}

			fmt.Println("New IssueEvent:")
			fmt.Printf("  Action: %+v\n", issue.Action)
			fmt.Printf("  Label: %+v\n", issue.Label.Name)

		case github.ProjectCardPayload:
			projectCard := payload.(github.ProjectCardPayload)

			fmt.Println("New ProjectCardEvent:")
			fmt.Printf("  %+v\n", projectCard.Action)
			fmt.Printf("  %+v\n", projectCard.ProjectCard)
		}
	})
	http.ListenAndServe(":3000", nil)
}
