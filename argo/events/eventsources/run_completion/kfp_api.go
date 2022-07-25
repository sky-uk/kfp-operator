package run_completion

import "context"

type KfpApi interface {
	 GetRunConfiguration(ctx context.Context, runId string) (string, error)
}
