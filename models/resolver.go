//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"
	"log"
	"sync"
	"time"

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
//ExecutionStartToCloseTimeout - тайм-айт выполнения рабочего процесса
//DecisionTaskStartToCloseTimeout - тайм-аут для обработки задачи решения с момента, когда воркер вытащил эту задачу
//EmailResponsible - e-mail ответственных
//EmailParticipants - e-mail остальных участников процесса
func (r *mutationResolver) WorkflowStart(ctx context.Context, id string, name string, input *string, ExecutionStartToCloseTimeout *int, DecisionTaskStartToCloseTimeout *int, EmailResponsible []*string, EmailParticipants []*string) (Workflow, error) {
	//r.mu.Lock()

	//token := new(string)

	//Запускаем задачу

	//r.mu.Unlock()

	workflowOptions := client.StartWorkflowOptions{
		ID:       id,
		TaskList: ApplicationName,
		//ExecutionStartToCloseTimeout - тайм-аут выполнения рабочего процесса
		ExecutionStartToCloseTimeout: time.Duration(*ExecutionStartToCloseTimeout) * time.Minute,
		//DecisionTaskTartToCloseTimeout - тайм-аут для обработки задачи решения с момента, когда рабочий
		// вытащил эту задачу. Если задача решения потеряна, она повторится после этого таймаута.
		DecisionTaskStartToCloseTimeout: time.Duration(*DecisionTaskStartToCloseTimeout) * time.Minute,
		WorkflowIDReusePolicy:           2, // см. ниже
		// 0
		// WorkflowIDReusePolicyAllowDuplicateFailedOnly позволяет запустить выполнение рабочего процесса
		// когда рабочий процесс не запущен, а состояние завершения последнего выполнения находится в
		// [завершено, отменено, время ожидания, не выполнено].
		// WorkflowIDReusePolicyAllowDuplicateFailedOnly WorkflowIDReusePolicy = iota
		// 1
		// WorkflowIDReusePolicyAllowDuplicate позволяет запустить выполнение рабочего процесса, используя
		// тот же идентификатор рабочего процесса, когда рабочий процесс не запущен.
		//WorkflowIDReusePolicyAllowDuplicate
		// 2
		// WorkflowIDReusePolicyRejectDuplicate не позволяет запустить выполнение рабочего процесса с использованием того же идентификатора рабочего процесса вообще
		//WorkflowIDReusePolicyRejectDuplicate
	}

	//log.Println("startWorkflow! " + name)

	fullname := PrefixWorkflowFunc + name

	wfId, wfRunId := h.StartWorkflow(workflowOptions, fullname, id, EmailConfiguration, EmailResponsible, EmailParticipants)

	cteatedAt := time.Now()

	activityID := "0"
	isApproved := false

	var x []Activity
	x = make([]Activity, 1)
	x[0] = Activity{
		ID:         "",
		ActivityID: &activityID,
		RunID:      &wfRunId,
		//Token:      token,
		IsApproved: &isApproved,
	}

	//TODO: продумать состав, надо пробросить тамауты, инпуты и т.д.
	wrf := Workflow{
		ID:         id,
		Name:       name,
		WorkflowID: wfId,
		RunID:      wfRunId,
		CreatedAt:  cteatedAt,
		Activities: x,
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

func (r *mutationResolver) WorkflowTerminate(ctx context.Context, id string, runID *string, reason *string, info *string) (*string, error) {
	result := "Бизнес-процесс прерван"
	details := *info

	err := workflowClient.TerminateWorkflow(context.Background(), id, *runID, *reason, []byte(details))
	if err != nil {
		log.Println("ОШИБКА! Не удалось прервать бизнес-процесс. " + err.Error())

		return nil, err
	}

	return &result, nil
}
func (r *mutationResolver) ActivityPerform(ctx context.Context, domain *string, workflowID string, runID *string, activityID string, info *string) (*string, error) {

	result := "Операция выполнена"

	log.Println("ActivityPerform! Старт")

	err := workflowClient.CompleteActivityByID(context.Background(), *domain, workflowID, *runID, activityID, *info, nil)
	if err != nil {

		log.Println("ОШИБКА! Не удалось выполнить операцию. " + err.Error())

		return nil, err
	}

	return &result, nil
}
func (r *mutationResolver) ActivityFailed(ctx context.Context, domain *string, workflowID string, runID *string, activityID string, info *string) (*string, error) {

	result := "Операция отменена"

	//добавить детали - причину отмены активности и кто ее отменил
	err := workflowClient.CompleteActivityByID(context.Background(), *domain, workflowID, *runID, activityID, nil, cadence.NewCustomError(*info, result))
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
