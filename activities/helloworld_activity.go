package activities

import (
	"context"

	"go.uber.org/cadence/activity"
)

func HelloworldActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("axibpm activity started")
	return "AXI-BPM. " + name + "!", nil
}
