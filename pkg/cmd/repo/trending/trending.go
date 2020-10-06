package trending

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/api"
	"github.com/cli/cli/internal/config"
	"github.com/cli/cli/internal/ghinstance"
	sharedRepo "github.com/cli/cli/pkg/cmd/repo/shared"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

type TrendingOptions struct {
	HttpClient func() (*http.Client, error)
	Config     func() (config.Config, error)
	IO         *iostreams.IOStreams

	Range    string
	Language string
	Method   string
}

func NewCmdTrending(f *cmdutil.Factory, runF func(*TrendingOptions) error) *cobra.Command {
	opts := &TrendingOptions{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
		Config:     f.Config,

		Method: "created",
	}

	cmd := &cobra.Command{
		Use:   "trending",
		Short: "Trending repositories",
		Long:  `List of trending repositories`,
		Example: heredoc.Doc(`
			$ gh repo trending
			$ gh repo trending -l Go -r week
	  `),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return trendingRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Range, "range", "r", "day", "day|week|month|year")
	cmd.Flags().StringVarP(&opts.Language, "language", "l", "", "Filter by language. Ex. Go")

	return cmd
}

func trendingRun(opts *TrendingOptions) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}

	apiClient := api.NewClientFromHTTP(httpClient)

	dateRange, err := dateRangeGenerator(opts.Method, opts.Range)
	if err != nil {
		return err
	}

	isTerminal := opts.IO.IsStdoutTTY()
	repos, total, err := api.GitHubTrending(apiClient, dateRange, opts.Language, ghinstance.Default())
	if err != nil {
		return err
	}

	err = opts.IO.StartPager()
	if err != nil {
		return err
	}
	defer opts.IO.StopPager()

	if isTerminal {
		title := fmt.Sprintf("Trending repositories (by %s)", opts.Range)
		fmt.Fprintf(opts.IO.Out, "\n%s\n\n", title)
	}

	sharedRepo.PrintRepositories(opts.IO, "", (total - len(repos)), repos)

	return nil
}

// the dateRangeGenerator generates the range by flag
func dateRangeGenerator(method, r string) (string, error) {
	now := time.Now()
	rounded := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	formatedToday := rounded.Format(time.RFC3339Nano)
	var rangeFrom string

	switch r {
	case "day":
		rangeFrom = rounded.Add(-24 * time.Hour).Format(time.RFC3339Nano)
	case "week":
		rangeFrom = rounded.Add(6 * -24 * time.Hour).Format(time.RFC3339Nano)
	case "month":
		rangeFrom = rounded.Add(30 * -24 * time.Hour).Format(time.RFC3339Nano)
	case "year":
		rangeFrom = rounded.Add(364 * -24 * time.Hour).Format(time.RFC3339Nano)
	default:
		return "", fmt.Errorf("search range is not correct:'%s'; try day, week, month or year", r)
	}

	return fmt.Sprintf("%s:%s..%s", method, rangeFrom, formatedToday), nil
}
