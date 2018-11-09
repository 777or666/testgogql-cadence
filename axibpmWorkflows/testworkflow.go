package axibpmWorkflows

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/777or666/testgogql-cadence/axibpmActivities"
	"github.com/777or666/testgogql-cadence/helpers"
)

type TestWorkflowConfiguration struct {
	WorkflowName            string `yaml:"workflowname"`
	ScheduleToStartTimeout1 int    `yaml:"scheduletostarttimeout1"`
	StartToCloseTimeout1    int    `yaml:"starttoclosetimeout1"`
	HeartbeatTimeout1       int    `yaml:"heartbeattimeout1"`
	ScheduleToStartTimeout2 int    `yaml:"scheduletostarttimeout2"`
	StartToCloseTimeout2    int    `yaml:"starttoclosetimeout2"`
	HeartbeatTimeout2       int    `yaml:"heartbeattimeout2"`
}

const (
	configfileTestworkflow = "axibpmWorkflows/testworkflow.yaml"
)

// Наименование воркфлоу
//const WorkflowName = "Согласование ТКП"

// Выполняем воркфлоу
func TestWorkflow(ctx workflow.Context, id string, emailconfig *helpers.EmailConfig) (result string, err error) {

	//Чтение файла конфигурации
	configData, err := ioutil.ReadFile(configfileTestworkflow)
	if err != nil {
		panic(fmt.Sprintf("Ошибка чтения файла: %v, Error: %v", configfileTestworkflow, err))
	}

	var Config TestWorkflowConfiguration

	if err := yaml.Unmarshal(configData, &Config); err != nil {
		panic(fmt.Sprintf("Ошибка инициализации конфигурации: %v", err))
	}
	//*******************************

	// ***Activity 1 - EmailSenderActivity***
	ao1 := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Duration(Config.ScheduleToStartTimeout1) * time.Minute,
		StartToCloseTimeout:    time.Duration(Config.StartToCloseTimeout1) * time.Minute,
		HeartbeatTimeout:       time.Duration(Config.HeartbeatTimeout1) * time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao1)

	logger := workflow.GetLogger(ctx)

	//workflow.Go(ctx, func(ctx workflow.Context) {

	testResult := ""
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + "Тема 1" + "\n"
	emailbody := subject + mime + "\n" + "Старт тестового процесса\n"

	addressees := []string{"kravetsmihail@mail.ru", "kravetsmihail@yandex.ru"}

	logger.Info("Start EmailSenderActivity")

	err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

	logger.Info("EmailSenderActivity result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
		return "", err
	}
	//********************************************

	//***Activity 2 - TestActivity***
	logger.Info("СТАРТ TestWorkflow! id=" + id)

	ao2 := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Duration(Config.ScheduleToStartTimeout2) * time.Minute,
		StartToCloseTimeout:    time.Duration(Config.StartToCloseTimeout2) * time.Minute,
		HeartbeatTimeout:       time.Duration(Config.HeartbeatTimeout2) * time.Minute,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao2)

	err = workflow.ExecuteActivity(ctx2, axibpmActivities.TestActivity, id).Get(ctx, &testResult)

	logger.Info("TestWorkflow result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Тестовая операция не выполнена", zap.Error(err))

		subject = "Subject: " + "ПРОЦЕСС НЕ ВЫПОЛНЕН" + "!\n"
		emailbody = subject + mime + "\n" + "Аварийное завершение процесса\n"
		err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

		if err != nil {
			logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
			return "", err
		}

		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))

	subject = "Subject: " + "ПРОЦЕСС ВЫПОЛНЕН" + "!\n"
	emailbody = subject + mime + "\n" + "Процесс успешнро завершен\n"
	err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

	if err != nil {
		logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
		return "", err
	}
	//****************************************
	//})

	return "COMPLETED", nil
}
