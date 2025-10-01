package cmd

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/amp-labs/cli/clerk"
	"github.com/amp-labs/cli/files"
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/openapi"
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/storage"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:     "deploy <ampYamlSourcePath>",
	Aliases: []string{"deploy:integration"},
	Short:   "Deploy changes to integrations",
	Long:    "Deploy changes to integrations, you can either provide a path to the folder that contains amp.yaml or a path to the file itself", //nolint:lll
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectId := flags.GetProjectOrFail()
		apiKey := flags.GetAPIKey()

		zipResult, err := files.Zip(args[0])
		if err != nil {
			if errors.Is(err, files.ErrBadManifest) {
				fmt.Println(err.Error())
				os.Exit(1)
			} else {
				logger.FatalErr("Unable to zip the source", err)
			}
		}

		// nosemgrep: go.lang.security.audit.crypto.use_of_weak_crypto.use-of-md5
		hash := md5.New() //nolint:gosec

		hash.Write(zipResult.Data)
		md5Bytes := hash.Sum(nil)
		md5String := base64.StdEncoding.EncodeToString(md5Bytes)

		client := request.NewAPIClient(projectId, &apiKey)

		// Before deploying, check if any read objects were removed from the manifest. If so, prompt the user to confirm
		// since this will stop all scheduled reads for those objects.
		if err := confirmReadObjectRemoval(cmd.Context(), client, zipResult.Manifest); err != nil {
			logger.FatalErr("Deployment cancelled", err)
		}

		signed, err := client.GetPreSignedUploadURL(cmd.Context(), md5String)
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to get pre-signed upload URL", err)
			}

			logger.FatalErr("Unable to get pre-signed upload URL", err)
		}

		if err := storage.Upload(cmd.Context(), zipResult.Data, signed.URL, md5String); err != nil {
			logger.FatalErr("Unable to upload to Google Cloud Storage", err)
		}

		if !strings.HasPrefix(signed.Path, "/") {
			signed.Path = "/" + signed.Path
		}

		gcsURL := fmt.Sprintf("gs://%s%s", signed.Bucket, signed.Path)

		logger.Debugf("Uploaded to %v", gcsURL)

		integrations, err := client.BatchUpsertIntegrations(cmd.Context(),
			request.BatchUpsertIntegrationsParams{SourceZipURL: gcsURL})
		if err != nil {
			logger.FatalErr(
				"Unable to deploy integrations, you can run the command again with '--debug' flag to troubleshoot.\n",
				err,
			)
		}

		names := make([]string, len(integrations))
		for idx, i := range integrations {
			names[idx] = i.Name
		}

		if len(names) == 0 {
			logger.Infof("No integrations were found in the source file.\n")
		} else if len(names) == 1 {
			logger.Infof("Successfully deployed your integration %s.\n", names[0])
		} else {
			logger.Infof("Successfully deployed your integrations %s.\n", strings.Join(names, ", "))
		}
	},
}

var errDeploymentCancelled = errors.New("user cancelled deployment")

type groupInfo struct {
	name string
	ref  string
}

func confirmReadObjectRemoval(
	ctx context.Context, client *request.APIClient, newManifest *openapi.Manifest,
) error {
	integrations, err := client.ListIntegrations(ctx)
	if err != nil {
		logger.Debugf("Unable to list integrations to check for removed read objects: %v", err)

		return nil
	}

	for _, newInteg := range newManifest.Integrations {
		if err := checkIntegrationForRemovedObjects(ctx, client, integrations, newInteg); err != nil {
			return err
		}
	}

	return nil
}

func checkIntegrationForRemovedObjects(
	ctx context.Context,
	client *request.APIClient,
	existingIntegrations []*request.Integration,
	newInteg openapi.Integration,
) error {
	// If the integration doesn't exist or doesn't have a latest revision, then
	// there's nothing to check for removed read objects.
	existingInteg := findIntegrationByName(existingIntegrations, newInteg.Name)
	if existingInteg == nil || existingInteg.LatestRevision == nil {
		return nil
	}

	removedObjects := files.GetRemovedReadObjects(&existingInteg.LatestRevision.Content, &newInteg)
	if len(removedObjects) == 0 {
		return nil
	}

	installations, err := client.ListInstallations(ctx, existingInteg.Id)
	if err != nil {
		logger.Debugf("Unable to list installations for integration %s: %v", existingInteg.Id, err)
	}

	confirmed, err := promptUserConfirmation(buildPrompt(newInteg.Name, removedObjects, installations))
	if err != nil {
		return err
	}

	if !confirmed {
		return errDeploymentCancelled
	}

	return nil
}

func findIntegrationByName(integrations []*request.Integration, name string) *request.Integration {
	for _, integ := range integrations {
		if integ.Name == name {
			return integ
		}
	}

	return nil
}

type promptData struct {
	integrationName   string
	removedObjects    []string
	installationCount int
	sampleGroups      []groupInfo
}

func buildPrompt(integrationName string, removedObjects []string, installations []*request.Installation) promptData {
	const maxGroupSamples = 5

	groups := make([]groupInfo, 0, maxGroupSamples)

	for _, inst := range installations {
		if inst.Group == nil {
			continue
		}

		// Use groupName if available, otherwise default to groupRef (which is always available)
		name := inst.Group.GroupName
		if name == "" {
			name = inst.Group.GroupRef
		}

		groups = append(groups, groupInfo{
			name: name,
			ref:  inst.Group.GroupRef,
		})

		if len(groups) >= maxGroupSamples {
			break
		}
	}

	return promptData{
		integrationName:   integrationName,
		removedObjects:    removedObjects,
		installationCount: len(installations),
		sampleGroups:      groups,
	}
}

func promptUserConfirmation(data promptData) (bool, error) {
	fmt.Println(formatPromptMessage(data))

	prompter := promptui.Prompt{
		Label:     "Confirm",
		IsConfirm: true,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}

	_, err := prompter.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func formatPromptMessage(data promptData) string {
	objectList := strings.Join(data.removedObjects, ", ")

	var message string
	if data.installationCount > 0 {
		message = fmt.Sprintf(
			"You are removing the following read action object(s) from integration '%s': %s.\n"+
				"Any active scheduled reads for these objects will be stopped.\n\n"+
				"This integration has %d installation(s).",
			data.integrationName, objectList, data.installationCount,
		)

		if len(data.sampleGroups) > 0 {
			message += formatAffectedInstallations(data.sampleGroups, data.installationCount)
		}

		message += "\n\nDo you still want to deploy a new revision of this integration?"
	} else {
		message = fmt.Sprintf(
			"You are removing the following read action object(s) from integration '%s': %s.\n"+
				"Any active scheduled reads for these objects will be stopped.\n\n"+
				"Do you still want to deploy a new revision of this integration?",
			data.integrationName, objectList,
		)
	}

	return message
}

func formatAffectedInstallations(groups []groupInfo, totalCount int) string {
	var result string
	for _, g := range groups {
		result += fmt.Sprintf("\n - %s (%s)", g.name, g.ref)
	}

	if totalCount > len(groups) {
		result += fmt.Sprintf("\n - and %d more", totalCount-len(groups))
	}

	return result
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
