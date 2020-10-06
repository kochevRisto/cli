package shared

import (
	"fmt"
	"strconv"

	"github.com/cli/cli/api"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/utils"
)

// PrintRepositories generates TablePrinter
func PrintRepositories(io *iostreams.IOStreams, prefix string, totalCount int, repos []api.Repository) {
	table := utils.NewTablePrinter(io)

	for _, repo := range repos {
		r := repo.Owner.Login + "/" + repo.Name
		table.AddField(r, nil, utils.Green)
		table.EndRow()
		table.AddField("  "+repo.URL, nil, nil)
		table.EndRow()
		table.AddField("  "+repo.Description, nil, utils.Gray)
		table.EndRow()

		// This must be one-liner beacuse of the table.Render error.
		// It expects only one AddField call (index out of range [1] with length 1)
		s := fmt.Sprintf(
			"  %s stars | %s forks | %s",
			strconv.Itoa(repo.StargazerCount),
			strconv.Itoa(repo.ForkCount),
			repo.PrimaryLanguage.Name,
		)
		table.AddField(s, nil, utils.Blue)
		table.EndRow()

		table.AddField("", nil, nil)
		table.EndRow()
	}

	_ = table.Render()

	isTerminal := io.IsStdoutTTY()
	remaining := totalCount - len(repos)
	if remaining > 0 && isTerminal {
		fmt.Fprintf(io.Out, utils.Gray("%sAnd %d more\n"), prefix, remaining)
	}
}
