package main

import (
	"net/http"
	"os"

	"github.com/openshift/assisted-installer/src/config"
	"github.com/openshift/assisted-installer/src/installer"
	"github.com/openshift/assisted-installer/src/inventory_client"
	"github.com/openshift/assisted-installer/src/k8s_client"
	"github.com/openshift/assisted-installer/src/ops"
	"github.com/openshift/assisted-installer/src/utils"
	"github.com/openshift/assisted-service/models"
	"github.com/openshift/assisted-service/pkg/secretdump"
)

func main() {
	config.ProcessArgs()
	logger := utils.InitLogger(config.GlobalConfig.Verbose, true)
	config.GlobalConfig.PullSecretToken = os.Getenv("PULL_SECRET_TOKEN")
	if config.GlobalConfig.PullSecretToken == "" {
		logger.Warnf("Missing Pull Secret Token environment variable")
	}

	logger.Infof("Assisted installer started. Configuration is:\n %s", secretdump.DumpSecretStruct(config.GlobalConfig))
	client, err := inventory_client.CreateInventoryClient(config.GlobalConfig.ClusterID, config.GlobalConfig.URL,
		config.GlobalConfig.PullSecretToken, config.GlobalConfig.SkipCertVerification, config.GlobalConfig.CACertPath, logger, http.ProxyFromEnvironment)
	if err != nil {
		logger.Fatalf("Failed to create inventory client %e", err)
	}

	ai := installer.NewAssistedInstaller(logger,
		config.GlobalConfig,
		ops.NewOps(logger, true),
		client,
		k8s_client.NewK8SClient,
	)
	if err := ai.InstallNode(); err != nil {
		ai.UpdateHostInstallProgress(models.HostStageFailed, err.Error())
		os.Exit(1)
	}
}
