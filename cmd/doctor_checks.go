package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/config"
	"github.com/jtprogru/srekit/internal/meta"
)

// doctorEnv is the resolved state the checks report on, computed once per
// invocation from the same helpers every other command calls. Resolution
// errors are carried rather than returned so the check that owns one can
// report it as its own finding.
type doctorEnv struct {
	cmd *cobra.Command

	configPath     string // the config file that will actually be read
	configExplicit bool   // --config was given, bypassing location precedence
	configPathErr  error

	templatesDir    string // resolved templates dir; "" when none is configured
	templatesSource string // what supplied it; "" when nothing did
	templatesErr    error
}

func newDoctorEnv(cmd *cobra.Command) *doctorEnv {
	env := &doctorEnv{cmd: cmd}
	env.configPath, env.configExplicit, env.configPathErr = doctorConfigPath(cmd)
	env.templatesDir, env.templatesSource, env.templatesErr = doctorTemplatesDir(cmd)
	return env
}

// usableTemplatesDir returns the configured templates directory when it
// exists and is a directory. The template health checks only have something
// to inspect in that case; otherwise generation is embedded-only.
func (e *doctorEnv) usableTemplatesDir() (string, bool) {
	if e.templatesErr != nil || e.templatesDir == "" {
		return "", false
	}
	if !dirExists(e.templatesDir) {
		return "", false
	}
	return e.templatesDir, true
}

// doctorConfigPath resolves the config file the CLI will read, mirroring
// configTargetPath: an explicit --config wins, otherwise the legacy path if
// it exists, otherwise the XDG path.
func doctorConfigPath(cmd *cobra.Command) (string, bool, error) {
	if f := cmd.Root().PersistentFlags().Lookup("config"); f != nil {
		if v := f.Value.String(); v != "" {
			path, err := expandHome(v)
			return path, true, err
		}
	}
	return resolveConfigPath(), false, nil
}

// doctorTemplatesDir resolves the templates directory the same way
// resolveTemplatesDir does, but splits the sources apart so the finding can
// name which one supplied the value. resolveTemplatesDir folds the
// environment and the config file into a single config lookup, which is the
// right thing for resolution and the wrong thing for a diagnostic.
func doctorTemplatesDir(cmd *cobra.Command) (string, string, error) {
	if f := cmd.Flags().Lookup("templates-dir"); f != nil {
		if v := f.Value.String(); v != "" {
			dir, err := expandHome(v)
			return dir, "--templates-dir", err
		}
	}
	envKey := config.EnvPrefix + "_TEMPLATES_DIR"
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		dir, err := expandHome(v)
		return dir, envKey, err
	}
	// Anything left comes from the config file layer (or a test override,
	// which is the same layer as far as a user is concerned).
	if v := strings.TrimSpace(config.Global().GetString("templates_dir")); v != "" {
		dir, err := expandHome(v)
		return dir, "the config file", err
	}
	return "", "", nil
}

// checkConfigFile reports which config file will be read and whether it is
// there. Its absence is the documented default on a fresh install, not a
// defect, so a missing file is ok.
func checkConfigFile(env *doctorEnv) checkResult {
	if env.configPathErr != nil {
		return checkResult{
			status:  statusError,
			summary: "config path could not be resolved: " + env.configPathErr.Error(),
			remedy:  "pass an explicit path with --config FILE",
		}
	}
	if env.configPath == "" {
		return checkResult{
			status:  statusWarn,
			summary: "no config path could be determined (no home directory?)",
			remedy:  "pass an explicit path with --config FILE",
		}
	}
	location := doctorConfigLocation(env)
	if fileExists(env.configPath) {
		return checkResult{
			status:  statusOK,
			summary: fmt.Sprintf("reading %s (%s)", env.configPath, location),
		}
	}
	return checkResult{
		status: statusOK,
		summary: fmt.Sprintf("no config file at %s (%s); flags, environment and defaults supply everything",
			env.configPath, location),
	}
}

func doctorConfigLocation(env *doctorEnv) string {
	switch {
	case env.configExplicit:
		return "explicit --config"
	case env.configPath == legacyConfigPath():
		return "legacy location"
	case env.configPath == xdgConfigPath():
		return "XDG location"
	default:
		return "resolved location"
	}
}

// checkConfigParse surfaces the load error that startup deliberately
// discards. Config is optional, so a malformed file never fails a command
// that needs nothing from it — which is exactly why nothing else reports it.
func checkConfigParse(env *doctorEnv) checkResult {
	if env.configPathErr != nil || env.configPath == "" || !fileExists(env.configPath) {
		return checkResult{status: statusOK, summary: "no config file to parse"}
	}
	// A throwaway instance, so inspecting the file cannot disturb the
	// process-wide config the other checks read.
	if err := config.New().Load(env.configPath); err != nil {
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("%s does not parse: %v (its values are ignored)", env.configPath, err),
			remedy:  "fix the YAML in " + env.configPath + ", or remove the file to fall back to flags and environment",
		}
	}
	return checkResult{status: statusOK, summary: env.configPath + " parses"}
}

// checkConfigShadowed catches the failure the XDG-with-legacy-precedence
// rule exists to avoid falling into silently: two config files present, one
// of them never read, and no other command's output shows it.
func checkConfigShadowed(env *doctorEnv) checkResult {
	legacy, xdg := legacyConfigPath(), xdgConfigPath()
	switch {
	case legacy == "" || xdg == "" || !fileExists(legacy) || !fileExists(xdg):
		return checkResult{status: statusOK, summary: "only one config file location is in use"}
	case env.configExplicit:
		return checkResult{
			status:  statusOK,
			summary: fmt.Sprintf("%s and %s both exist, but --config %s overrides both", legacy, xdg, env.configPath),
		}
	default:
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("%s (legacy) is in effect; %s (XDG) also exists and is never read", legacy, xdg),
			remedy:  fmt.Sprintf("move your settings into one file and delete the other, or pass --config %s to choose explicitly", xdg),
		}
	}
}

// checkTemplatesShadowed is the templates-directory half of the same trap.
// When nothing is configured, generators use the embedded templates and the
// `srekit templates` subcommands operate on the legacy directory, so the XDG
// one sits there doing nothing.
func checkTemplatesShadowed(env *doctorEnv) checkResult {
	legacy, xdg := legacyTemplatesDir(), xdgTemplatesDir()
	switch {
	case legacy == "" || xdg == "" || !dirExists(legacy) || !dirExists(xdg):
		return checkResult{status: statusOK, summary: "only one templates directory location is in use"}
	case env.templatesSource != "":
		return checkResult{
			status:  statusOK,
			summary: fmt.Sprintf("%s and %s both exist, but %s overrides both", legacy, xdg, env.templatesSource),
		}
	default:
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("'srekit templates' operates on %s (legacy); %s (XDG) also exists and is ignored", legacy, xdg),
			remedy:  "merge the two directories and delete the other, or set templates_dir in the config file to choose explicitly",
		}
	}
}

// checkConfigWritable answers "can `srekit config init` write here" without
// writing anything: it walks up to the nearest directory that exists and
// judges it from its mode bits and ownership.
func checkConfigWritable(env *doctorEnv) checkResult {
	if env.configPathErr != nil || env.configPath == "" {
		return checkResult{status: statusOK, summary: "no config path to check"}
	}
	dir, err := nearestExistingDir(filepath.Dir(env.configPath))
	if err != nil {
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("could not inspect the parent directories of %s: %v", env.configPath, err),
			remedy:  "check the permissions on the path leading to " + env.configPath,
		}
	}
	info, err := os.Stat(dir)
	if err != nil {
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("could not stat %s: %v", dir, err),
			remedy:  "check the permissions on " + dir,
		}
	}
	writable, known := dirWritable(info)
	switch {
	case !known:
		return checkResult{status: statusOK, summary: "writability of " + dir + " cannot be determined on this platform"}
	case writable:
		return checkResult{status: statusOK, summary: dir + " is writable"}
	default:
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("%s is not writable; 'srekit config init' cannot create %s", dir, env.configPath),
			remedy:  fmt.Sprintf("fix the permissions on %s, or pass --config FILE pointing somewhere writable", dir),
		}
	}
}

// nearestExistingDir walks up from dir until it finds a path that exists,
// which is where a write would actually be attempted.
func nearestExistingDir(dir string) (string, error) {
	for {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir, nil
		}
		dir = parent
	}
}

// checkConfigEnv lists the SREKIT_ variables currently overriding config
// values, by name — the values themselves show up in the checks that
// consume them.
func checkConfigEnv(_ *doctorEnv) checkResult {
	prefix := config.EnvPrefix + "_"
	var names []string
	for _, kv := range os.Environ() {
		name, _, ok := strings.Cut(kv, "=")
		if ok && strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return checkResult{status: statusOK, summary: "no " + prefix + " environment overrides are set"}
	}
	return checkResult{
		status:  statusOK,
		summary: fmt.Sprintf("%d environment override(s) in effect: %s", len(names), strings.Join(names, ", ")),
	}
}

// checkTemplatesDir reports where the templates directory resolved, what
// supplied it, and whether it is usable. A configured directory that is
// missing is a warning rather than an error: generation still works from the
// embedded set, but the user's overrides are silently not applied.
func checkTemplatesDir(env *doctorEnv) checkResult {
	if env.templatesErr != nil {
		return checkResult{
			status:  statusError,
			summary: "templates directory could not be resolved: " + env.templatesErr.Error(),
			remedy:  "pass an absolute path with --templates-dir DIR",
		}
	}
	if env.templatesDir == "" {
		return checkResult{
			status: statusOK,
			summary: fmt.Sprintf("no templates directory configured (built-in default); generation uses the embedded templates, "+
				"and 'srekit templates init' would scaffold %s", resolveDefaultTemplatesDir()),
		}
	}
	info, err := os.Stat(env.templatesDir)
	switch {
	case err != nil:
		return checkResult{
			status: statusWarn,
			summary: fmt.Sprintf("%s (from %s) cannot be read: %v; generation is falling back to the embedded templates",
				env.templatesDir, env.templatesSource, err),
			remedy: fmt.Sprintf("run 'srekit templates init %s', or point %s somewhere that exists",
				env.templatesDir, env.templatesSource),
		}
	case !info.IsDir():
		return checkResult{
			status: statusWarn,
			summary: fmt.Sprintf("%s (from %s) is not a directory; generation is falling back to the embedded templates",
				env.templatesDir, env.templatesSource),
			remedy: "point " + env.templatesSource + " at a directory",
		}
	default:
		return checkResult{
			status:  statusOK,
			summary: fmt.Sprintf("%s (from %s) is in use", env.templatesDir, env.templatesSource),
		}
	}
}

// checkIdentity reports whether an author can be resolved at all, using the
// same resolver every generator calls, and names the source each value came
// from. The values are printed as they are: they are already stamped into
// every artifact srekit generates, and redacting them would make the check
// useless for diagnosing a wrong author.
func checkIdentity(_ *doctorEnv) checkResult {
	author, _ := meta.Resolve(config.Global(), "", "")
	switch {
	case author.Name != "" && author.Email != "":
		return checkResult{
			status: statusOK,
			summary: fmt.Sprintf("%s <%s> (name from %s, email from %s)",
				author.Name, author.Email, authorNameSource(), authorEmailSource()),
		}
	case author.Name != "":
		return checkResult{
			status:  statusError,
			summary: fmt.Sprintf("author name resolves (%s from %s) but no email does; every generator that stamps an author will fail", author.Name, authorNameSource()),
			remedy:  identityRemedy,
		}
	case author.Email != "":
		return checkResult{
			status:  statusError,
			summary: fmt.Sprintf("author email resolves (%s from %s) but no name does; every generator that stamps an author will fail", author.Email, authorEmailSource()),
			remedy:  identityRemedy,
		}
	default:
		return checkResult{
			status:  statusError,
			summary: "no author name or email can be resolved; every generator that stamps an author will fail",
			remedy:  identityRemedy,
		}
	}
}

const identityRemedy = "run 'srekit config init', pass --author/--email, or set git config user.name and user.email"

// authorNameSource and authorEmailSource mirror meta.Resolve's precedence to
// name where a resolved value came from. The config package folds its
// environment and file layers into one lookup, so the environment is checked
// separately here and anything left is reported as the config file.
func authorNameSource() string {
	switch {
	case envOverride("AUTHOR") != "":
		return config.EnvPrefix + "_AUTHOR"
	case envOverride("FULL_NAME") != "":
		return config.EnvPrefix + "_FULL_NAME"
	case strings.TrimSpace(config.Global().GetString("author")) != "":
		return "the config file (author)"
	case strings.TrimSpace(config.Global().GetString("full_name")) != "":
		return "the config file (full_name)"
	default:
		return "git config user.name"
	}
}

func authorEmailSource() string {
	switch {
	case envOverride("EMAIL") != "":
		return config.EnvPrefix + "_EMAIL"
	case strings.TrimSpace(config.Global().GetString("email")) != "":
		return "the config file (email)"
	default:
		return "git config user.email"
	}
}

func envOverride(key string) string {
	return strings.TrimSpace(os.Getenv(config.EnvPrefix + "_" + key))
}

// checkTemplatesParse runs the same per-file dispatch `templates validate`
// uses over the user's directory. A parse failure is an error: the generator
// backed by that artifact cannot render at all.
func checkTemplatesParse(env *doctorEnv) checkResult {
	dir, ok := env.usableTemplatesDir()
	if !ok {
		return checkResult{status: statusOK, summary: "no templates directory in use; only the embedded templates are in play"}
	}
	names, err := artifactFileNames(dir)
	if err != nil {
		return checkResult{
			status:  statusError,
			summary: fmt.Sprintf("%s cannot be read: %v", dir, err),
			remedy:  "fix the permissions on " + dir + ", or point --templates-dir elsewhere",
		}
	}
	if len(names) == 0 {
		return checkResult{status: statusOK, summary: "no template artifacts in " + dir}
	}
	var failures []string
	for _, name := range names {
		body, readErr := os.ReadFile(filepath.Join(dir, name))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: read: %v", name, readErr))
			continue
		}
		if _, parseErr := validateArtifactBody(name, body); parseErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", name, parseErr))
		}
	}
	if len(failures) > 0 {
		return checkResult{
			status: statusError,
			summary: fmt.Sprintf("%d of %d artifact(s) in %s fail to parse: %s",
				len(failures), len(names), dir, strings.Join(failures, "; ")),
			remedy: "fix the reported files; 'srekit templates validate' reports them one per line",
		}
	}
	return checkResult{status: statusOK, summary: fmt.Sprintf("all %d artifact(s) in %s parse", len(names), dir)}
}

// checkTemplatesLegacy reports pre-v1.0 template files. They still load
// through 1.x, but they are dead weight next to the v1 artifact that
// supersedes them.
func checkTemplatesLegacy(env *doctorEnv) checkResult {
	dir, ok := env.usableTemplatesDir()
	if !ok {
		return checkResult{status: statusOK, summary: "no templates directory in use; only the embedded templates are in play"}
	}
	names, err := artifactFileNames(dir)
	if err != nil {
		return checkResult{
			status:  statusError,
			summary: fmt.Sprintf("%s cannot be read: %v", dir, err),
			remedy:  "fix the permissions on " + dir + ", or point --templates-dir elsewhere",
		}
	}
	var legacy []string
	for _, name := range names {
		if isLegacyArtifactName(name) {
			legacy = append(legacy, name)
		}
	}
	if len(legacy) == 0 {
		return checkResult{status: statusOK, summary: "no pre-v1.0 template files in " + dir}
	}
	return checkResult{
		status: statusWarn,
		summary: fmt.Sprintf("%d pre-v1.0 template file(s) in %s: %s",
			len(legacy), dir, strings.Join(legacy, ", ")),
		remedy: "run 'srekit templates migrate' to convert them to the v1 artifact format",
	}
}

// checkTemplatesDrift reuses `templates list`'s classification so this count
// and that command's output can never disagree.
func checkTemplatesDrift(env *doctorEnv) checkResult {
	dir, ok := env.usableTemplatesDir()
	if !ok {
		return checkResult{status: statusOK, summary: "no templates directory in use; only the embedded templates are in play"}
	}
	entries, err := classifyTemplates(dir)
	if err != nil {
		return checkResult{
			status:  statusError,
			summary: fmt.Sprintf("%s could not be compared against the embedded templates: %v", dir, err),
			remedy:  "fix the permissions on " + dir + ", or point --templates-dir elsewhere",
		}
	}
	var customized, embeddedOnly int
	for _, e := range entries {
		switch e.Status {
		case "customized":
			customized++
		case "embedded-only":
			embeddedOnly++
		}
	}
	if customized == 0 && embeddedOnly == 0 {
		return checkResult{status: statusOK, summary: dir + " matches the embedded templates"}
	}
	return checkResult{
		status: statusWarn,
		summary: fmt.Sprintf("%d artifact(s) in %s differ from the embedded version, %d embedded artifact(s) are absent from it",
			customized, dir, embeddedOnly),
		remedy: "run 'srekit templates diff' to see what changed, then 'srekit templates upgrade' to sync",
	}
}

// checkGit is the only external-program check, and the only subprocess
// doctor starts. An absent git is a warning: author metadata and the
// changelog repository slug fall back to flags and config, so most
// generation still works.
func checkGit(env *doctorEnv) checkResult {
	path, err := exec.LookPath("git")
	if err != nil {
		return checkResult{
			status: statusWarn,
			summary: "git is not on PATH; author name and email fall back to --author/--email and the config file, " +
				"and 'srekit changelog' cannot detect the repository slug",
			remedy: "install git, or pass --author/--email and --repo OWNER/REPO explicitly",
		}
	}
	// Run through the command context so Ctrl-C tears the subprocess down.
	// path comes from exec.LookPath("git"), never from user input.
	out, err := exec.CommandContext(env.cmd.Context(), path, "--version").Output()
	if err != nil {
		return checkResult{
			status:  statusWarn,
			summary: fmt.Sprintf("%s could not be run: %v", path, err),
			remedy:  "reinstall git, or fix the permissions on " + path,
		}
	}
	return checkResult{
		status:  statusOK,
		summary: fmt.Sprintf("%s (%s)", path, strings.TrimSpace(string(out))),
	}
}
