package cmd

import (
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
	"github.com/amp-labs/cli/request"
	"github.com/amp-labs/cli/storage"
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

		zippedData, err := files.Zip(args[0])
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

		hash.Write(zippedData)
		md5Bytes := hash.Sum(nil)
		md5String := base64.StdEncoding.EncodeToString(md5Bytes)

		client := request.NewAPIClient(projectId, &apiKey)

		signed, err := client.GetPreSignedUploadURL(cmd.Context(), md5String)
		if err != nil {
			if errors.Is(err, clerk.ErrNoSessions) {
				logger.FatalErr("Authenticated session has expired, please log in using amp login", err)
			} else {
				logger.FatalErr("Unable to get pre-signed upload URL", err)
			}

			logger.FatalErr("Unable to get pre-signed upload URL", err)
		}

		if err := storage.Upload(cmd.Context(), zippedData, signed.URL, md5String); err != nil {
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

func init() {
	rootCmd.AddCommand(deployCmd)
}
