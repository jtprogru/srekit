package meta

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

type Author struct {
	Name  string
	Email string
}

func Resolve(flagName, flagEmail string) (Author, error) {
	a := Author{Name: flagName, Email: flagEmail}
	if a.Name == "" {
		a.Name = firstNonEmpty(viper.GetString("author"), viper.GetString("full_name"), gitConfig("user.name"))
	}
	if a.Email == "" {
		a.Email = firstNonEmpty(viper.GetString("email"), gitConfig("user.email"))
	}
	if a.Name == "" {
		return a, fmt.Errorf("author is not set: pass --author, set SREKIT_AUTHOR, or configure git user.name")
	}
	if a.Email == "" {
		return a, fmt.Errorf("email is not set: pass --email, set SREKIT_EMAIL, or configure git user.email")
	}
	return a, nil
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if s := strings.TrimSpace(x); s != "" {
			return s
		}
	}
	return ""
}

var gitRunner = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func gitConfig(key string) string {
	out, err := gitRunner("config", "--get", key)
	if err != nil {
		return ""
	}
	return out
}
