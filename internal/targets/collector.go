package targets

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/haltman-io/reverse-whois/internal/util"
)

type CollectOptions struct {
	TargetValues []string
	ListFiles    []string
	ReadStdin    bool
	Logger       *util.Logger
}

func CollectTargets(opts CollectOptions) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string

	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	// stdin
	if opts.ReadStdin && util.HasStdinData() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			add(line)
		}
		if err := sc.Err(); err != nil {
			return nil, fmt.Errorf("failed reading stdin: %w", err)
		}
		if opts.Logger != nil {
			opts.Logger.Debugf("collected %d target(s) from stdin", len(out))
		}
	}

	// list files
	for _, filePath := range opts.ListFiles {
		paths := util.SplitCSV(filePath)
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			lines, err := util.ReadLines(p)
			if err != nil {
				return nil, fmt.Errorf("failed reading target list %q: %w", p, err)
			}
			for _, l := range lines {
				add(l)
			}
		}
	}

	// direct targets
	for _, t := range opts.TargetValues {
		for _, p := range util.SplitCSV(t) {
			add(p)
		}
	}

	return out, nil
}
