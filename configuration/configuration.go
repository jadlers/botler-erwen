package configuration

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Conf struct {
	Owner      string
	Repository string

	GitHubAppID          int64
	GitHubAppPrivateKey  string
	GitHubInstallationID int64
	GitHubWebhookSecret  string

	syncStates []syncState
}

type syncState struct {
	Name          string
	ProjectName   string
	ProjectColumn string
	labelNames    []string
}

// Init will collect configuration into a Configuration
func Init() *Conf {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading '.env' file")
	} else {
		log.Println("Reading environment from '.env' file")
	}

	stateNames := [5]string{"Pending", "In Consideration", "Accepted", "Rejected", "Added"}
	var syncStates []syncState
	for _, stateName := range stateNames {
		syncStates = append(syncStates, syncState{
			Name:          stateName,
			ProjectName:   "Suggestions overview",
			ProjectColumn: stateName,
			labelNames:    []string{"Suggestion", stateName},
		})
	}

	conf := &Conf{
		Owner:               "jadlers",
		Repository:          "webhook-testing-TMP",
		GitHubAppPrivateKey: os.Getenv("GITHUB_PRIVATE_KEY"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		syncStates:          syncStates,
	}

	appID, err := getNumericEnv("GITHUB_APP_ID", true)
	if err != nil {
		log.Fatal(err)
	}
	conf.GitHubAppID = int64(appID)

	installationID, err := getNumericEnv("GITHUB_INSTALLATION_ID", true)
	if err != nil {
		log.Fatal(err)
	}
	conf.GitHubInstallationID = int64(installationID)

	return conf
}

func getNumericEnv(variable string, require bool) (int, error) {
	val := os.Getenv(variable)
	num, err := strconv.Atoi(val) // Defauts to 0 if errors
	if require && err != nil {
		return num, fmt.Errorf("Could not cast required environment '%s'='%s' to an integer\n", variable, val)
	}
	return num, nil
}