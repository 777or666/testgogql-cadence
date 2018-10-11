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
	token := string(activityInfo.TaskToken)

	logger.Info("axibpmActivities: TestActivity")
	logger.Info(string(activityInfo.TaskToken))

	// ErrActivityResultPending возвращается из активности, чтобы указать, что действие не завершено.
	// действие будет выполняться асинхронно при вызове Client.CompleteActivity ().
	return token, activity.ErrResultPending
}
