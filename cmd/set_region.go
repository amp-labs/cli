package cmd

import (
	"strings"

	"github.com/amp-labs/cli/appdata"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/region"
	"github.com/amp-labs/cli/vars"
	"github.com/spf13/cobra"
)

var setRegionCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "set:region <region>",
	Short: `Save the region the CLI should talk to (default "` + string(region.Default) + `")`,
	Long: "Save the region the CLI should talk to, e.g. " + strings.Join(region.Known(), " or ") + ". " +
		"Other names are accepted as-is. Credentials are scoped to the region they were issued for, " +
		"so after switching you may need to run 'amp login' again.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		selected, err := region.Parse(args[0])
		if err != nil {
			logger.FatalErr("Unable to set the region", err)
		}

		err = appdata.Set(appdata.Config{Region: string(selected)})
		if err != nil {
			logger.FatalErr("Unable to save the region", err)
		}

		logger.Infof("Region set to %s.", selected)

		if !region.IsKnown(selected) {
			logger.Infof("Warning: this version of amp does not know region %q; requests will go to %s.",
				selected, selected.Regionalize(vars.ApiURL))
		}
	},
}

func init() {
	rootCmd.AddCommand(setRegionCmd)
}
