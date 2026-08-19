package cmd

import (
	"context"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployDestinationCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "deploy:destination -i <input file path> [-o <output file path>] [-f <format>]",
	Short: "Deploy a destination",
	Long:  "Create or update a destination from a JSON or YAML file.",
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()

		input := viper.GetString("input")
		if input == "" {
			logger.Fatal("Must provide an input file path")
		}

		var dest request.Destination

		_, err := utils.ReadStructFromFile(input, &dest)
		if err != nil {
			logger.FatalErr("Unable to read destination file", err)
		}

		client := request.NewAPIClient(projectId, &apiKey)
		oldDest := getOldDest(cmd.Context(), client, &dest)

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			logger.FatalErr("Unable to read output format", err)
		}

		var output *request.Destination

		if oldDest == nil {
			output, err = client.CreateDestination(cmd.Context(), &dest)
		} else {
			patch := generatePatch(oldDest, &dest)
			if len(patch.UpdateMask) == 0 {
				err := utils.WriteStructToFile(viper.GetString("output"),
					utils.Format(format), oldDest)
				if err != nil {
					logger.FatalErr("Unable to write destination file", err)
				}

				return
			}

			output, err = client.PatchDestination(cmd.Context(), oldDest.Id, patch)
		}

		if err != nil {
			logger.FatalErr("Unable to deploy destination", err)
		}

		err = utils.WriteStructToFile(viper.GetString("output"),
			utils.Format(format), output)
		if err != nil {
			logger.FatalErr("Unable to write destination file", err)
		}
	},
}

func generatePatch(oldDest *request.Destination, newDest *request.Destination) *request.PatchDestination {
	patch := &request.PatchDestination{
		Destination: make(map[string]any),
	}

	if oldDest.Name != newDest.Name {
		patch.Destination["name"] = newDest.Name
		patch.UpdateMask = append(patch.UpdateMask, "name")
	}

	if oldDest.Type != newDest.Type {
		patch.Destination["type"] = newDest.Type
		patch.UpdateMask = append(patch.UpdateMask, "type")
	}

	if newDest.Metadata != nil && (oldDest.Metadata == nil || oldDest.Metadata.URL != newDest.Metadata.URL) {
		patch.Destination["metadata"] = map[string]any{"url": newDest.Metadata.URL}
		patch.UpdateMask = append(patch.UpdateMask, "metadata.url")
	}

	return patch
}

func getOldDest(ctx context.Context, client *request.APIClient, dest *request.Destination) *request.Destination {
	destinations, err := client.ListDestinations(ctx)
	if err != nil {
		logger.FatalErr("Unable to list destinations", err)
	}

	for _, existing := range destinations {
		if dest.Id != "" && existing.Id == dest.Id {
			return existing
		}

		if dest.Id == "" && existing.Name == dest.Name {
			return existing
		}
	}

	return nil
}

func init() {
	deployDestinationCmd.Flags().StringP("input", "i", "", "The input file path")

	err := viper.BindPFlag("input", deployDestinationCmd.Flags().Lookup("input"))
	if err != nil {
		logger.FatalErr("unable to bind flag", err)
	}

	deployDestinationCmd.Flags().StringP("output", "o", "-", "The output file path")

	err = viper.BindPFlag("output", deployDestinationCmd.Flags().Lookup("output"))
	if err != nil {
		logger.FatalErr("unable to bind flag", err)
	}

	err = flags.InitAndBindFormatFlag(deployDestinationCmd)
	if err != nil {
		logger.FatalErr("unable to bind flag", err)
	}

	rootCmd.AddCommand(deployDestinationCmd)
}
