package bot

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v33/github"
	"github.com/jadlers/botler-erwen/configuration"
	webhook "gopkg.in/go-playground/webhooks.v5/github"
)

type Bot struct {
	gh   *github.Client
	hook *webhook.Webhook
	conf *configuration.Conf
}

func New(conf *configuration.Conf) *Bot {
	bot := &Bot{}
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, conf.GitHubAppID, conf.GitHubInstallationID, conf.GitHubAppPrivateKey)
	if err != nil {
		fmt.Printf("Could not initialise transport layer: %v\n", err)
		os.Exit(1)
	}

	bot.gh = github.NewClient(&http.Client{Transport: itr})

	return bot
}

func (b *Bot) ConnectionStatus() (bool, error) {
	zen, _, err := b.gh.Zen(context.Background())
	if err != nil {
		return false, fmt.Errorf("Can not connect to GitHub. %s\n", err)
	}
	fmt.Printf("[INFO] Got zen for testing connectivity: '%s'\n", zen)
	return true, nil
}

func (b *Bot) RateLimitStatus() *github.Rate {
	rl, _, err := b.gh.RateLimits(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	return rl.Core
}