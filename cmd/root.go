/*
Copyright © 2023 Vishal Tewatia <tewatiavishal3@gmail.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lunny/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vishu42/github-token/pkg"
)

var (
	cfgFile           string
	AppID             int64
	AppInstallationID int64
	AppPrivateKey     string
)

const (
	tokenFile = "/tmp/githubtoken/token.txt"
	expFile   = "/tmp/githubtoken/token-expiry.txt"
)

func ensureDir(dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "github-token",
	Short: "Get a GitHub token for a GitHub App",
	Run: func(cmd *cobra.Command, args []string) {
		// make sure we have all the required flags

		// get flags from viper
		AppID = viper.GetInt64("app-id")
		AppInstallationID = viper.GetInt64("app-installation-id")
		AppPrivateKey = viper.GetString("app-private-key")

		// make sure we have all the required flags
		if AppID == 0 || AppInstallationID == 0 || AppPrivateKey == "" {
			fmt.Fprintln(os.Stderr, "app-id, app-installation-id and app-private-key are required")
			os.Exit(1)
		}

		// initialize GitHubAppAuth struct
		appAuth := &pkg.GitHubAppAuth{
			AppID:             AppID,
			AppInstallationID: AppInstallationID,
			AppPrivateKey:     AppPrivateKey,
		}

		// get token from file
		err := ensureDir(filepath.Dir(tokenFile))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		file, err := os.OpenFile(tokenFile, os.O_RDONLY|os.O_CREATE, 0o666)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer file.Close()

		sc := bufio.NewScanner(file)
		sc.Scan()
		tokenStr := sc.Text()

		// get exp from file
		err = ensureDir(filepath.Dir(expFile))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		file, err = os.OpenFile(expFile, os.O_RDONLY|os.O_CREATE, 0o666)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer file.Close()

		sc = bufio.NewScanner(file)
		sc.Scan()
		githubTokenExp := sc.Text()

		var token *pkg.AccessToken

		// if token is empty -> fetch the token
		// if token is not empty
		//     parse the expiry and if token is about to expire or expired -> fetch the token
		// use the token

		// if token is empty, fetch the token
		if tokenStr == "" {
			// get the token
			t, err := pkg.FetchAccessToken(cmd.Context(), "https://api.github.com", appAuth)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			// set the token
			token = t

		}

		// token must be non-empty now
		// check if the github token exsits
		if githubTokenExp != "" {
			// parse the expiry
			exp, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", githubTokenExp)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			// if the token is expired or about to expire, fetch a new token
			if time.Now().After(exp) || time.Now().Add(5*time.Minute).After(exp) {
				// get the token
				log.Println("Token is expired or about to expire, fetching a new token")
				t, err := pkg.FetchAccessToken(cmd.Context(), "https://api.github.com", appAuth)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}

				token = t

			}
		}

		// set the token if it is not set
		if token == nil {
			exp, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", githubTokenExp)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			token = &pkg.AccessToken{
				Token:     tokenStr,
				ExpiresAt: exp,
			}
		}

		// export the token in a file
		err = os.WriteFile(tokenFile, []byte(token.Token), 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// export the token expiry in a file
		err = os.WriteFile(expFile, []byte(token.ExpiresAt.String()), 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// print the token
		fmt.Println(token.Token)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.github-token.yaml)")
	rootCmd.PersistentFlags().StringVar(&AppPrivateKey, "app-private-key", "", "GitHub App private key")
	rootCmd.PersistentFlags().Int64Var(&AppID, "app-id", 0, "GitHub App ID")
	rootCmd.PersistentFlags().Int64Var(&AppInstallationID, "app-installation-id", 0, "GitHub App installation ID")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
