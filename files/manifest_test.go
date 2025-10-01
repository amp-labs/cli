package files

import (
	"testing"

	"github.com/amp-labs/cli/openapi"
)

func TestGetRemovedReadObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		oldRevision *openapi.Integration
		newInteg    *openapi.Integration
		want        []string
	}{
		{
			name: "no read objects removed",
			oldRevision: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Accounts"},
						{ObjectName: "contacts"},
					},
				},
			},
			newInteg: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "accounts"},
						{ObjectName: "contacts"},
					},
				},
			},
			want: nil,
		},
		{
			name: "one object removed (case-insensitive)",
			oldRevision: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "AccounTs"},
						{ObjectName: "contacts"},
					},
				},
			},
			newInteg: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "accounts"},
					},
				},
			},
			want: []string{"contacts"},
		},
		{
			name: "all objects removed",
			oldRevision: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "accounts"},
						{ObjectName: "contacts"},
					},
				},
			},
			newInteg: &openapi.Integration{
				Read: nil,
			},
			want: []string{"accounts", "contacts"},
		},
		{
			name:        "no old read config",
			oldRevision: &openapi.Integration{},
			newInteg: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "accounts"},
					},
				},
			},
			want: nil,
		},
		{
			name:        "nil old revision",
			oldRevision: nil,
			newInteg: &openapi.Integration{
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "accounts"},
					},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetRemovedReadObjects(tt.oldRevision, tt.newInteg)

			if len(got) != len(tt.want) {
				t.Errorf("GetRemovedReadObjects() = %v, want %v", got, tt.want)

				return
			}

			if tt.want == nil && got == nil {
				return
			}

			// Check that all expected objects are in the result
			gotMap := make(map[string]bool)
			for _, obj := range got {
				gotMap[obj] = true
			}

			for _, wantObj := range tt.want {
				if !gotMap[wantObj] {
					t.Errorf("GetRemovedReadObjects() missing %s, got %v, want %v", wantObj, got, tt.want)
				}
			}
		})
	}
}
