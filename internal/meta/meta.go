package meta

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Repo struct {
	Owner string
	Name  string
}

func (r Repo) Slug() string { return r.Owner + "/" + r.Name }

var (
	githubSSHRe   = regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
	githubHTTPSRe = regexp.MustCompile(`^https?://github\.com/([^/]+)/(.+?)(?:\.git)?/?$`)
)

func DetectRepo() (Repo, error) {
	url, err := gitRunner("config", "--get", "remote.origin.url")
	if err != nil || url == "" {
		return Repo{}, errors.New("no git remote.origin.url configured")
	}
	for _, re := range []*regexp.Regexp{githubSSHRe, githubHTTPSRe} {
		if m := re.FindStringSubmatch(url); m != nil {
			return Repo{Owner: m[1], Name: m[2]}, nil
		}
	}
	return Repo{}, fmt.Errorf("unrecognized git remote URL: %s", url)
}

type Author struct {
	Name  string
	Email string
}

// Lookup is the minimal viper surface Resolve needs. Pass viper.GetViper()
// in production code, or a fresh *viper.Viper in tests for isolation.
type Lookup interface {
	GetString(key string) string
}

func Resolve(v Lookup, flagName, flagEmail string) (Author, error) {
	a := Author{Name: flagName, Email: flagEmail}
	if a.Name == "" {
		a.Name = firstNonEmpty(v.GetString("author"), v.GetString("full_name"), gitConfig("user.name"))
	}
	if a.Email == "" {
		a.Email = firstNonEmpty(v.GetString("email"), gitConfig("user.email"))
	}
	if a.Name == "" {
		return a, errors.New("author is not set: pass --author, set SREKIT_AUTHOR, or configure git user.name")
	}
	if a.Email == "" {
		return a, errors.New("email is not set: pass --email, set SREKIT_EMAIL, or configure git user.email")
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
