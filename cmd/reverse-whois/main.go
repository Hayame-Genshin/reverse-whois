package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/haltman-io/reverse-whois/internal/api"
	"github.com/haltman-io/reverse-whois/internal/cli"
	"github.com/haltman-io/reverse-whois/internal/output"
	"github.com/haltman-io/reverse-whois/internal/ratelimit"
	"github.com/haltman-io/reverse-whois/internal/targets"
	"github.com/haltman-io/reverse-whois/internal/util"
)

const (
	ToolName    = "reverse-whois"
	ToolVersion = "v1.0.1-stable"
)

type job struct {
	SearchTerm string
}

type result struct {
	Lines       []string
	OutputItems []string // clean items intended for --output file
	Err         error
}

func main() {
	os.Exit(run())
}

func run() int {
	cfg, wantHelp, err := cli.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if wantHelp {
		cli.PrintUsage(os.Stdout)
		return 0
	}

	logger := util.NewLogger(os.Stderr, cfg.Verbose || cfg.Debug, cfg.Silent || cfg.Quiet)

	// Config file near executable: create if missing.
	configPath, err := util.ResolveConfigPathNearExecutable(".reverse-whois.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to resolve config path:", err.Error())
		return 2
	}

	keys, err := util.LoadOrInitAPIKeysYAML(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read config:", err.Error())
		return 2
	}

	var keyProvider util.APIKeyProvider
	if cfg.APIKey != "" {
		keyProvider = util.NewStaticKeyProvider(cfg.APIKey)
	} else {
		if len(keys) == 0 {
			fmt.Fprintln(os.Stderr, "No API key configured. Provide --api-key or add at least one key to .reverse-whois.yaml (api_keys).")
			return 2
		}
		keyProvider = util.NewRotatingKeyProvider(keys)
	}

	// Collect targets from stdin + flags + list files (deduped).
	collectedTargets, err := targets.CollectTargets(targets.CollectOptions{
		TargetValues: cfg.Targets,
		ListFiles:    cfg.TargetLists,
		ReadStdin:    true,
		Logger:       logger,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if len(collectedTargets) == 0 {
		fmt.Fprintln(os.Stderr, "No targets provided. Use --target/-t, --target-list/-tL, or stdin.")
		return 2
	}

	// Validate exclude terms max 4.
	if len(cfg.ExcludeTerms) > 4 {
		fmt.Fprintln(os.Stderr, "Exclude terms limit exceeded: maximum 4 items allowed.")
		return 2
	}

	// searchType default current; historic if flag.
	searchType := "current"
	if cfg.Historic {
		searchType = "historic"
	}

	// mode default purchase; preview if flag.
	mode := "purchase"
	if cfg.Preview {
		mode = "preview"
	}

	// Context fixed max: 30 rps; user cannot exceed it.
	effectiveRPS := 30
	if cfg.RateLimit > 0 {
		if cfg.RateLimit > 30 {
			fmt.Fprintln(os.Stderr, "--rate-limit exceeds API maximum (30 rps).")
			return 2
		}
		if cfg.RateLimit < effectiveRPS {
			effectiveRPS = cfg.RateLimit
		}
	}
	if effectiveRPS < 1 {
		fmt.Fprintln(os.Stderr, "--rate-limit must be >= 1.")
		return 2
	}

	if cfg.Threads < 1 {
		fmt.Fprintln(os.Stderr, "--threads must be >= 1.")
		return 2
	}

	limiter := ratelimit.NewLimiter(effectiveRPS)

	// Output writers: stdout (formatted) + optional file (clean results).
	stdoutWriter, fileWriter, closer, err := util.BuildOutputWriters(os.Stdout, cfg.OutputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	if closer != nil {
		defer func() { _ = closer.Close() }()
	}

	colorEnabled := !cfg.NoColor
	stdoutPrinter := output.NewPrinter(stdoutWriter, output.NewColorizer(colorEnabled))

	if !(cfg.Silent || cfg.Quiet) {
		output.PrintBanner(stdoutPrinter, ToolName, ToolVersion)
	}


	httpClient, err := api.NewClient(api.ClientOptions{
		Endpoint:  "https://reverse-whois.whoisxmlapi.com/api/v2",
		ProxyURL:  cfg.ProxyURL,
		NoProxy:   cfg.NoProxy,
		Insecure:  cfg.Insecure,
		Timeout:   30 * time.Second,
		UserAgent: fmt.Sprintf("%s/%s", ToolName, ToolVersion),
		Logger:    logger,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}

	ctx := context.Background()

	jobs := make(chan job)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		workerID := i + 1
		go func() {
			defer wg.Done()
			for j := range jobs {
				lines, items, e := handleOne(ctx, httpClient, limiter, keyProvider, searchType, mode, cfg.ExcludeTerms, j.SearchTerm, logger, workerID, stdoutPrinter)
				results <- result{Lines: lines, OutputItems: items, Err: e}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		defer close(jobs)
		for _, t := range collectedTargets {
			jobs <- job{SearchTerm: t}
		}
	}()

	exitCode := 0
	fileSeen := make(map[string]struct{})

	for res := range results {
		if res.Err != nil {
			exitCode = 1
			continue
		}

		// stdout: keep formatted lines
		for _, line := range res.Lines {
			stdoutPrinter.Println(line)
		}

		// file: clean results only (dedup)
		if fileWriter != nil {
			for _, item := range res.OutputItems {
				if item == "" {
					continue
				}
				if _, ok := fileSeen[item]; ok {
					continue
				}
				fileSeen[item] = struct{}{}
				_, _ = fmt.Fprintln(fileWriter, item)
			}
		}
	}

	return exitCode
}

func handleOne(
	ctx context.Context,
	client *api.Client,
	limiter *ratelimit.Limiter,
	keyProvider util.APIKeyProvider,
	searchType string,
	mode string,
	exclude []string,
	term string,
	logger *util.Logger,
	workerID int,
	stdoutPrinter *output.Printer,
) ([]string, []string, error) {
	// Global rate limit across all threads.
	if err := limiter.Take(ctx); err != nil {
		line := output.FormatErrorLine(stdoutPrinter.Colors(), term, mode, searchType, "rate_limit", err.Error())
		stdoutPrinter.Println(line)
		return nil, nil, err
	}

	apiKey := keyProvider.Next()

	req := api.Request{
		APIKey:     apiKey,
		SearchType: searchType,
		Mode:       mode,
		Punycode:   true,
		BasicSearchTerms: api.BasicSearchTerms{
			Include: []string{term},
			Exclude: exclude,
		},
	}

	logger.Debugf("worker=%d term=%q mode=%q searchType=%q exclude=%d", workerID, term, mode, searchType, len(exclude))

	resp, httpStatus, err := client.Search(ctx, req)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) {
			human := api.HumanHTTPError(httpErr.StatusCode)
			line := output.FormatErrorLine(stdoutPrinter.Colors(), term, mode, searchType, "http_error", human)
			stdoutPrinter.Println(line)
			return nil, nil, err
		}

		line := output.FormatErrorLine(stdoutPrinter.Colors(), term, mode, searchType, "request_error", err.Error())
		stdoutPrinter.Println(line)
		return nil, nil, err
	}

	if httpStatus != 200 {
		human := api.HumanHTTPError(httpStatus)
		line := output.FormatErrorLine(stdoutPrinter.Colors(), term, mode, searchType, "http_error", human)
		stdoutPrinter.Println(line)
		return nil, nil, fmt.Errorf("unexpected http status: %d", httpStatus)
	}

	// Preview: stdout prints domainsCount; output file stores "term<TAB>count"
	if mode == "preview" {
		line := output.FormatKeyValueLine(stdoutPrinter.Colors(), term, mode, searchType, "domainsCount", fmt.Sprintf("%d", resp.DomainsCount))
		outItem := fmt.Sprintf("%s\t%d", term, resp.DomainsCount)
		return []string{line}, []string{outItem}, nil
	}

	// Purchase: stdout prints [domain: X]; output file stores "X" (clean)
	lines := make([]string, 0, len(resp.DomainsList))
	items := make([]string, 0, len(resp.DomainsList))

	for _, d := range resp.DomainsList {
		lines = append(lines, output.FormatKeyValueLine(stdoutPrinter.Colors(), term, mode, searchType, "domain", d))
		items = append(items, d)
	}

	// If provider returns purchase with no domainsList, still emit count line; file stores term\tcount.
	if len(lines) == 0 {
		lines = append(lines, output.FormatKeyValueLine(stdoutPrinter.Colors(), term, mode, searchType, "domainsCount", fmt.Sprintf("%d", resp.DomainsCount)))
		items = append(items, fmt.Sprintf("%s\t%d", term, resp.DomainsCount))
	}

	return lines, items, nil
}
