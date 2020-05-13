// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package common

import (
	"fmt"
	"os"
	"path"
	"strings"

	proc_config "github.com/DataDog/datadog-agent/pkg/process/config"

	"github.com/DataDog/datadog-agent/pkg/config"
)

// SetupConfig fires up the configuration system
func SetupConfig(confFilePath string) error {
	return setupConfig(confFilePath, "", false)
}

// SetupConfigWithoutSecrets fires up the configuration system without secrets support
func SetupConfigWithoutSecrets(confFilePath string, configName string) error {
	return setupConfig(confFilePath, configName, true)
}

func setupConfig(confFilePath string, configName string, withoutSecrets bool) error {
	if configName != "" {
		config.Datadog.SetConfigName(configName)
	}
	// set the paths where a config file is expected
	if len(confFilePath) != 0 {
		// if the configuration file path was supplied on the command line,
		// add that first so it's first in line
		config.Datadog.AddConfigPath(confFilePath)
		// If they set a config file directly, let's try to honor that
		if strings.HasSuffix(confFilePath, ".yaml") {
			config.Datadog.SetConfigFile(confFilePath)
		}
	}
	config.Datadog.AddConfigPath(DefaultConfPath)
	// load the configuration
	var err error
	if withoutSecrets {
		err = config.LoadWithoutSecret()
	} else {
		err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("unable to load Datadog config file: %s", err)
	}
	return nil
}

// SetupSystemProbeConfig reads the system-probe.yaml into the global config object
func SetupSystemProbeConfig(sysProbeConfFilePath string) error {
	// Open the system-probe.yaml file if it's in a custom location
	if sysProbeConfFilePath != "" {
		// If file is not set directly assume we need to add /system-probe.yaml
		if !strings.HasSuffix(sysProbeConfFilePath, ".yaml") {
			sysProbeConfFilePath = path.Join(sysProbeConfFilePath, "/system-probe.yaml")
		}
	} else {
		// Assume it is in the default location if nothing is passed in
		sysProbeConfFilePath = path.Join(DefaultConfPath, "/system-probe.yaml")
	}

	file, err := os.Open(sysProbeConfFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Merge config with an IO reader since this lets us merge the configs without changing
	// the config file set with viper
	if err := config.Datadog.MergeConfig(file); err != nil {
		return err
	}

	// The full path to the location of the unix socket where connections will be accessed
	// This is not necessarily set in the system-probe.yaml, so set it manually if it is not
	if !config.Datadog.IsSet("system_probe_config.sysprobe_socket") {
		config.Datadog.Set("system_probe_config.sysprobe_socket", proc_config.GetSocketPath())
	}

	// Load the env vars last to overwrite values
	proc_config.LoadSysProbeEnvVariables()
	return nil
}
