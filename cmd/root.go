/*
Copyright © 2023 Vishal Tewatia <tewatiavishal3@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vishu42/github-token/cmd/run"
)

var cfgFile string

func CreateRootCmd() *cobra.Command {
	cobra.OnInitialize(initConfig)

	o := &run.RootFlags{}

	rootCmd := &cobra.Command{
		Use:   "github-token",
		Short: "Get a GitHub token for a GitHub App",
		Run:   run.RunRoot(o),
		PreRun: func(cmd *cobra.Command, args []string) {
			o.AppID = viper.GetInt64("app-id")
			o.AppInstallationID = viper.GetInt64("app-installation-id")
			o.AppPrivateKey = viper.GetString("app-private-key")
			o.ForceRefresh = viper.GetBool("force-refresh")
			o.Debug = viper.GetBool("debug")
		},
	}

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.github-token.yaml)")
	rootCmd.PersistentFlags().String("app-private-key", "", "GitHub App private key")
	rootCmd.PersistentFlags().Int64("app-id", 0, "GitHub App ID")
	rootCmd.PersistentFlags().Int64("app-installation-id", 0, "GitHub App installation ID")
	rootCmd.PersistentFlags().Bool("force-refresh", false, "Force refresh token")
	rootCmd.PersistentFlags().Bool("debug", true, "Debug mode")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := CreateRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".github-token" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".github-token")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
