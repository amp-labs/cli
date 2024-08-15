package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/openapi"
	"github.com/amp-labs/cli/request"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var ErrBadInput = errors.New("bad input")

const NumProvidersShown = 10

func nonEmpty(fieldName string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%w: %s cannot be empty", ErrBadInput, fieldName)
		}

		return nil
	}
}

var initCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "init",
	Short: "create a new Ampersand project",
	Long:  "Create a new Ampersand project.",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := promptString("Name your integration", nonEmpty("Integration name"))
		if err != nil {
			logger.FatalErr("Unable to prompt for integration name", err)
		}

		provider, err := selectProvider(cmd.Context())
		if err != nil {
			logger.FatalErr("Unable to select provider", err)
		}

		disp := "my integration"

		var integ openapi.Integration
		integ.Name = name
		integ.DisplayName = &disp
		integ.Provider = provider.Name

		if provider.Support.Proxy {
			if err := setupProxy(&integ, provider); err != nil {
				logger.FatalErr("Unable to setup proxy", err)
			}
		}

		if provider.Support.Read {
			if err := setupRead(&integ, provider); err != nil {
				logger.FatalErr("Unable to setup read", err)
			}
		}

		if provider.Support.Write {
			if err := setupWrite(&integ, provider); err != nil {
				logger.FatalErr("Unable to setup write", err)
			}
		}

		path, err := promptString("Save YAML to file path", nonEmpty("File path"))
		if err != nil {
			logger.FatalErr("Unable to prompt for file path", err)
		}

		manifest := &openapi.Manifest{
			SpecVersion: "1.0.0",
			Integrations: []openapi.Integration{
				integ,
			},
		}

		ys, err := yaml.Marshal(manifest)
		if err != nil {
			logger.FatalErr("Unable to marshal manifest", err)
		}

		if err := os.WriteFile(path, ys, 0o644); err != nil { //nolint:gosec
			logger.FatalErr("Unable to write manifest to file", err)
		}

		fmt.Println("Integration manifest written to", path)
	},
}

func addWriteObject(write *openapi.IntegrationWrite, provider *openapi.ProviderInfo) error {
	obj := &openapi.IntegrationWriteObject{}

	objName, err := promptString(provider.DisplayName + " object name")
	if err != nil {
		return err
	}

	obj.ObjectName = objName

	if write.Objects == nil {
		tmp := make([]openapi.IntegrationWriteObject, 0)
		write.Objects = &tmp
	}

	*write.Objects = append(*write.Objects, *obj)

	return nil
}

func setupWrite(integ *openapi.Integration, provider *openapi.ProviderInfo) error {
	wantWrite, err := promptBool("Enable write support for " + provider.DisplayName)
	if err != nil {
		return err
	}

	if !wantWrite {
		return nil
	}

	write := &openapi.IntegrationWrite{}

	for {
		if err := addWriteObject(write, provider); err != nil {
			return err
		}

		another, err := promptBool("Add another object")
		if err != nil {
			return err
		}

		if !another {
			break
		}
	}

	integ.Write = write

	return nil
}

func setupBackfill(obj *openapi.IntegrationObject) error {
	full, err := promptBool("Backfill full history for " + obj.ObjectName)
	if err != nil {
		return err
	}

	if full {
		obj.Backfill = &openapi.Backfill{
			DefaultPeriod: openapi.DefaultPeriod{
				FullHistory: &full,
			},
		}

		return nil
	}

	days, err := promptInt("Number of days to backfill for " + obj.ObjectName)
	if err != nil {
		return err
	}

	if days > 0 {
		tmp := int(days)
		obj.Backfill = &openapi.Backfill{
			DefaultPeriod: openapi.DefaultPeriod{
				Days: &tmp,
			},
		}
	}

	return nil
}

func getIntegrationField() (*openapi.IntegrationField, error) { //nolint:funlen,cyclop
	field := &openapi.IntegrationField{}

	fieldName, err := promptString("Field name")
	if err != nil {
		return nil, err
	}

	remap, err := promptBool("Do you want to let users choose a different name for this field")
	if err != nil {
		return nil, err
	}

	if !remap {
		if err := field.FromIntegrationFieldExistent(openapi.IntegrationFieldExistent{
			FieldName: fieldName,
		}); err != nil {
			return nil, err
		}

		return field, nil
	}

	displayName, err := promptString("Friendly display name for " + fieldName)
	if err != nil {
		return nil, err
	}

	wantDefault, err := promptBool("Do you want to provide a default suggested value for this field")
	if err != nil {
		return nil, err
	}

	var defaultValue string
	if wantDefault {
		defaultValue, err = promptString("Default value for " + fieldName)
		if err != nil {
			return nil, err
		}
	}

	wantPrompt, err := promptBool("Do you want to provide custom prompt text for this field")
	if err != nil {
		return nil, err
	}

	var promptText string
	if wantPrompt {
		promptText, err = promptString("Prompt text for " + fieldName)
		if err != nil {
			return nil, err
		}
	}

	mapping := openapi.IntegrationFieldMapping{
		MapToName:        fieldName,
		MapToDisplayName: &displayName,
	}

	if wantDefault {
		mapping.Default = &defaultValue
	}

	if wantPrompt {
		mapping.Prompt = &promptText
	}

	if err := field.FromIntegrationFieldMapping(mapping); err != nil {
		return nil, err
	}

	return field, nil
}

func addReadObject(read *openapi.IntegrationRead, provider *openapi.ProviderInfo) error { //nolint:funlen,cyclop,gocognit
	obj := &openapi.IntegrationObject{}

	objName, err := promptString(provider.DisplayName + " object name")
	if err != nil {
		return err
	}

	obj.ObjectName = objName

	destination, err := promptString("Destination for " + objName)
	if err != nil {
		return err
	}

	obj.Destination = destination

	schedule, err := promptString("Cron schedule for " + objName)
	if err != nil {
		return err
	}

	obj.Schedule = schedule

	wantBackfill, err := promptBool("Configure backfill period for " + objName)
	if err != nil {
		return err
	}

	if wantBackfill {
		if err := setupBackfill(obj); err != nil {
			return err
		}
	}

	counter := 0

	for {
		if counter == 0 { //nolint:nestif
			wantRequired, err := promptBool("Add a required field for " + objName)
			if err != nil {
				return err
			}

			if !wantRequired {
				break
			}
		} else {
			wantRequired, err := promptBool("Add another required field for " + objName)
			if err != nil {
				return err
			}

			if !wantRequired {
				break
			}
		}

		counter++

		field, err := getIntegrationField()
		if err != nil {
			return err
		}

		if obj.RequiredFields == nil {
			tmp := make([]openapi.IntegrationField, 0)
			obj.RequiredFields = &tmp
		}

		*obj.RequiredFields = append(*obj.RequiredFields, *field)
	}

	counter = 0

	for {
		if counter == 0 { //nolint:nestif
			wantOptional, err := promptBool("Add an optional field for " + objName)
			if err != nil {
				return err
			}

			if !wantOptional {
				break
			}
		} else {
			wantOptional, err := promptBool("Add another optional field for " + objName)
			if err != nil {
				return err
			}

			if !wantOptional {
				break
			}
		}

		counter++

		field, err := getIntegrationField()
		if err != nil {
			return err
		}

		if obj.OptionalFields == nil {
			tmp := make([]openapi.IntegrationField, 0)
			obj.OptionalFields = &tmp
		}

		*obj.OptionalFields = append(*obj.OptionalFields, *field)
	}

	wantOptionalAuto, err := promptBool("Allow users to add their own optional fields for " + objName)
	if err != nil {
		return err
	}

	if wantOptionalAuto {
		tmp := openapi.All
		obj.OptionalFieldsAuto = &tmp
	}

	if read.Objects == nil {
		tmp := make([]openapi.IntegrationObject, 0)
		read.Objects = &tmp
	}

	*read.Objects = append(*read.Objects, *obj)

	return nil
}

func setupRead(integ *openapi.Integration, provider *openapi.ProviderInfo) error {
	wantRead, err := promptBool("Enable read support for " + provider.DisplayName)
	if err != nil {
		return err
	}

	if !wantRead {
		return nil
	}

	read := &openapi.IntegrationRead{}

	for {
		if err := addReadObject(read, provider); err != nil {
			return err
		}

		another, err := promptBool("Add another object")
		if err != nil {
			return err
		}

		if !another {
			break
		}
	}

	integ.Read = read

	return nil
}

func setupProxy(integ *openapi.Integration, provider *openapi.ProviderInfo) error {
	wantProxy, err := promptBool("Enable proxy support for " + provider.DisplayName)
	if err != nil {
		return err
	}

	if wantProxy {
		integ.Proxy = &openapi.IntegrationProxy{
			Enabled: &wantProxy,
		}
	}

	return nil
}

func promptString(prompt string, validate ...func(string) error) (string, error) {
	prompter := promptui.Prompt{
		Label: prompt,
		Validate: func(s string) error {
			for _, fn := range validate {
				if err := fn(s); err != nil {
					return err
				}
			}

			return nil
		},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	return prompter.Run()
}

func promptBool(prompt string) (bool, error) {
	prompter := promptui.Prompt{
		Label:     prompt,
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

func promptInt(prompt string) (int64, error) {
	prompter := promptui.Prompt{
		Label: prompt,
		Validate: func(s string) error {
			_, err := strconv.ParseInt(s, 10, 64)

			return err
		},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	std, err := prompter.Run()
	if err != nil {
		return 0, err
	}

	ival, err := strconv.ParseInt(std, 10, 64)
	if err != nil {
		return 0, err
	}

	return ival, nil
}

func selectProvider(ctx context.Context) (*openapi.ProviderInfo, error) {
	apiKey := flags.GetAPIKey()

	client := request.NewAPIClient("ignore", &apiKey)

	cat, err := client.GetCatalog(ctx)
	if err != nil {
		logger.FatalErr("Unable to get catalog", err)
	}

	providers := make([]openapi.ProviderInfo, 0, len(cat))
	for _, provider := range cat {
		providers = append(providers, provider)
	}

	sort.Slice(providers, func(i, j int) bool {
		return providers[i].DisplayName < providers[j].DisplayName
	})

	prompt := promptui.Select{
		Label:             "Select a provider",
		Items:             providers,
		Size:              NumProvidersShown,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			provider := providers[index]

			return input == "" ||
				strings.Contains(provider.DisplayName, input) ||
				strings.Contains(provider.Name, input)
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ .DisplayName }}",
			Active:   "& {{ .DisplayName | cyan }}",
			Inactive: "  {{ .DisplayName | cyan }}",
			Selected: "Selected {{ .DisplayName | red }}",
		},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &providers[idx], nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
