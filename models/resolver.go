//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"
	"log"
	"sync"
	"time"

	"github.com/777or666/testgogql-cadence/axibpmActivities"
	"github.com/777or666/testgogql-cadence/axibpmWorkflows"

	"github.com/777or666/testgogql-cadence/helpers"

	"encoding/json"

	"github.com/go-resty/resty"
	//"github.com/pborman/uuid"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

var UrlRestService string
var h helpers.SampleHelper
var workflowClient client.Client
var ApplicationName string //ВСЕГДА должно совпадать в воркере и всех его воркфлоу

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

//****************************************
// Регистрируем все воркфлоу и активности
func init() {
	workflow.Register(axibpmWorkflows.TestWorkflow)
	activity.Register(axibpmActivities.TestActivity)
}

//****************************************

func New(urlRestService string, applicationName string) Config {

	UrlRestService = urlRestService
	ApplicationName = applicationName

	h.SetupServiceConfig()
	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}
	startWorkers(&h)

	return Config{Resolvers: &resolver{}}
}

type mutationResolver struct{ *resolver }

// Запускаем воркера
func startWorkers(h *helpers.SampleHelper) {
	// Конфигурация воркера
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

//Запуск задачи
//id -идентификатор
//name - программное наименование функции воркфлоу с пакетом (пример, "github.com/777or666/testgogql-cadence/axibpmWorkflows.TestWorkflow")
//taskList - наименование типа списка задач
func startWorkflow(h *helpers.SampleHelper, id string, name string, token *string) (string, string) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              id,
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    2 * time.Minute,
		DecisionTaskStartToCloseTimeout: 2 * time.Minute,
	}

	log.Println("startWorkflow! " + name)

	wfid, rid := h.StartWorkflow(workflowOptions, name, id, token)

	return wfid, rid
}

//TODO: input возможно не нужны!
//TODO: taskList точно не нужен, так как должен точно совпадать с аппнэймом в воркере!
func (r *mutationResolver) WorkflowStart(ctx context.Context, id string, name string, taskList string, input *string) (Workflow, error) {
	//r.mu.Lock()

	token := new(string)

	//Запускаем задачу
	wfId, wfRunId := startWorkflow(&h, id, name, token)

	//r.mu.Unlock()

	cteatedt := time.Now()

	var x []Activity
	x = make([]Activity, 1)
	x[0] = Activity{
		ID:    "123123123123123",
		Token: *token,
	}

	//TODO: продумать состав, надо пробросить тамауты, инпуты и т.д.
	wrf := Workflow{
		ID:         id,
		Name:       name,
		WorkflowID: wfId,
		RunID:      wfRunId,
		TaskList:   taskList,
		CreatedAt:  cteatedt,
		Activities: x,
	}

	return wrf, nil // nil заменить на err
}
func (r *mutationResolver) WorkflowCancel(ctx context.Context, id string) (Workflow, error) {
	panic("not implemented")
}
func (r *mutationResolver) ActivityApproval(ctx context.Context, token string) (*bool, error) {

	state := "APPROVED"

	log.Println("пришел token: " + token)

	var result bool

	err := workflowClient.CompleteActivity(context.Background(), []byte(token), state, nil)
	if err != nil {

		log.Println("ОШИБКА! Задача не выполнена. " + err.Error())
		result = false
		return nil, err
	}

	result = true
	return &result, nil
}
func (r *mutationResolver) ActivityReject(ctx context.Context, token string) (*bool, error) {
	panic("not implemented")
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
