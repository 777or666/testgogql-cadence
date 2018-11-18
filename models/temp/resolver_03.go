//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"
	"log"
	"sync"
	"time"

	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/777or666/testgogql-cadence/helpers"

	"encoding/json"

	"github.com/go-resty/resty"
	//"github.com/pborman/uuid"
	"go.uber.org/cadence"
	"go.uber.org/cadence/client"
)

var UrlRestService string
var h *helpers.SampleHelper
var workflowClient client.Client
var ApplicationName string //ВСЕГДА должно совпадать в воркере и всех его воркфлоу
var PrefixWorkflowFunc string
var EmailConfiguration *helpers.EmailConfig

const (
	pathtoworkflows = "axibpmWorkflows/"
)

type resolver struct {
	mu sync.Mutex // nolint: structcheck
}

func (r *resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

func New(urlRestService string, applicationName string, prefixworkflowfunc string, emailconfig *helpers.EmailConfig, hw *helpers.SampleHelper) Config {
	UrlRestService = urlRestService
	ApplicationName = applicationName
	PrefixWorkflowFunc = prefixworkflowfunc
	EmailConfiguration = emailconfig
	h = hw

	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()

	if err != nil {
		log.Println("ОШИБКА при BuildCadenceClient: " + err.Error())
		panic(err)
	}

	return Config{Resolvers: &resolver{}}
}

type mutationResolver struct{ *resolver }

//Запуск бизнес-процесса
//id - идентификатор процесса
//name - программное наименование функции воркфлоу (пример, "TestWorkflow")
//input - JSON информация вида:
//{
//  "user": {
//    "useremail": "777@mail.ru",
//    "username": "Иванов И.И.",
//    "department": "Коммерческий отдел"
//  },
//  "workflowdata": {
//    "objectId": "8e0928cc-43f4-4c60-9d1a-c3f13a2787ef",
//    "objectHref": "https://mediametrics.ru/rating/ru/online.html",
//    "objectName": "КП для Газпрома на 1000 УУГ",
//    "objectType": "Коммерческое предложение (КП)",
//    "activity": "Новое ТКП",
//    "comment": "СРОЧНО! Необходимо!"
//  },
//  "workflowsettings": {
//    "WorkflowId": "TKP-181116-62301",
//    "ExecutionStartToCloseTimeout": "10",
//    "DecisionTaskStartToCloseTimeout": "10"
//  },
//  "workflowemails": {
//    "emailResponsible": [
//      "kravetsmihail@mail.ru"
//    ],
//    "emailParticipants": [
//      "kravetsmihail@mail.ru",
//      "forum@axitech.ru",
//      "kravetsmihail@yandex.ru"
//    ]
//  }
//}
//ExecutionStartToCloseTimeout - тайм-айт выполнения рабочего процесса
//DecisionTaskStartToCloseTimeout - тайм-аут для обработки задачи решения с момента, когда воркер вытащил эту задачу
//EmailResponsible - e-mail ответственных
//EmailParticipants - e-mail остальных участников процесса
func (r *mutationResolver) WorkflowStart(ctx context.Context, id string, name string, input *string, ExecutionStartToCloseTimeout *int, DecisionTaskStartToCloseTimeout *int, EmailResponsible []*string, EmailParticipants []*string) (Workflow, error) {
	//r.mu.Lock()
	//r.mu.Unlock()

	cteatedAt := time.Now()

	//Чтение файла конфигурации workflow
	wfData, err := ioutil.ReadFile(pathtoworkflows + "/" + name + ".yaml")
	if err != nil {
		log.Println("Ошибка чтения файла конфигурации процесса", err.Error())
		return Workflow{}, err
	}

	Config := helpers.WorkflowConfiguration{}

	if err := yaml.Unmarshal(wfData, &Config); err != nil {
		log.Println("Ошибка инициализации файла конфигурации процесса", err.Error())
		return Workflow{}, err
	}
	//****************************************

	//Маршалинг данных из input
	wrfinput := helpers.WorkflowInput{}

	if err = json.Unmarshal([]byte(*input), &wrfinput); err != nil {
		log.Println("Ошибка маршалинга JSON данных из input", err.Error())
		return Workflow{}, err
	}

	wrfinput.WorkflowConfig = Config
	wrfinput.WorkflowEmailConfig = *EmailConfiguration

	workflowOptions := client.StartWorkflowOptions{
		ID:       wrfinput.WorkflowSettings.WorkflowId,
		TaskList: ApplicationName,
		//Тайм-аут выполнения рабочего процесса
		ExecutionStartToCloseTimeout: time.Duration(wrfinput.WorkflowSettings.ExecutionStartToCloseTimeout) * time.Minute,
		//Тайм-аут для обработки задачи с момента, когда рабочий
		//вытащил эту задачу. Если задача решения потеряна, она повторится после этого таймаута.
		DecisionTaskStartToCloseTimeout: time.Duration(wrfinput.WorkflowSettings.DecisionTaskStartToCloseTimeout) * time.Minute,
		WorkflowIDReusePolicy:           2, // см. описание в cadence
	}

	fullname := PrefixWorkflowFunc + "." + name

	//wfId, wfRunId := h.StartWorkflow(workflowOptions, fullname, id, EmailConfiguration, EmailResponsible, EmailParticipants, *input)
	wfId, wfRunId := h.StartWorkflow(workflowOptions, fullname, wrfinput)

	var m []Activity
	//генерим массив операций []Activity
	for _, v := range Config.WorkflowActivity {
		if v.ActivityId != "" {
			temp := Activity{
				ActivityID:          v.ActivityId,
				Description:         v.Description,
				Operation:           v.Operation,
				Roles:               v.Roles,
				Starttoclosetimeout: v.StartToCloseTimeout,
			}
			m = append(m, temp)
		}
	}

	//TODO: продумать состав, надо пробросить тамауты, инпуты и т.д.
	wrf := Workflow{
		ID:         wrfinput.WorkflowSettings.WorkflowId,
		Name:       name,
		WorkflowID: wfId,
		RunID:      wfRunId,
		CreatedAt:  cteatedAt,
		StartTime:  time.Now(),
		Activities: m,
		TaskList:   ApplicationName,
	}

	return wrf, nil //nil заменить на err
}
func (r *mutationResolver) WorkflowCancel(ctx context.Context, id string, runID *string) (*string, error) {
	result := "Бизнес-процесс отменен"

	err := workflowClient.CancelWorkflow(context.Background(), id, *runID)
	if err != nil {
		log.Println("ОШИБКА! Не удалось отменить бизнес-процесс. " + err.Error())

		return nil, err
	}

	return &result, nil
}
func (r *mutationResolver) WorkflowTerminate(ctx context.Context, id string, runID *string, reason *string, input *string) (*string, error) {
	result := "Бизнес-процесс прерван"
	details := *input

	err := workflowClient.TerminateWorkflow(context.Background(), id, *runID, *reason, []byte(details))
	if err != nil {
		log.Println("ОШИБКА! Не удалось прервать бизнес-процесс. " + err.Error())

		return nil, err
	}

	return &result, nil
}
func (r *mutationResolver) ActivityPerform(ctx context.Context, domain *string, workflowID string, runID *string, activityID string, input *string) (*string, error) {
	result := "Операция выполнена"

	log.Println("ActivityPerform! Старт")

	err := workflowClient.CompleteActivityByID(context.Background(), *domain, workflowID, *runID, activityID, *input, nil)
	if err != nil {

		log.Println("ОШИБКА! Не удалось выполнить операцию. " + err.Error())

		return nil, err
	}

	return &result, nil
}
func (r *mutationResolver) ActivityFailed(ctx context.Context, domain *string, workflowID string, runID *string, activityID string, input *string) (*string, error) {
	result := "Операция отменена"

	//добавить детали - причину отмены активности и кто ее отменил
	err := workflowClient.CompleteActivityByID(context.Background(), *domain, workflowID, *runID, activityID, nil, cadence.NewCustomError(*input, result))
	if err != nil {

		log.Println("ОШИБКА! Не удалось отменить операцию. " + err.Error())

		return nil, err
	}

	return &result, nil
}

type queryResolver struct{ *resolver }

func (r *queryResolver) Domain(ctx context.Context, name *string) (*Domain, error) {
	//log.Println(UrlRestService)

	resp, err := resty.R().Get(UrlRestService + "/api/domain/" + *name)

	if err != nil {
		panic(err.Error())
	}

	var domain Domain

	if err := json.Unmarshal(resp.Body(), &domain); err != nil {
		panic(err.Error())
	}

	return &domain, nil
}
func (r *queryResolver) Workflow(ctx context.Context, id string) (*Workflow, error) {
	panic("not implemented")
}
func (r *queryResolver) AllWorkflows(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *string, domain *string) ([]*Workflow, error) {
	panic("not implemented")
}
func (r *queryResolver) AllOpenWorkflows(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *string, domain *string) ([]*Workflow, error) {
	panic("not implemented")
}
func (r *queryResolver) AllCloseWorkflows(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *string, domain *string) ([]*Workflow, error) {
	panic("not implemented")
}

type subscriptionResolver struct{ *resolver }

func (r *subscriptionResolver) Workflow(ctx context.Context) (<-chan Workflow, error) {
	panic("not implemented")
}
