package axibpmActivities

import (
	"context"

	"go.uber.org/cadence/activity"
)

func TestActivity(ctx context.Context, id string) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("axibpmActivities: TestActivity")

	activityInfo := activity.GetInfo(ctx)

	logger.Info("axibpmActivities token: " + string(activityInfo.TaskToken))

	result := "Операция выполена!"

	// ErrActivityResultPending возвращается из активности, чтобы указать, что действие не завершено.
	// действие будет выполняться асинхронно при вызове Client.CompleteActivity ().
	return result, activity.ErrResultPending
}
