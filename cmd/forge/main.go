package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pablogore/devforge/internal/adapters/logger"
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/detection"
	"github.com/pablogore/devforge/internal/doctor"
	initcmd "github.com/pablogore/devforge/internal/init"
	"github.com/pablogore/devforge/internal/profiles"
	"github.com/pablogore/devforge/internal/runtime"
	"github.com/pablogore/devforge/internal/tools"
)

const (
	ExitCodeSuccess = 0
	ExitCodeFailure = 1
	ExitCodeUsage   = 2
)

func main() {
	env := runtime.DetectEnvironment()
	if os.Getenv("DEVFORGE_DEBUG") != "" && env.IsGitHub {
		fmt.Fprintln(os.Stderr, "Running inside GitHub Actions")
	}

	if len(os.Args) < 2 {
		rootUsage()
		os.Exit(ExitCodeUsage)
	}
	if handleHelp() {
		return
	}
	os.Exit(dispatchCommand())
}

// handleHelp returns true if help was shown (caller should exit). Exits internally on help.
func handleHelp() bool {
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		rootUsage()
		os.Exit(ExitCodeUsage)
	}
	if len(os.Args) < 3 {
		return false
	}
	if os.Args[2] != "-h" && os.Args[2] != "--help" {
		return false
	}
	showCommandHelp(os.Args[1])
	os.Exit(ExitCodeUsage)
	return true
}

func showCommandHelp(cmd string) {
	switch cmd {
	case "pr":
		prUsage()
	case "release":
		releaseUsage()
	case "doctor":
		doctorUsage()
	case "init":
		initUsage()
	case "run":
		runUsage()
	default:
		//nolint:gosec // CLI stderr output; not user-facing HTML
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", cmd)
		rootUsage()
	}
}

// dispatchCommand runs the subcommand and returns the exit code.
func dispatchCommand() int {
	switch os.Args[1] {
	case "pr":
		runPR()
		return ExitCodeSuccess
	case "release":
		runRelease()
		return ExitCodeSuccess
	case "doctor":
		runDoctor()
		return ExitCodeSuccess
	case "run":
		runRun()
		return ExitCodeSuccess
	case "init":
		runInit()
		return ExitCodeSuccess
	default:
		//nolint:gosec // CLI stderr output; not user-facing HTML
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", os.Args[1])
		rootUsage()
		return ExitCodeUsage
	}
}

func rootUsage() {
	fmt.Printf(`DevForge — Local-first CI automation engine

Usage:
  forge <command> [flags]

Commands:
  pr         Run PR validation
  release    Run release flow
  doctor     Validate local environment
  init       Initialize DevForge configuration
  run        Execute specific steps locally

Examples:
  forge pr --profile go-lib
  forge release --profile go-lib
  forge doctor --profile go-lib
  forge init
  forge run gofmt staticcheck

Use "forge <command> --help" for more information about a command.
`)
}

func prUsage() {
	fmt.Printf(`Usage:
  forge pr [flags]

Description:
  Run PR validation checks including linting, testing, and conventional commit validation.
  Profile and mode are resolved in order: CLI flags override .devforge.yml, then auto-detection (profile) or default (mode: full).

Flags:
  --profile        Profile to use (optional; auto-detected if not set): %s
  --mode           (optional) quick | full | deep (default: full)
  --workdir        Working directory (default: ".")
  --base-ref       Base ref for PR (default: "origin/main")
  --title          PR title override (optional, uses latest commit if not set)
  -h, --help       Show help

Examples:
  forge pr
  forge pr --profile go-service
  forge pr --mode deep
  forge pr --profile go-lib --title "fix: bug fix"
`, strings.Join(profiles.List(), ", "))
}

func releaseUsage() {
	fmt.Printf(`Usage:
  forge release [flags]

Description:
  Run release flow including version derivation, tagging, and artifact publishing.
  If --profile is not provided, DevForge will attempt to detect the repository type.

Flags:
  --profile        Profile to use (optional; auto-detected if not set): %s
  --workdir        Working directory (default: ".")
  -h, --help       Show help

Examples:
  forge release --profile go-lib
`, strings.Join(profiles.List(), ", "))
}

func doctorUsage() {
	fmt.Printf(`Usage:
  forge doctor [flags]

Description:
  Validate local environment prerequisites for release including git, goreleaser,
  branch status, working tree, and tag accessibility.
  If --profile is not provided, DevForge will attempt to detect the repository type.

Flags:
  --profile           Profile to use (optional; auto-detected if not set): %s
  --workdir           Working directory (default: ".")
  --generate-policies Generate policy packs under .devforge/policies/ from analysis
  -h, --help          Show help

Examples:
  forge doctor --profile go-lib
  forge doctor --generate-policies
`, strings.Join(profiles.List(), ", "))
}

func initUsage() {
	fmt.Printf(`Usage:
  forge init [flags]

Description:
  Initialize DevForge configuration for this repository.
  Creates .devforge/, .devforge/policies/, and config files.
  If config files exist, prompts to overwrite unless -f/--force is set.

Flags:
  -f, --force  Overwrite existing .devforge.yml and .golangci.yml without prompting
  --workdir    Working directory (default: ".")

Examples:
  forge init
  forge init -f
  forge init --workdir /path/to/repo
`)
}

func runUsage() {
	fmt.Printf(`Usage:
  forge run [flags] <step> [<step> ...]

Description:
  Execute specific steps locally. Steps run in the given order.
  Use --workdir to set the working directory (default: ".").

Flags:
  --workdir    Working directory (default: ".")
  -h, --help   Show help

Available steps: %s

Examples:
  forge run gofmt
  forge run golangci-lint
  forge run govulncheck
  forge run --workdir /path/to/repo golangci-lint govulncheck
`, strings.Join(application.ListSteps(), ", "))
}

func runInit() {
	workdir := getFlagValueWithDefault("--workdir", ".")
	force := getFlagBool("-f") || getFlagBool("--force")

	existing := initcmd.ExistingConfigFiles(workdir)
	if len(existing) > 0 && !force {
		fmt.Fprintln(os.Stderr, "Warning: the following config files already exist and would be overwritten:")
		for _, p := range existing {
			fmt.Fprintln(os.Stderr, "  - "+p)
		}
		fmt.Fprint(os.Stderr, "Continue and overwrite? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Aborted (no input).")
			os.Exit(ExitCodeSuccess)
		}
		answer := strings.TrimSpace(strings.ToLower(line))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "Aborted.")
			os.Exit(ExitCodeSuccess)
		}
		force = true
	}

	result, err := initcmd.InitRepository(workdir, force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	fmt.Println("Initialized DevForge for this repository.")
	if len(result.Created) > 0 {
		fmt.Println()
		fmt.Println("Created:")
		for _, path := range result.Created {
			fmt.Println("  " + path)
		}
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  git add .devforge .devforge.yml")
	fmt.Println("  git commit")
	fmt.Println()
	fmt.Println("DevForge will now enforce policies in PR pipelines.")
	if result.Suggestion != "" {
		fmt.Println()
		fmt.Println(result.Suggestion)
	}
}

func runRun() {
	if err := tools.EnsureTools(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	workdir := "."
	var stepNames []string
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--workdir" && i+1 < len(os.Args) {
			workdir = os.Args[i+1]
			i++ // skip value; loop post will advance again
			continue
		}
		if os.Args[i] == "-h" || os.Args[i] == "--help" {
			runUsage()
			os.Exit(ExitCodeUsage)
		}
		stepNames = append(stepNames, os.Args[i])
	}
	if len(stepNames) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one step is required")
		runUsage()
		os.Exit(ExitCodeUsage)
	}
	err := profiles.RunSteps(workdir, stepNames)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitCodeFailure)
	}
}

func runPR() {
	if err := tools.EnsureTools(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	workdir := getFlagValueWithDefault("--workdir", ".")
	if workdir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			os.Exit(ExitCodeFailure)
		}
		workdir = cwd
	}
	cfg, err := config.LoadConfig(workdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	if os.Getenv("DEVFORGE_DEBUG") != "" {
		hasPolicy := cfg != nil && cfg.Policies != nil && cfg.Policies.Coverage != nil
		fmt.Fprintf(os.Stderr, "[devforge] config from %s/.devforge.yml policies_coverage=%v\n", workdir, hasPolicy)
		if hasPolicy {
			fmt.Fprintf(os.Stderr, "[devforge] coverage policy threshold=%d packages=%v\n", cfg.Policies.Coverage.Threshold, cfg.Policies.Coverage.Packages)
		}
	}
	// Priority: 1 CLI flags, 2 .devforge.yml, 3 auto detection
	profileName := getFlagValue("--profile")
	if profileName == "" && cfg.Profile != "" {
		profileName = cfg.Profile
	}
	if profileName == "" {
		profileName = detection.DetectProfile(workdir)
		log := logger.New("info", "text")
		log.Info("profile auto-detected", "profile", profileName)
	}
	modeStr := getFlagValue("--mode")
	if modeStr == "" {
		modeStr = cfg.Mode
	}
	if modeStr == "" {
		modeStr = "full"
	}
	baseRef := getFlagValueWithDefault("--base-ref", "origin/main")
	title := getFlagValue("--title")

	mode, err := application.ParseMode(modeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid mode: %s\n", modeStr)
		os.Exit(ExitCodeFailure)
	}

	p, ok := profiles.Get(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown profile '%s'\n", profileName)
		prUsage()
		os.Exit(ExitCodeUsage)
	}
	err = p.RunPRWithMode(workdir, baseRef, title, mode, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitCodeFailure)
	}
}

func runRelease() {
	if err := tools.EnsureTools(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	workdir := getFlagValueWithDefault("--workdir", ".")
	cfg, err := config.LoadConfig(workdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	// Priority: 1 CLI flags, 2 .devforge.yml, 3 auto detection
	profileName := getFlagValue("--profile")
	if profileName == "" && cfg.Profile != "" {
		profileName = cfg.Profile
	}
	if profileName == "" {
		profileName = detection.DetectProfile(workdir)
		log := logger.New("info", "text")
		log.Info("profile auto-detected", "profile", profileName)
	}

	p, ok := profiles.Get(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown profile '%s'\n", profileName)
		releaseUsage()
		os.Exit(ExitCodeUsage)
	}
	version, err := p.RunRelease(workdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitCodeFailure)
	}
	fmt.Printf("Release completed: %s\n", version)
}

func runDoctor() {
	if err := tools.EnsureTools(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	workdir := getFlagValueWithDefault("--workdir", ".")
	cfg, err := config.LoadConfig(workdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(ExitCodeFailure)
	}
	// Priority: 1 CLI flags, 2 .devforge.yml, 3 auto detection
	profileName := getFlagValue("--profile")
	if profileName == "" && cfg.Profile != "" {
		profileName = cfg.Profile
	}
	if profileName == "" {
		profileName = detection.DetectProfile(workdir)
		log := logger.New("info", "text")
		log.Info("profile auto-detected", "profile", profileName)
	}

	p, ok := profiles.Get(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown profile '%s'\n", profileName)
		doctorUsage()
		os.Exit(ExitCodeUsage)
	}
	result, err := p.RunDoctor(workdir)

	if hasFlag("--generate-policies") {
		generated, genErr := doctor.GeneratePolicies(workdir)
		if genErr != nil {
			fmt.Fprintf(os.Stderr, "Error generating policies: %v\n", genErr)
			os.Exit(ExitCodeFailure)
		}
		if len(generated) == 0 {
			fmt.Println("No policies needed.")
		} else {
			fmt.Println("Generated policy packs:")
			for _, path := range generated {
				fmt.Println("  " + path)
			}
			fmt.Println("\nThese policies will now be enforced by DevForge PR pipelines.")
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(ExitCodeFailure)
		}
		return
	}

	if result != nil {
		if err != nil {
			fmt.Println("Doctor check FAILED:")
		} else {
			fmt.Println("Doctor check results:")
		}
		for _, check := range result.Checks {
			status := "PASS"
			if !check.Passed {
				status = "FAIL"
			}
			fmt.Printf("  [%s] %s: %s\n", status, check.Name, check.Message)
		}
	}
	// Policy suggestions (never modify repo; recommendations only)
	suggestions := doctor.GenerateSuggestions(workdir)
	fmt.Println()
	if len(suggestions) == 0 {
		fmt.Println("No policy suggestions detected.")
	} else {
		fmt.Println("Suggested policies:")
		for _, s := range suggestions {
			fmt.Printf("\n%s\n", s.File)
			for _, r := range s.Rules {
				fmt.Printf("  %s\n", r)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitCodeFailure)
	}
}

func hasFlag(name string) bool {
	for _, arg := range os.Args {
		if arg == name {
			return true
		}
	}
	return false
}

func getFlagValue(name string) string {
	for i, arg := range os.Args {
		if arg == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

func getFlagValueWithDefault(name, def string) string {
	val := getFlagValue(name)
	if val == "" {
		return def
	}
	return val
}

func getFlagBool(name string) bool {
	return hasFlag(name)
}
