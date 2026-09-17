// Package region resolves which Ampersand deployment region the CLI talks to, and
// rewrites Ampersand URLs into their equivalent in that region.
package region

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/amp-labs/cli/appdata"
)

// Region is an Ampersand deployment region.
type Region string

const (
	US Region = "us" //nolint:varnamelen // read package-qualified, as region.US
	EU Region = "eu" //nolint:varnamelen // read package-qualified, as region.EU

	// Default is the region used when none has been selected.
	Default = US
)

// FlagName is the name of the persistent flag and viper key that select the region.
// The flags package binds both; see flags.GetRegion.
const FlagName = "region"

// EnvVar is the environment variable that selects the region when neither --region nor
// 'amp set:region' has. It is deliberately not bound to viper, which would rank it above the
// saved region; Resolve reads it directly instead.
const EnvVar = "AMP_REGION"

// baseDomain is the top-level domain that every Ampersand hostname sits under.
const baseDomain = "withampersand.com"

// known is the set of regions supported at the time of this build of the CLI.
var known = map[Region]struct{}{ //nolint:gochecknoglobals
	US: {},
	EU: {},
}

// Known lists the regions supported at the time of this build of the CLI.
// Default ("us") first and the rest sorted. It is for help text and warnings only.
// Parse accepts any region name, so when new regions launch users can set them without needing
// to upgrade to a newer CLI build.
func Known() []string {
	others := make([]string, 0, len(known))

	for region := range known {
		if region == Default {
			continue
		}

		others = append(others, string(region))
	}

	sort.Strings(others)

	return append([]string{string(Default)}, others...)
}

// IsKnown reports whether this region was supported at the time this CLI was built. See Known.
func IsKnown(region Region) bool {
	_, ok := known[region]

	return ok
}

// ErrInvalid is returned when a region name could not be used as a DNS label.
var ErrInvalid = errors.New("invalid region name")

// labelPattern is what a region name must look like to serve as a DNS label: lowercase
// letters, digits and hyphens, starting with a letter and not ending in a hyphen.
var labelPattern = regexp.MustCompile(`^[a-z](?:[a-z0-9-]*[a-z0-9])?$`) //nolint:gochecknoglobals

// maxLabelLen is the DNS limit on the length of a single label.
const maxLabelLen = 63

// Parse normalizes a region name, and checks that it is a well-formed DNS label.
// It deliberately does not check the name against Known, so that as new regions are supported,
// users can set them without having to upgrade the build of their CLI.
//
// An empty name is not valid: callers that want the fallback should use Resolve.
func Parse(name string) (Region, error) {
	region := Region(strings.ToLower(strings.TrimSpace(name)))

	if len(region) > maxLabelLen || !labelPattern.MatchString(string(region)) {
		return "", fmt.Errorf("%w: %q (e.g. %s)", ErrInvalid, name, strings.Join(Known(), ", "))
	}

	return region, nil
}

// RegionalizeURL rewrites an Ampersand hostname into its equivalent in this region, eg:
//
//	clerk.withampersand.com         ->      clerk.eu.withampersand.com
//	cli-signin.withampersand.com    ->      cli-signin.eu.withampersand.com
//	staging-api.withampersand.com   ->      staging-api.eu.withampersand.com
func (region Region) RegionalizeURL(rawURL string) string {
	label := region.label()
	if label == "" {
		// if US region, no <region> label, return original URL
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		// on parsing error, fallback to original URL
		return rawURL
	}

	baseURLSuffix := "." + baseDomain

	host := parsed.Hostname()
	if !strings.HasSuffix(host, baseURLSuffix) {
		// if not a *.withampersand.com URL, return original URL
		return rawURL
	}

	prefix := strings.TrimSuffix(host, baseURLSuffix)
	if prefix == label || strings.HasSuffix(prefix, "."+label) {
		// if already regionalized, return original URL
		return rawURL
	}

	// otherwise, regionalize the *.withampersand.com URL
	regionalHost := prefix + "." + label + "." + baseDomain
	if port := parsed.Port(); port != "" {
		regionalHost = net.JoinHostPort(regionalHost, port)
	}

	parsed.Host = regionalHost

	return parsed.String()
}

// label is the DNS label that identifies the region in an Ampersand hostname.
//
// The US is the original, label-less deployment (api.withampersand.com), so it has none.
// Every other region inserts its name ahead of the base domain (api.eu.withampersand.com).
func (region Region) label() string {
	if region == US {
		return ""
	}

	return string(region)
}

// Resolve returns the region to use, by priority:
//  1. --region
//  2. 'amp set:region' (config/appdata file)
//  3. AMP_REGION
//  4. region.Default ("us")
//
// The saved region deliberately outranks AMP_REGION, unlike viper's usual env-over-config order.
//
// A missing config file is not an error: it means no region has been saved yet.
// A malformed region name, or a config file that exists but cannot be read, is.
func Resolve(flagValue string) (Region, error) {
	// --region invokes this function with the provided value
	if flagValue != "" {
		return Parse(flagValue)
	}

	// otherwise, check if the config file has a region set
	config, err := appdata.Get()
	if err != nil {
		return "", fmt.Errorf("could not read the saved region: %w", err)
	}

	if config.Region != "" {
		return Parse(config.Region)
	}

	// otherwise, check the environment variable
	if fromEnv := os.Getenv(EnvVar); fromEnv != "" {
		return Parse(fromEnv)
	}

	// otherwise, default to US
	return Default, nil
}
