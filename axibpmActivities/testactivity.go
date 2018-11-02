package axibpmActivities

import (
	"context"
	//"errors"

	"go.uber.org/cadence/activity"
	//"go.uber.org/zap"
)

func TestActivity(ctx context.Context, id string) (string, error) {
	logger := activity.GetLogger(ctx)

	activityInfo := activity.GetInfo(ctx)

	logger.Info("axibpmActivities: TestActivity")
	logger.Info("axibpmActivities token: " + string(activityInfo.TaskToken))

	// ErrActivityResultPending возвращается из активности, чтобы указать, что действие не завершено.
	// действие будет выполняться асинхронно при вызове Client.CompleteActivity ().
	return "", activity.ErrResultPending
}
