package configuration

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Environment string

const (
	Development = "DEVELOPMENT"
	Testing     = "TESTING"
	Production  = "PRODUCTION"
)

type Conf struct {
	Environment Environment
	Log         *logrus.Logger

	Owner      string
	Repository string

	GitHubAppID          int64
	GitHubAppPrivateKey  string
	GitHubInstallationID int64
	GitHubWebhookSecret  string
}

// Init will collect configuration into a Configuration
func Init() *Conf {
	log := logrus.New()
	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading '.env' file")
	} else {
		log.Info("Reading environment from '.env' file")
	}

	conf := &Conf{
		Owner:               "jadlers",
		Repository:          "webhook-testing-TMP",
		GitHubAppPrivateKey: os.Getenv("GITHUB_PRIVATE_KEY"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}

	switch os.Getenv("ENV") {
	case "testing":
		conf.Environment = Testing
		log.SetLevel(logrus.WarnLevel)
	case "production":
		conf.Environment = Production
		log.SetLevel(logrus.WarnLevel)
	default:
		log.Infof("Using Environment='%s' since none set\n", Development)
		fallthrough
	case "development":
		conf.Environment = Development
		log.SetLevel(logrus.DebugLevel)
	}

	conf.Log = log

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
