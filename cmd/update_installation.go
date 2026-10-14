package cmd

import (
	"errors"
	"os"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
)

var updateInstallationInput string //nolint:gochecknoglobals

var updateInstallationCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "update:installation <integrationId> <installationId> --input <path>",
	Short: "Update an installation",
	Long:  "Update explicit installation fields from a JSON or YAML patch containing installation and updateMask.",
	Args:  cobra.ExactArgs(2), //nolint:mnd
	Run: func(cmd *cobra.Command, args []string) {
		var patch request.PatchInstallation

		_, err := utils.ReadStructFromFile(updateInstallationInput, &patch)
		if err != nil {
			logger.FatalErr("Unable to read installation patch", err)
		}

		if len(patch.Installation) == 0 || len(patch.UpdateMask) == 0 {
			logger.Fatal("Installation patch must contain installation and updateMask")
		}

		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()
		client := request.NewAPIClient(projectId, &apiKey)

		installation, err := client.PatchInstallation(cmd.Context(), args[0], args[1], &patch)
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to update installation", err)
			}
		}

		err = utils.WriteStruct(os.Stdout, flags.GetOutputFormatForCommand(cmd), summarizeInstallation(installation))
		if err != nil {
			logger.FatalErr("Unable to write installation", err)
		}
	},
}

func init() {
	updateInstallationCmd.Flags().StringVarP(
		&updateInstallationInput, "input", "i", "", "Path to a JSON or YAML installation patch, or - for stdin",
	)

	err := updateInstallationCmd.MarkFlagRequired("input")
	if err != nil {
		logger.FatalErr("unable to require input flag", err)
	}

	err = flags.InitAndBindFormatFlag(updateInstallationCmd)
	if err != nil {
		logger.FatalErr("unable to initialize flags", err)
	}

	rootCmd.AddCommand(updateInstallationCmd)
}
