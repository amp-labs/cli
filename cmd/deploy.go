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
	"github.com/gertd/go-pluralize"
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
				fmt.Fprint(os.Stdout, err.Error()+"\n")
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

		// Before deploying, check if any read objects were removed from the manifest. If so, prompt the user to decide
		// whether to pause scheduled reads for those objects.
		shouldPauseReads, err := confirmReadObjectRemoval(cmd.Context(), client, zipResult.Manifest)
		if err != nil {
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
			request.BatchUpsertIntegrationsParams{
				SourceZipURL: gcsURL,
				Destructive:  shouldPauseReads,
			})
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
) (bool, error) {
	integrations, err := client.ListIntegrations(ctx)
	if err != nil {
		logger.Debugf("Unable to list integrations to check for removed read objects: %v", err)

		return false, nil
	}

	// Collect all integrations with removed objects
	var integrationsWithRemovedObjects []integrationRemovedObjectsInfo

	for _, newInteg := range newManifest.Integrations {
		info, err := getIntegrationRemovedObjectsInfo(ctx, client, integrations, newInteg)
		if err != nil {
			return false, err
		}
		if info != nil {
			integrationsWithRemovedObjects = append(integrationsWithRemovedObjects, *info)
		}
	}

	// If no integrations have removed objects, nothing to prompt
	if len(integrationsWithRemovedObjects) == 0 {
		return false, nil
	}

	// Prompt once globally for all integrations
	choice, err := promptUserConfirmationGlobal(integrationsWithRemovedObjects)
	if err != nil {
		return false, err
	}

	if choice == choiceCancel {
		return false, errDeploymentCancelled
	}

	return choice == choicePause, nil
}

type integrationRemovedObjectsInfo struct {
	integrationName   string
	removedObjects    []string
	installationCount int
	sampleGroups      []groupInfo
}

func getIntegrationRemovedObjectsInfo(
	ctx context.Context,
	client *request.APIClient,
	existingIntegrations []*request.Integration,
	newInteg openapi.Integration,
) (*integrationRemovedObjectsInfo, error) {
	// If the integration doesn't exist or doesn't have a latest revision, then
	// there's nothing to check for removed read objects.
	existingInteg := findIntegrationByName(existingIntegrations, newInteg.Name)
	if existingInteg == nil || existingInteg.LatestRevision == nil {
		return nil, nil
	}

	removedObjects := files.GetRemovedReadObjects(&existingInteg.LatestRevision.Content, &newInteg)
	if len(removedObjects) == 0 {
		return nil, nil
	}

	installations, err := client.ListInstallations(ctx, existingInteg.Id)
	if err != nil {
		logger.Debugf("Unable to list installations for integration %s: %v", existingInteg.Id, err)
	}

	// Extract group names and refs for display (max 5)
	groups := make([]groupInfo, 0, 5)
	for _, inst := range installations {
		if inst.Group == nil {
			continue
		}

		// Use groupName if available, otherwise default to groupRef
		name := inst.Group.GroupName
		if name == "" {
			name = inst.Group.GroupRef
		}

		groups = append(groups, groupInfo{
			name: name,
			ref:  inst.Group.GroupRef,
		})

		if len(groups) >= 5 {
			break
		}
	}

	return &integrationRemovedObjectsInfo{
		integrationName:   newInteg.Name,
		removedObjects:    removedObjects,
		installationCount: len(installations),
		sampleGroups:      groups,
	}, nil
}

func findIntegrationByName(integrations []*request.Integration, name string) *request.Integration {
	for _, integ := range integrations {
		if integ.Name == name {
			return integ
		}
	}

	return nil
}

type userChoice int

const (
	choicePause userChoice = iota
	choiceKeep
	choiceCancel
)

func promptUserConfirmationGlobal(integrations []integrationRemovedObjectsInfo) (userChoice, error) {
	fmt.Println()
	fmt.Println(formatGlobalPromptMessage(integrations))

	pl := pluralize.NewClient()
	integrationWord := pl.Pluralize("integration", len(integrations), false)

	var items []string
	if len(integrations) == 1 {
		items = []string{
			"Continue reading these objects across all installations & continue deployment",
			"Stop reading these objects across all installations & continue deployment",
			"Cancel current deployment",
		}
	} else {
		items = []string{
			fmt.Sprintf("Continue reading these objects across all installations of these %d %s & continue deployment", len(integrations), integrationWord),
			fmt.Sprintf("Stop reading these objects across all installations of these %d %s & continue deployment", len(integrations), integrationWord),
			"Cancel current deployment",
		}
	}

	selectPrompt := promptui.Select{
		Label:     "Select",
		Items:     items,
		CursorPos: 0, // Default to "Continue reading"
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}

	index, selectedItem, err := selectPrompt.Run()
	if err != nil {
		return choiceCancel, err
	}

	// If they selected cancel, no need to confirm
	if index == 2 {
		fmt.Println()
		return choiceCancel, nil
	}

	// Print their choice and ask for confirmation
	fmt.Printf("\nYou selected: %s\n", selectedItem)

	confirmPrompt := promptui.Prompt{
		Label:     "Confirm (yes/no)",
		IsConfirm: true,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}

	_, err = confirmPrompt.Run()
	if err != nil {
		// User said no or aborted
		fmt.Println()
		return choiceCancel, nil
	}

	fmt.Println()

	// Map the selection to our choices
	switch index {
	case 0:
		return choiceKeep, nil
	case 1:
		return choicePause, nil
	default:
		return choiceCancel, nil
	}
}

func formatGlobalPromptMessage(integrations []integrationRemovedObjectsInfo) string {
	pl := pluralize.NewClient()
	var message string

	if len(integrations) == 1 {
		// Single integration
		info := integrations[0]
		objectList := strings.Join(info.removedObjects, ", ")

		objectWord := pl.Pluralize("object", len(info.removedObjects), false)
		installationWord := pl.Pluralize("installation", info.installationCount, false)

		message = fmt.Sprintf(
			"⚠️  You are removing the following read action %s from integration '%s': %s.\n\n"+
				"   This integration has %d %s.",
			objectWord, info.integrationName, objectList, info.installationCount, installationWord,
		)

		if len(info.sampleGroups) > 0 {
			message += formatAffectedInstallations(info.sampleGroups, info.installationCount)
		}

		message += "\n\n❓ Do you want to stop reading these objects across all installations?"
	} else {
		// Multiple integrations
		integrationWord := pl.Pluralize("integration", len(integrations), false)
		message = fmt.Sprintf("⚠️  You are removing read action objects from %d %s:\n\n", len(integrations), integrationWord)

		for _, info := range integrations {
			objectList := strings.Join(info.removedObjects, ", ")
			installationWord := pl.Pluralize("installation", info.installationCount, false)
			message += fmt.Sprintf("   • %s: %s (%d %s)\n",
				info.integrationName, objectList, info.installationCount, installationWord)
		}

		message += fmt.Sprintf(
			"\n\n❓ Do you want to stop reading these objects across all installations of these %d %s?\n\n"+
				"   Note: To stop reads for some integrations & not all, deploy changes to one integration at a time.",
			len(integrations), integrationWord,
		)
	}

	return message
}

func formatAffectedInstallations(groups []groupInfo, totalCount int) string {
	var result string
	for _, g := range groups {
		result += fmt.Sprintf("\n    - %s (%s)", g.name, g.ref)
	}

	if totalCount > len(groups) {
		result += fmt.Sprintf("\n    - and %d more", totalCount-len(groups))
	}

	return result
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
