package cmd

import (
	"context"
	"fmt"

	"github.com/netlify/gocommerce/api"
	"github.com/netlify/gocommerce/conf"
	"github.com/netlify/gocommerce/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var multiCmd = cobra.Command{
	Use:  "multi",
	Long: "Start multi-tenant API server",
	Run:  multi,
}

func multi(cmd *cobra.Command, args []string) {
	globalConfig, log, err := conf.LoadGlobal(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}
	if globalConfig.OperatorToken == "" {
		logrus.Fatal("Operator token secret is required")
	}

	db, err := models.Connect(globalConfig, log.WithField("component", "db"))
	if err != nil {
		logrus.Fatalf("Error opening database: %+v", err)
	}
	defer db.Close()

	bgDB, err := models.Connect(globalConfig, log.WithField("component", "db").WithField("bgdb", true))
	if err != nil {
		logrus.Fatalf("Error opening database: %+v", err)
	}
	defer bgDB.Close()

	globalConfig.MultiInstanceMode = true
	api := api.NewAPIWithVersion(context.Background(), globalConfig, log, db.Debug(), Version)

	l := fmt.Sprintf("%v:%v", globalConfig.API.Host, globalConfig.API.Port)
	logrus.Infof("GoCommerce API started on: %s", l)

	models.RunHooks(bgDB, logrus.WithField("component", "hooks"))

	api.ListenAndServe(l)
}
