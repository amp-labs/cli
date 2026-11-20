package clerk

import (
	"testing"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/region"
)

const (
	stageProd    = "prod"
	stageStaging = "staging"
	stageDev     = "dev"
)

//nolint:paralleltest // mutates viper and the environment
func TestGetJwtFile(t *testing.T) {
	tests := []struct {
		name   string
		region region.Region
		stage  string
		want   string
	}{
		{name: "default prod", region: region.Default, stage: stageProd, want: "amp/jwt.json"},
		{name: "us prod", region: region.US, stage: stageProd, want: "amp/jwt.json"},
		{name: "us staging", region: region.US, stage: stageStaging, want: "amp/jwt-staging.json"},
		{name: "us dev", region: region.US, stage: stageDev, want: "amp/jwt-dev.json"},
		{name: "eu prod", region: region.EU, stage: stageProd, want: "amp/jwt-eu.json"},
		{name: "eu staging", region: region.EU, stage: stageStaging, want: "amp/jwt-eu-staging.json"},
		{name: "eu dev", region: region.EU, stage: stageDev, want: "amp/jwt-eu-dev.json"},
	}

	t.Cleanup(func() { flags.SetRegion(region.Default) })

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("AMP_STAGE_OVERRIDE", testCase.stage)
			flags.SetRegion(testCase.region)

			got := GetJwtFile()
			if got != testCase.want {
				t.Errorf("GetJwtFile() = %q, want %q", got, testCase.want)
			}
		})
	}
}

//nolint:paralleltest // mutates viper and the environment
func TestGetJwtFileUnique(t *testing.T) {
	t.Cleanup(func() { flags.SetRegion(region.Default) })

	seen := make(map[string]string)

	for _, name := range region.Known() {
		for _, stage := range []string{stageProd, stageStaging, stageDev, "local"} {
			reg, err := region.Parse(name)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", name, err)
			}

			t.Setenv("AMP_STAGE_OVERRIDE", stage)
			flags.SetRegion(reg)

			path := GetJwtFile()
			key := name + "/" + stage

			if other, ok := seen[path]; ok {
				t.Errorf("GetJwtFile() = %q for both %s and %s", path, other, key)
			}

			seen[path] = key
		}
	}
}
