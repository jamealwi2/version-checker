package main

import (
	utils "deploy-version-checker/utils"
	"go.uber.org/zap"
	"os"
	"strings"
)

var (
	sugar      *zap.SugaredLogger
	namespaces []string
	input      string
)

func main() {
	sugar.Info("Reading deployment groups.")
	g := utils.NewGroups()
	err := g.ReadGroups(input)
	if err != nil {
		sugar.Fatalf("failed to read deployment groups: %v", err)
	}
	details := make(map[string]string)
	for _, ns := range namespaces {
		details, err = utils.GetAppDetails(ns)
		if err != nil {
			sugar.Fatalf("failed to get app details: %v", err)
		}
	}
	g.CheckContainerImage(utils.FindMismatch, details)
}

func init() {
	sugar = utils.InitializeLogger().Sugar()
	if os.Getenv(utils.NAMESPACES) != "" {
		namespaces = strings.Split(os.Getenv(utils.NAMESPACES), ",")
	} else {
		namespaces = []string{utils.DEFAULT_NAMESPACE}
	}
	if os.Getenv("groupsConfig") != "" {
		input = os.Getenv("groupsConfig")
		sugar.Infof("Using groups config from environment variable.")
	} else {
		input = utils.InputJSON
	}
}
