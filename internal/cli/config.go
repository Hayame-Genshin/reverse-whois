package cli

type Config struct {
	Targets     []string
	TargetLists []string

	APIKey string

	Historic     bool
	Preview      bool
	ExcludeTerms []string

	Silent  bool
	Quiet   bool
	Verbose bool
	Debug   bool
	NoColor bool

	OutputPath string

	ProxyURL  string
	NoProxy   bool
	Insecure  bool
	Threads   int
	RateLimit int
}
