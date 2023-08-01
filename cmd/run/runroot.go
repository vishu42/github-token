package run

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/vishu42/github-token/pkg"
)

const (
	tempDir   = "/tmp/githubtoken"
	tokenFile = "token.txt"
	expFile   = "token-expiry.txt"
	hashFile  = "hash.txt"
)

type RootFlags struct {
	AppID             int64
	AppInstallationID int64
	AppPrivateKey     string
	ForceRefresh      bool
	Debug             bool
}

func ensureDir(dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}
	return nil
}

func isTokenAboutToExpire(t string) bool {
	if t == "" {
		return true
	}

	ce, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", t)
	if err != nil {
		log.Fatalf("failed to parse expiry: %v", err)
	}

	// if the token is expired or about to expire, fetch a new token
	if time.Now().After(ce) || time.Now().Add(5*time.Minute).After(ce) {
		return true
	}

	return false
}

func RunRootFunc(cmd *cobra.Command, log pkg.Logger, o *RootFlags) {
	// make sure we have all the required flags
	if o.AppID == 0 || o.AppInstallationID == 0 || o.AppPrivateKey == "" {
		log.Fatalf("missing required flags - app-id: %d, app-installation-id: %d, app-private-key: %s", o.AppID, o.AppInstallationID, o.AppPrivateKey)
	}

	// initialize GitHubAppAuth struct
	appAuth := &pkg.GitHubAppAuth{
		AppID:             o.AppID,
		AppInstallationID: o.AppInstallationID,
		AppPrivateKey:     o.AppPrivateKey,
	}

	// create token dir
	err := ensureDir(tempDir)
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}

	// get token from file
	ct, err := pkg.ReadFile(filepath.Join(tempDir, tokenFile))
	if err != nil {
		log.Fatalf("failed to read token file: %v", err)
	}

	// get exp from file and parse it
	ce, err := pkg.ReadFile(filepath.Join(tempDir, expFile))
	if err != nil {
		log.Fatalf("failed to read exp file: %v", err)
	}

	// get hash from file
	ch, err := pkg.ReadFile(filepath.Join(tempDir, hashFile))
	if err != nil {
		log.Fatalf("failed to read hash file: %v", err)
	}

	// create new hash
	nh, err := pkg.CreateHash(o.AppPrivateKey)
	if err != nil {
		log.Fatalf("failed to create hash: %v", err)
	}

	// debug logs
	if ch != nh {
		log.Debugf("hash changed: %s -> %s", ch, nh)
	}
	if ct == "" {
		log.Debugf("token is empty")
	}
	if ce == "" {
		log.Debugf("exp is empty")
	}
	if isTokenAboutToExpire(ce) {
		log.Debugf("token is about to expire")
	}

	if ch != nh || isTokenAboutToExpire(ce) || o.ForceRefresh || ct == "" {
		// fetch new token
		log.Infof("fetching new token")
		t, err := pkg.FetchAccessToken(cmd.Context(), "https://api.github.com", appAuth)
		if err != nil {
			log.Fatalf("failed to fetch access token: %v", err)
		}

		// write token to file
		err = os.WriteFile(filepath.Join(tempDir, tokenFile), []byte(t.Token), 0o644)
		if err != nil {
			log.Fatalf("failed to write token file: %v", err)
		}

		// write exp to file
		err = os.WriteFile(filepath.Join(tempDir, expFile), []byte(t.ExpiresAt.String()), 0o644)
		if err != nil {
			log.Fatalf("failed to write exp file: %v", err)
		}

		log.Infof(t.Token)
	} else {
		log.Infof("using existing token")
		log.Infof(ct)
	}

	if nh != ch {
		// write hash to file
		err = os.WriteFile(filepath.Join(tempDir, hashFile), []byte(nh), 0o644)
		if err != nil {
			log.Fatalf("failed to write hash file: %v", err)
		}
	}
}

func RunRoot(o *RootFlags) func(*cobra.Command, []string) {
	// create a logger
	l, err := pkg.New(o.Debug)
	if err != nil {
		l.Fatalf("failed to create logger: %v", err)
	}

	return func(c *cobra.Command, s []string) {
		RunRootFunc(c, l, o)
	}
}
