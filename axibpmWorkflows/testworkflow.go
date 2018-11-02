package axibpmWorkflows

import (
	//"context"

	//"log"
	"time"

	"github.com/777or666/testgogql-cadence/axibpmActivities"

	//	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// Наименование воркфлоу
const WorkflowName = "Согласование ТКП"

// Выполняем наш воркфлоу
func TestWorkflow(ctx workflow.Context, id string) (result string, err error) {

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	//workflow.Go(ctx, func(ctx workflow.Context) {
	var testResult string

	err = workflow.ExecuteActivity(ctx, axibpmActivities.TestActivity, id).Get(ctx, &testResult)

	logger.Info("TestWorkflow result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Активность не выполнена", zap.Error(err))
		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//})

	return "COMPLETED", nil
}
