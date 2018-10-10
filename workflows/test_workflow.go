package axibpmWorkflow

import (
	//"context"
	"log"
	"time"

	"github.com/777or666/testgogql-cadence/activities"

	//	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// Наименование воркфлоу
const WorkflowName = "Согласование ТКП"

// Регистрируем активности
//func init() {
//workflow.Register(TestWorkflow)
//	activity.Register(axibpm_activities.TestActivity)
//}

// Выполняем наш воркфлоу
func TestWorkflow(ctx workflow.Context, id string, token *string) error {
	log.Println("START WORKFLOW!")
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 1 * time.Minute,
		StartToCloseTimeout:    1 * time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	//workflow.Go(ctx, func(ctx workflow.Context) {
	logger.Info("СТАРТ: " + id)

	var testResult string
	err := workflow.ExecuteActivity(ctx, axibpmActivities.TestActivity, id, token).Get(ctx, &testResult)
	if err != nil {
		logger.Error("ОШИБКА! Активность не выполнена", zap.Error(err))
		//return err
	}

	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//})

	return nil
}
