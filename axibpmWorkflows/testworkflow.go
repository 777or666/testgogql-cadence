package axibpmWorkflows

import (
	//"context"

	//"log"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/777or666/testgogql-cadence/axibpmActivities"
)

// Наименование воркфлоу
const WorkflowName = "Согласование ТКП"

// Выполняем наш воркфлоу
func TestWorkflow(ctx workflow.Context, id string) (result string, err error) {

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       5 * time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	//workflow.Go(ctx, func(ctx workflow.Context) {
	//var testResult string

	testResult := ""

	addressees := []string{"kravetsmihail@mail.ru"}

	logger.Info("Start EmailSenderActivity")

	err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees).Get(ctx, &testResult)

	logger.Info("EmailSenderActivity result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Активность не выполнена", zap.Error(err))
		return "", err
	}

	logger.Info("СТАРТ TestWorkflow! id=" + id)

	err = workflow.ExecuteActivity(ctx1, axibpmActivities.TestActivity, id).Get(ctx, &testResult)
	//err = workflow.ExecuteActivity(ctx1, axibpmActivities.TestActivity, id).Get(ctx, nil)

	logger.Info("TestWorkflow result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Активность не выполнена", zap.Error(err))
		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//})

	return "COMPLETED", nil
}
