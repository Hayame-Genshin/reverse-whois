package cli

import (
	"fmt"
	"io"
)

func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, "reverse-whois - Reverse WHOIS OSINT CLI (purchase/preview)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  reverse-whois --api-key <KEY> -t <term> [options]")
	fmt.Fprintln(w, "  cat targets.txt | reverse-whois --api-key <KEY> [options]")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Targets:")
	fmt.Fprintln(w, "  --target, -t <target>           Define the search term (repeatable; supports comma-separated values)")
	fmt.Fprintln(w, "  --target-list, -tL <file>       Define a file containing targets (repeatable; supports comma-separated paths)")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "API key:")
	fmt.Fprintln(w, "  --api-key <API_KEY>             API key (highest priority over .reverse-whois.yaml)")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Query options:")
	fmt.Fprintln(w, "  --history, --historic           searchType=historic (default: current)")
	fmt.Fprintln(w, "  --preview, --check              mode=preview (default: purchase)")
	fmt.Fprintln(w, "  --exclude, -e <term>            Add exclude term(s) (repeatable; comma-separated; max 4)")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Networking / TLS:")
	fmt.Fprintln(w, "  --proxy <url>                   Proxy URL (http://, https://, socks5://)")
	fmt.Fprintln(w, "  --no-proxy                      Disable proxy usage even if env vars exist")
	fmt.Fprintln(w, "  --insecure, -k                  Skip TLS certificate verification (unsafe)")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Concurrency / rate limit:")
	fmt.Fprintln(w, "  --threads <n>                   Worker threads (default: 1)")
	fmt.Fprintln(w, "  --rate-limit, -rl <rps>         Global max requests/second (max: 30)")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Output:")
	fmt.Fprintln(w, "  --output, -o, -out <path>       Write results to a file (overwrite by default)")
	fmt.Fprintln(w, "  --silent, -s                    Results only (no logs)")
	fmt.Fprintln(w, "  --quiet, -q                     Results only (no logs)")
	fmt.Fprintln(w, "  --verbose, -v                   Enable debug logs (stderr)")
	fmt.Fprintln(w, "  --debug                         Enable debug logs (stderr)")
	fmt.Fprintln(w, "  --no-color, -nc                 Disable ANSI colors")
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "Help:")
	fmt.Fprintln(w, "  --help, -h                      Show this help message")
	fmt.Fprintln(w, "")
}
