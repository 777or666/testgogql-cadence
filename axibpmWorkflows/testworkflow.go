package axibpmWorkflows

import (
	//"encoding/json"
	//"io/ioutil"
	s "strings"
	"time"

	//"gopkg.in/yaml.v2"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/777or666/testgogql-cadence/axibpmActivities"
	"github.com/777or666/testgogql-cadence/helpers"
)

const (
	configfileTestworkflow   = "axibpmWorkflows/testworkflow.yaml"
	templatefileTestworkflow = "templates/emailmaintemplate.html"
)

// Выполняем воркфлоу
//func TestWorkflow(ctx workflow.Context, id string, emailconfig *helpers.EmailConfig, EmailResponsible []*string, EmailParticipants []*string, input string) (result string, err error) {
func TestWorkflow(ctx workflow.Context, wrfinput helpers.WorkflowInput) (result string, err error) {
	logger := workflow.GetLogger(ctx)

	config := wrfinput.WorkflowConfig
	emails := wrfinput.WorkflowEmails
	emailconfig := &wrfinput.WorkflowEmailConfig
	id := wrfinput.WorkflowSettings.WorkflowId

	// ***Activity 1 - EmailSenderActivity***
	ao1 := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Duration(config.WorkflowActivity[1].ScheduleToStartTimeout) * time.Minute,
		StartToCloseTimeout:    time.Duration(config.WorkflowActivity[1].StartToCloseTimeout) * time.Minute,
		HeartbeatTimeout:       time.Duration(config.WorkflowActivity[1].HeartbeatTimeout) * time.Minute,
	}

	ctx1 := workflow.WithActivityOptions(ctx, ao1)

	testResult := ""

	subject := config.WorkflowName + ". СТАРТ (" + id + ")"
	emailbody := ""

	var addressees []string
	//ВНИМАНИЕ! Пока все адреса добавляются в кучу!! ПЕРЕДЕЛАТЬ
	//добавляем адреса ответственных за процесс
	for _, value := range emails.EmailResponsible {
		addressees = append(addressees, value)
	}
	//добавляем адреса участников процесса
	for _, value := range emails.EmailParticipants {
		addressees = append(addressees, value)
	}

	emailrequest := helpers.EmailRequest{
		To:      addressees,
		Subject: subject,
		Body:    emailbody,
		Config:  emailconfig,
	}

	//	wrfinput := helpers.WorkflowInput{}
	//	json.Unmarshal([]byte(input), &wrfinput)

	emaildata := helpers.EmailRequestData{
		Message:      config.WorkflowName + " => Старт процесса",
		WorkflowData: wrfinput,
	}

	//Формируем body письма из шаблона и данных
	emailrequest.ParseEmailTemplate(templatefileTestworkflow, emaildata)

	testWorkflow_sendEmail(ctx1, logger, emailrequest)
	//********************************************

	//***Activity 2 - TestActivity***
	logger.Info("---СТАРТ TestActivity 1---")

	var processingDone bool

	ao2 := workflow.ActivityOptions{
		ActivityID:             config.WorkflowActivity[2].ActivityId,
		ScheduleToStartTimeout: time.Duration(config.WorkflowActivity[2].ScheduleToStartTimeout) * time.Minute,
		StartToCloseTimeout:    time.Duration(config.WorkflowActivity[2].StartToCloseTimeout) * time.Minute,
		HeartbeatTimeout:       time.Duration(config.WorkflowActivity[2].HeartbeatTimeout) * time.Minute,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao2)
	//запускаем таймеры оповещений в данном случае только для второй Activity
	childCtx, cancelHandler := workflow.WithCancel(ctx2)
	selector := workflow.NewSelector(ctx2)

	f := workflow.ExecuteActivity(ctx2, axibpmActivities.TestActivity, id)

	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// отключаем timerFuture
		cancelHandler()
	})

	bodycounter := 1

	for _, v := range config.WorkflowActivity[2].ActivityReminders {
		timerFuture := workflow.NewTimer(childCtx, time.Minute*time.Duration(v.ReminderTime))

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			if !processingDone {
				// обработка еще не завершена, когда срабатывает таймер, отправляем уведомление по электронной почте
				subject = config.WorkflowName + ". ОПОВЕЩЕНИЕ (" + id + ")"
				emailbody = ""

				emailrequest = helpers.EmailRequest{
					To:      addressees,
					Subject: subject,
					Body:    emailbody,
					Config:  emailconfig,
				}

				emaildata = helpers.EmailRequestData{
					Message:      config.WorkflowActivity[2].ActivityReminders[bodycounter].ReminderText,
					WorkflowData: wrfinput,
				}

				emailrequest.ParseEmailTemplate(templatefileTestworkflow, emaildata)

				testWorkflow_sendEmail(ctx1, logger, emailrequest)

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
		logger.Error("(TKP1) ОШИБКА! Тестовая операция не выполнена", zap.Error(err))

		subject = config.WorkflowName + ". НЕ ВЫПОЛНЕНО (" + id + ")"
		emailbody = ""
		emailrequest = helpers.EmailRequest{
			To:      addressees,
			Subject: subject,
			Body:    emailbody,
			Config:  emailconfig,
		}

		var prichina string = ""

		if err.Error() == "CanceledError" {
			prichina = "Процесс принудительно остановлен."
		} else if s.Contains(err.Error(), "TimeoutType") {
			prichina = "Просрочен срок."
		} else {
			prichina = err.Error()
		}

		emaildata = helpers.EmailRequestData{
			Message:      "Аварийное завершение процесса. Взять в работу. Причина: " + prichina,
			WorkflowData: wrfinput,
		}

		emailrequest.ParseEmailTemplate(templatefileTestworkflow, emaildata)
		//отсылаем письмо напрямую (не через activity) так как из-за ошибки контекста воркфлоу уже нет
		resemail, emailerr := emailrequest.SendEmail()
		if !resemail {
			logger.Info("(TKP1) Не удалось отправить е-маил! Ошибка: " + emailerr.Error())
		}

		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//****************************************

	//***Activity 3 - TestActivity***
	logger.Info("---СТАРТ TestActivity 2---")

	processingDone = false

	ao3 := workflow.ActivityOptions{
		ActivityID:             config.WorkflowActivity[3].ActivityId,
		ScheduleToStartTimeout: time.Duration(config.WorkflowActivity[3].ScheduleToStartTimeout) * time.Minute,
		StartToCloseTimeout:    time.Duration(config.WorkflowActivity[3].StartToCloseTimeout) * time.Minute,
		HeartbeatTimeout:       time.Duration(config.WorkflowActivity[3].HeartbeatTimeout) * time.Minute,
	}
	ctx3 := workflow.WithActivityOptions(ctx, ao3)
	//запускаем таймеры оповещений в данном случае только для второй Activity
	childCtx, cancelHandler = workflow.WithCancel(ctx3)
	selector = workflow.NewSelector(ctx3)

	f = workflow.ExecuteActivity(ctx3, axibpmActivities.TestActivity, id)

	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// отключаем timerFuture
		cancelHandler()
	})

	bodycounter = 1

	for _, v := range config.WorkflowActivity[3].ActivityReminders {
		timerFuture := workflow.NewTimer(childCtx, time.Minute*time.Duration(v.ReminderTime))

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			if !processingDone {
				// обработка еще не завершена, когда срабатывает таймер, отправляем уведомление по электронной почте

				subject = config.WorkflowName + ". ОПОВЕЩЕНИЕ (" + id + ")"
				emailbody = config.WorkflowActivity[3].ActivityReminders[bodycounter].ReminderText + "\n" +
					"id: " + id + "\n"

				emailrequest = helpers.EmailRequest{
					To:      addressees,
					Subject: subject,
					Body:    emailbody,
					Config:  emailconfig,
				}
				emaildata = helpers.EmailRequestData{
					Message:      config.WorkflowActivity[3].ActivityReminders[bodycounter].ReminderText,
					WorkflowData: wrfinput,
				}

				emailrequest.ParseEmailTemplate(templatefileTestworkflow, emaildata)

				testWorkflow_sendEmail(ctx1, logger, emailrequest)
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
		logger.Error("(TKP2) ОШИБКА! Тестовая операция не выполнена", zap.Error(err))

		subject = config.WorkflowName + ". НЕ ВЫПОЛНЕНО (" + id + ")"
		emailbody = ""

		emailrequest = helpers.EmailRequest{
			To:      addressees,
			Subject: subject,
			Body:    emailbody,
			Config:  emailconfig,
		}

		var prichina string = ""

		if err.Error() == "CanceledError" {
			prichina = "Процесс принудительно остановлен."
		} else if s.Contains(err.Error(), "TimeoutType") {
			prichina = "Просрочен срок."
		} else {
			prichina = err.Error()
		}

		emaildata = helpers.EmailRequestData{
			Message:      "Аварийное завершение процесса. Согласовать. Причина: " + prichina,
			WorkflowData: wrfinput,
		}

		emailrequest.ParseEmailTemplate(templatefileTestworkflow, emaildata)
		//отсылаем письмо напрямую (не через activity) так как из-за ошибки контекста воркфлоу уже нет
		resemail, emailerr := emailrequest.SendEmail()
		if !resemail {
			logger.Info("(TKP2) Не удалось отправить е-маил! Ошибка: " + emailerr.Error())
		}

		return "", err
	}
	logger.Info("ВЫПОЛНЕНО: "+id, zap.String("Result", testResult))
	//****************************************

	// ***Activity 4 - EmailSenderActivity***
	subject = config.WorkflowName + ". ВЫПОЛНЕНО  (" + id + ")"
	emailbody = ""
	emailrequest = helpers.EmailRequest{
		To:      addressees,
		Subject: subject,
		Body:    emailbody,
		Config:  emailconfig,
	}

	emaildata = helpers.EmailRequestData{
		Message:      "<= Процесс успешно завершен",
		WorkflowData: wrfinput,
	}

	emailrequest.ParseEmailTemplate(templatefileTestworkflow, emaildata)

	testWorkflow_sendEmail(ctx1, logger, emailrequest)
	//****************************************

	return "COMPLETED", nil
}

//Рассылка сообщений на е-маил
func testWorkflow_sendEmail(ctx workflow.Context, logger *zap.Logger, emailrequest helpers.EmailRequest) {

	emailerr := workflow.ExecuteActivity(ctx, axibpmActivities.EmailSenderActivity, &emailrequest).Get(ctx, nil)

	if emailerr != nil {
		logger.Error("ОШИБКА! Выполнить е-маил рассылку не удалось.", zap.Error(emailerr))
	}
}
