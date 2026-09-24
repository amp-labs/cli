package cmd

import (
	"reflect"
	"testing"
	"time"

	"github.com/amp-labs/cli/request"
)

func TestSummarizeOperationsOmitsProviderResource(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(time.Minute)
	operations := []*request.Operation{
		{
			Id:         "operation-id",
			ActionType: "proxy",
			Status:     "success",
			Resource:   "/1.0/projects/provider-record-id",
			CreateTime: startedAt,
			UpdateTime: &completedAt,
		},
	}

	got := summarizeOperations(operations)
	wantCompletedAt := "2026-08-19T12:01:00Z"
	want := []operationSummary{
		{
			Id:          "operation-id",
			Action:      "proxy",
			Status:      "success",
			StartedAt:   "2026-08-19T12:00:00Z",
			CompletedAt: &wantCompletedAt,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summarizeOperations() = %#v, want %#v", got, want)
	}

	if reflect.ValueOf(got[0]).FieldByName("Resource").IsValid() {
		t.Fatal("operation summary includes provider resource")
	}
}
