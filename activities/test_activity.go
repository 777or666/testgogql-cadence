package axibpmActivities

import (
	"context"
	"log"

	"go.uber.org/cadence/activity"
)

func TestActivity(ctx context.Context, id string, token *string) (string, error) {
	logger := activity.GetLogger(ctx)

	log.Println("!!TestActivity!!")

	activityInfo := activity.GetInfo(ctx)
	*token = string(activityInfo.TaskToken)

	log.Println(*token)

	logger.Info("AXIBPM_ACTIVITIES: НАЧАЛО")
	return "AXI-BPM. " + id + "!", nil
}
