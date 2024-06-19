package cmd

import (
	"context"
	"errors"
	"reflect"

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
	Long:  "Deploy a destination",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := flags.GetAPIKey()

		projectId := flags.GetProjectId()
		if projectId == "" {
			logger.Fatal("Must provide a project ID in the --project flag")
		}

		input := viper.GetString("input")
		if input == "" {
			logger.Fatal("Must provide an input file path")
		}

		var dest request.Destination
		if _, err := utils.ReadStructFromFile(input, &dest); err != nil {
			logger.FatalErr("Unable to read destination file", err)
		}

		client := request.NewAPIClient(projectId, &apiKey)
		oldDest := getOldDest(cmd.Context(), client, &dest)

		var output *request.Destination
		var err error

		if oldDest == nil {
			output, err = client.CreateDestination(cmd.Context(), &dest)
		} else {
			patch := generatePatch(oldDest, &dest)
			if len(patch.UpdateMask) == 0 {
				if err := utils.WriteStructToFile(viper.GetString("output"),
					flags.GetOutputFormat(), oldDest); err != nil {
					logger.FatalErr("Unable to write destination file", err)
				}

				return
			}

			output, err = client.PatchDestination(cmd.Context(), oldDest.Id, patch)
		}

		if err != nil {
			logger.FatalErr("Unable to deploy destination", err)
		}

		if err := utils.WriteStructToFile(viper.GetString("output"),
			flags.GetOutputFormat(), output); err != nil {
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

	if oldDest.Metadata != nil { //nolint:nestif
		if newDest.Metadata == nil {
			patch.Destination["metadata"] = nil
			patch.UpdateMask = append(patch.UpdateMask, "metadata")
		} else if !reflect.DeepEqual(oldDest.Metadata, newDest.Metadata) {
			patch.Destination["metadata"] = newDest.Metadata
			patch.UpdateMask = append(patch.UpdateMask, "metadata")
		}
	} else {
		if newDest.Metadata != nil {
			patch.Destination["metadata"] = newDest.Metadata
			patch.UpdateMask = append(patch.UpdateMask, "metadata")
		}
	}

	return patch
}

func getOldDest(ctx context.Context, client *request.APIClient, dest *request.Destination) *request.Destination {
	id := findDestId(ctx, client, dest)
	if id == "" {
		return nil
	}

	dst, err := client.GetDestination(ctx, id)
	if err != nil {
		if errors.Is(err, request.ErrNotFound) {
			return nil
		} else {
			logger.FatalErr("Unable to get destination", err)
		}
	}

	return dst
}

func findDestId(ctx context.Context, client *request.APIClient, dest *request.Destination) string {
	if dest.Id != "" {
		return dest.Id
	}

	dests, err := client.ListDestinations(ctx)
	if err != nil {
		logger.FatalErr("Unable to list destinations", err)
	}

	for _, d := range dests {
		if d.Name == dest.Name {
			return d.Id
		}
	}

	return ""
}

func init() {
	deployDestinationCmd.Flags().StringP("input", "i", "", "The input file path")

	if err := viper.BindPFlag("input", deployDestinationCmd.Flags().Lookup("input")); err != nil {
		logger.FatalErr("unable to bind flag", err)
	}

	deployDestinationCmd.Flags().StringP("output", "o", "-", "The output file path")

	if err := viper.BindPFlag("output", deployDestinationCmd.Flags().Lookup("output")); err != nil {
		logger.FatalErr("unable to bind flag", err)
	}

	rootCmd.AddCommand(deployDestinationCmd)
}
