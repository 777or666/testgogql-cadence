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

const (
	configfileTestworkflow = "axibpmWorkflows/testworkflow.yaml"
)

// Выполняем воркфлоу
func TestWorkflow(ctx workflow.Context, id string, emailconfig *helpers.EmailConfig, EmailResponsible []*string, EmailParticipants []*string) (result string, err error) {

	//Чтение файла конфигурации
	configData, err := ioutil.ReadFile(configfileTestworkflow)
	if err != nil {
		panic(fmt.Sprintf("Ошибка чтения файла: %v, Error: %v", configfileTestworkflow, err))
	}

	Config := helpers.WorkflowConfiguration{}

	if err := yaml.Unmarshal(configData, &Config); err != nil {
		panic(fmt.Sprintf("Ошибка инициализации конфигурации: %v", err))
	}
	//*******************************

	// ***Activity 1 - EmailSenderActivity***
	ao1 := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Duration(Config.WorkflowActivity[1].ScheduleToStartTimeout) * time.Minute,
		StartToCloseTimeout:    time.Duration(Config.WorkflowActivity[1].StartToCloseTimeout) * time.Minute,
		HeartbeatTimeout:       time.Duration(Config.WorkflowActivity[1].HeartbeatTimeout) * time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao1)
	logger := workflow.GetLogger(ctx)
	//запускаем таймеры оповещений в данном случае только для второй Activity
	//childCtx, cancelHandler := workflow.WithCancel(ctx)
	//selector := workflow.NewSelector(ctx)
	//var processingDone bool

	//workflow.Go(ctx, func(ctx workflow.Context) {

	testResult := ""
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + Config.WorkflowName + ". Старт процесса - " + id + "\n"
	emailbody := subject + mime + "\n" +
		"Старт тестового процесса\n" + "id: " +
		id + "\n"

	var addressees []string
	//ВНИМАНИЕ! Пока все адреса добавляются в кучу!! ПЕРЕДЕЛАТЬ
	//добавляем адреса ответственных за процесс
	for _, value := range EmailResponsible {
		addressees = append(addressees, *value)
	}
	//добавляем адреса участников процесса
	for _, value := range EmailParticipants {
		addressees = append(addressees, *value)
	}

	logger.Info("Start EmailSenderActivity")

	err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

	logger.Info("EmailSenderActivity result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
		return "", err
	}
	//********************************************

	//***Activity 2 - TestActivity***
	logger.Info("---СТАРТ TestActivity 1!---")

	var processingDone bool

	ao2 := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Duration(Config.WorkflowActivity[2].ScheduleToStartTimeout) * time.Minute,
		StartToCloseTimeout:    time.Duration(Config.WorkflowActivity[2].StartToCloseTimeout) * time.Minute,
		HeartbeatTimeout:       time.Duration(Config.WorkflowActivity[2].HeartbeatTimeout) * time.Minute,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao2)
	//запускаем таймеры оповещений в данном случае только для второй Activity
	childCtx, cancelHandler := workflow.WithCancel(ctx2)
	selector := workflow.NewSelector(ctx2)

	//err = workflow.ExecuteActivity(ctx2, axibpmActivities.TestActivity, id).Get(ctx, &testResult)

	f := workflow.ExecuteActivity(ctx2, axibpmActivities.TestActivity, id)
	//err = f.Get(ctx, &testResult)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// отключаем timerFuture
		cancelHandler()
	})

	bodycounter := 1

	for _, v := range Config.WorkflowActivity[2].ActivityReminders {
		timerFuture := workflow.NewTimer(childCtx, time.Minute*time.Duration(v.ReminderTime))

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			if !processingDone {
				// обработка еще не завершена, когда срабатывает таймер, отправляем уведомление по электронной почте

				subject = "Subject: " + Config.WorkflowName + ". ОПОВЕЩЕНИЕ" + "! (" + id + ")\n"
				emailbody = subject + mime + "\n" +
					Config.WorkflowActivity[2].ActivityReminders[bodycounter].ReminderText + "\n" +
					"id: " + id + "\n"
				err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)
				bodycounter++
			}
		})
	}

	// ждем окончания таймера или выполнения операции
	selector.Select(ctx2)

	// теперь завершается выполнение операции или запускается таймер
	if !processingDone {
		// операция еще не выполнена, поэтому обработчик для таймера отправит уведомление по электронной почте.
		// мы все еще хотим, чтобы операция закончилась, поэтому ждем
		selector.Select(ctx2)
	}

	err = f.Get(ctx, &testResult)

	logger.Info("TestWorkflow result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Тестовая операция не выполнена", zap.Error(err))

		subject = "Subject: " + Config.WorkflowName + ". ПРОЦЕСС НЕ ВЫПОЛНЕН" + "!\n"
		emailbody = subject + mime + "\n" + "Аварийное завершение процесса - " + id + "\n"

		err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

		if err != nil {
			logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
			return "", err
		}

		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//****************************************

	//***Activity 3 - TestActivity***
	logger.Info("---СТАРТ TestActivity 2!---")

	processingDone = false

	ao3 := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Duration(Config.WorkflowActivity[3].ScheduleToStartTimeout) * time.Minute,
		StartToCloseTimeout:    time.Duration(Config.WorkflowActivity[3].StartToCloseTimeout) * time.Minute,
		HeartbeatTimeout:       time.Duration(Config.WorkflowActivity[3].HeartbeatTimeout) * time.Minute,
	}
	ctx3 := workflow.WithActivityOptions(ctx, ao3)
	//запускаем таймеры оповещений в данном случае только для второй Activity
	childCtx, cancelHandler = workflow.WithCancel(ctx3)
	selector = workflow.NewSelector(ctx3)

	//err = workflow.ExecuteActivity(ctx2, axibpmActivities.TestActivity, id).Get(ctx, &testResult)

	f = workflow.ExecuteActivity(ctx3, axibpmActivities.TestActivity, id)
	//err = f.Get(ctx, &testResult)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// отключаем timerFuture
		cancelHandler()
	})

	bodycounter = 1

	for _, v := range Config.WorkflowActivity[3].ActivityReminders {
		timerFuture := workflow.NewTimer(childCtx, time.Minute*time.Duration(v.ReminderTime))

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			if !processingDone {
				// обработка еще не завершена, когда срабатывает таймер, отправляем уведомление по электронной почте

				subject = "Subject: " + Config.WorkflowName + ". ОПОВЕЩЕНИЕ" + "! (" + id + ")\n"
				emailbody = subject + mime + "\n" +
					Config.WorkflowActivity[3].ActivityReminders[bodycounter].ReminderText + "\n" +
					"id: " + id + "\n"
				err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)
				bodycounter++
			}
		})
	}

	// ждем окончания таймера или выполнения операции
	selector.Select(ctx3)

	// теперь завершается выполнение операции или запускается таймер
	if !processingDone {
		// операция еще не выполнена, поэтому обработчик для таймера отправит уведомление по электронной почте.
		// мы все еще хотим, чтобы операция закончилась, поэтому ждем
		selector.Select(ctx3)
	}

	err = f.Get(ctx, &testResult)

	logger.Info("TestWorkflow result: " + testResult)

	if err != nil {
		logger.Error("ОШИБКА! Тестовая операция не выполнена", zap.Error(err))

		subject = "Subject: " + Config.WorkflowName + ". ПРОЦЕСС НЕ ВЫПОЛНЕН" + "!\n"
		emailbody = subject + mime + "\n" + "Аварийное завершение процесса - " + id + "\n"

		err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

		if err != nil {
			logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
			return "", err
		}

		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//****************************************

	// ***Activity 4 - EmailSenderActivity***
	subject = "Subject: " + Config.WorkflowName + ". ПРОЦЕСС ВЫПОЛНЕН - " + id + "!\n"
	emailbody = subject + mime + "\n" + "Процесс успешно завершен\n" +
		"id: " + id + "\n"
	err = workflow.ExecuteActivity(ctx1, axibpmActivities.EmailSenderActivity, addressees, emailbody, emailconfig).Get(ctx, &testResult)

	if err != nil {
		logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(err))
		return "", err
	}
	//****************************************
	//})

	return "COMPLETED", nil
}
