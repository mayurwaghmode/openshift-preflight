package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFileUsed bool

func resultsFilenameWithExtension(ext string) string {
	return strings.Join([]string{"results", ext}, ".")
}

func initConfig() {
	// set up ENV var support
	viper.SetEnvPrefix("pflt")
	viper.AutomaticEnv()

	// set up optional config file support
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	configFileUsed = true
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configFileUsed = false
		}
	}

	// Set up logging config defaults
	viper.SetDefault("logfile", DefaultLogFile)
	viper.SetDefault("loglevel", DefaultLogLevel)
	viper.SetDefault("artifacts", DefaultArtifactsDir)

	// Set up cluster defaults
	viper.SetDefault("namespace", DefaultNamespace)
	viper.SetDefault("serviceaccount", DefaultServiceAccount)

	// Set up scorecard wait time default
	viper.SetDefault("scorecard_wait_time", DefaultScorecardWaitTime)
}

// preRunConfig is used by cobra.PreRun in all non-root commands to load all necessary configurations
func preRunConfig(cmd *cobra.Command, args []string) {
	// set up logging
	logname := viper.GetString("logfile")
	logFile, err := os.OpenFile(logname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err == nil {
		mw := io.MultiWriter(os.Stderr, logFile)
		log.SetOutput(mw)
	} else {
		log.Debug("Failed to log to file, using default stderr")
	}
	if ll, err := log.ParseLevel(viper.GetString("loglevel")); err == nil {
		log.SetLevel(ll)
	}

	log.SetFormatter(&log.TextFormatter{})
	if !configFileUsed {
		log.Debug("config file not found, proceeding without it")
	}
}

func buildConnectURL(projectID string) string {
	connectURL := fmt.Sprintf("https://connect.redhat.com/projects/%s", projectID)

	if viper.GetString("pyxis_env") != "prod" {
		connectURL = fmt.Sprintf("https://connect.%s.redhat.com/projects/%s", viper.GetString("pyxis_env"), projectID)
	}

	return connectURL
}

func buildOverviewURL(projectID string) string {
	return fmt.Sprintf("%s/overview", buildConnectURL(projectID))
}

func buildScanResultsURL(projectID string, imageID string) string {
	return fmt.Sprintf("%s/images/%s/scan-results", buildConnectURL(projectID), imageID)
}

func convertPassedOverall(passedOverall bool) string {
	if passedOverall {
		return "PASSED"
	}

	return "FAILED"
}
