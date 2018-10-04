//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"

	//"log"

	"encoding/json"

	"github.com/go-resty/resty"
)

var UrlRestService string

type resolver struct{}

func (r *resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

func New(urlRestService string) Config {

	UrlRestService = urlRestService

	return Config{Resolvers: &resolver{}}
}

type mutationResolver struct{ *resolver }

func (r *mutationResolver) StartWorkflow(ctx context.Context, id string, name string, taskList string, input *string) (Workflow, error) {
	panic("not implemented")
}
func (r *mutationResolver) CancelWorkflow(ctx context.Context, id string) (Workflow, error) {
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
