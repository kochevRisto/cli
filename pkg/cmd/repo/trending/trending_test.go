package trending

import (
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/internal/config"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/httpmock"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/test"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
)

func initFakeHTTP() *httpmock.Registry {
	return &httpmock.Registry{}
}

func runTrendingCommand(httpClient *http.Client, cli string) (*test.CmdOut, error) {

	io, stdin, stdout, stderr := iostreams.Test()
	fac := &cmdutil.Factory{
		IOStreams: io,
		HttpClient: func() (*http.Client, error) {
			return httpClient, nil
		},
		Config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
	}

	cmd := NewCmdTrending(fac, nil)

	argv, err := shlex.Split(cli)
	cmd.SetArgs(argv)

	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	if err != nil {
		panic(err)
	}

	_, err = cmd.ExecuteC()
	return &test.CmdOut{OutBuf: stdout, ErrBuf: stderr}, err
}

func TestTrending(t *testing.T) {
	reg := initFakeHTTP()
	defer reg.Verify(t)

	reg.Register(
		httpmock.GraphQL(`query TrendingRepositories\b`),
		httpmock.FileResponse("./fixtures/trending.json"),
	)

	httpClient := &http.Client{Transport: reg}
	output, err := runTrendingCommand(httpClient, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		heredoc.Doc(`IEEE-VIT/termiboard
			  https://github.com/IEEE-VIT/termiboard
			  A smart CLI Dashboard to fetch cpu, memory and network stats!
			  17 stars | 11 forks | Go

			tcard/enumtag
			  https://github.com/tcard/enumtag
			  Package enumtag.
			  5 stars | 0 forks | Go

		`),
		output.String(),
	)
	assert.Equal(t, ``, output.Stderr())
}
func TestTrendingRequest(t *testing.T) {
	reg := initFakeHTTP()
	defer reg.Verify(t)

	reg.Register(
		httpmock.GraphQL(`query TrendingRepositories\b`),
		httpmock.GraphQLQuery(`{}`, func(_ string, params map[string]interface{}) {
			assert.Equal(t, float64(30), params["first"])
			assert.Equal(t, "REPOSITORY", params["type"])
			assert.NotEqual(t, "", params["query"])
		}),
	)

	httpClient := &http.Client{Transport: reg}

	output, err := runTrendingCommand(httpClient, "")
	if err != nil {
		t.Fatalf("error running command `repo trending`: %v", err)
	}

	assert.Equal(t, "", output.String())
	assert.Equal(t, "", output.Stderr())
}

func TestTrending_withInvalidRangeFlag(t *testing.T) {
	reg := initFakeHTTP()
	defer reg.Verify(t)

	httpClient := &http.Client{Transport: reg}

	_, err := runTrendingCommand(httpClient, "-r abc")
	if err != nil && err.Error() != "search range is not correct:'abc'; try day, week, month or year" {
		t.Fatalf("error running command `repo trending`: %v", err)
	}
}
