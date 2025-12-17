package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/haltman-io/reverse-whois/internal/util"
)

type multiString struct {
	values *[]string
}

func (m *multiString) String() string {
	if m == nil || m.values == nil {
		return ""
	}
	return strings.Join(*m.values, ",")
}

func (m *multiString) Set(s string) error {
	if m == nil || m.values == nil {
		return fmt.Errorf("internal flag error: nil receiver")
	}
	parts := util.SplitCSV(s)
	for _, p := range parts {
		if p == "" {
			continue
		}
		*m.values = append(*m.values, p)
	}
	return nil
}

func ParseFlags(args []string) (Config, bool, error) {
	var cfg Config
	var help bool

	fs := flag.NewFlagSet("reverse-whois", flag.ContinueOnError)
	fs.SetOutput(util.NewDiscardWriter())

	// Help
	fs.BoolVar(&help, "help", false, "Show help.")
	fs.BoolVar(&help, "h", false, "Alias for --help.")

	// Targets
	targetMulti := &multiString{values: &cfg.Targets}
	listMulti := &multiString{values: &cfg.TargetLists}
	exclMulti := &multiString{values: &cfg.ExcludeTerms}

	fs.Var(targetMulti, "target", "Define the search term target (repeatable; supports comma-separated values).")
	fs.Var(targetMulti, "t", "Alias for --target.")

	fs.Var(listMulti, "target-list", "Define a file containing targets (repeatable; supports comma-separated paths).")
	fs.Var(listMulti, "tL", "Alias for --target-list.")

	// API key
	fs.StringVar(&cfg.APIKey, "api-key", "", "API key (highest priority over .reverse-whois.yaml).")

	// Historic / preview
	fs.BoolVar(&cfg.Historic, "history", false, "Search in historical WHOIS records (sets searchType=historic).")
	fs.BoolVar(&cfg.Historic, "historic", false, "Alias for --history.")

	fs.BoolVar(&cfg.Preview, "preview", false, "Preview mode: return only domainsCount (sets mode=preview).")
	fs.BoolVar(&cfg.Preview, "check", false, "Alias for --preview.")

	// Exclude terms (max 4 enforced later too)
	fs.Var(exclMulti, "exclude", "Exclude term(s), repeatable and comma-separated (max 4 total).")
	fs.Var(exclMulti, "e", "Alias for --exclude.")

	// Output & logging
	fs.BoolVar(&cfg.Silent, "silent", false, "Results only (no logs).")
	fs.BoolVar(&cfg.Silent, "s", false, "Alias for --silent.")

	fs.BoolVar(&cfg.Quiet, "quiet", false, "Results only (no logs).")
	fs.BoolVar(&cfg.Quiet, "q", false, "Alias for --quiet.")

	fs.BoolVar(&cfg.Verbose, "verbose", false, "Enable debug logs (stderr).")
	fs.BoolVar(&cfg.Verbose, "v", false, "Alias for --verbose.")

	fs.BoolVar(&cfg.Debug, "debug", false, "Enable debug logs (stderr).")

	fs.BoolVar(&cfg.NoColor, "no-color", false, "Disable ANSI colors.")
	fs.BoolVar(&cfg.NoColor, "nc", false, "Alias for --no-color.")

	// Output to file
	fs.StringVar(&cfg.OutputPath, "output", "", "Write results to a file (overwrite by default).")
	fs.StringVar(&cfg.OutputPath, "o", "", "Alias for --output.")
	fs.StringVar(&cfg.OutputPath, "out", "", "Alias for --output.")

	// Proxy / TLS
	fs.StringVar(&cfg.ProxyURL, "proxy", "", "Proxy URL (http://, https://, socks5://).")
	fs.BoolVar(&cfg.NoProxy, "no-proxy", false, "Disable proxy usage even if environment variables exist.")
	fs.BoolVar(&cfg.Insecure, "insecure", false, "Skip TLS certificate verification (unsafe).")
	fs.BoolVar(&cfg.Insecure, "k", false, "Alias for --insecure.")

	// Threads / rate limit
	fs.IntVar(&cfg.Threads, "threads", 1, "Number of worker threads (default: 1).")
	fs.IntVar(&cfg.RateLimit, "rate-limit", 0, "Global max requests per second (RPS). Max allowed: 30.")
	fs.IntVar(&cfg.RateLimit, "rl", 0, "Alias for --rate-limit.")

	if err := fs.Parse(args); err != nil {
		return Config{}, false, fmt.Errorf("flag parsing error: %w\n\nRun with -h for help.", err)
	}

	if help {
		return cfg, true, nil
	}

	if cfg.Threads < 1 {
		return Config{}, false, fmt.Errorf("--threads must be >= 1")
	}
	if cfg.RateLimit < 0 {
		return Config{}, false, fmt.Errorf("--rate-limit must be >= 1")
	}
	if cfg.RateLimit > 30 {
		return Config{}, false, fmt.Errorf("--rate-limit exceeds API maximum (30 rps)")
	}
	if len(cfg.ExcludeTerms) > 4 {
		return Config{}, false, fmt.Errorf("exclude terms limit exceeded: maximum 4 items allowed")
	}

	return cfg, false, nil
}
