//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"

	//"fmt"

	"encoding/json"

	"github.com/go-resty/resty"
)

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

func New() Config {
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

//структура для теста маршалинга из json
type cadenceDomain struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	OwnerEmail  string `json:"ownerEmail"`
}

func (r *queryResolver) Domain(ctx context.Context, id string) (*Domain, error) {

	resp, err := resty.R().Get("http://10.174.18.121:8088/api/domain/axibpm-domain")

	var domain Domain

	if err == nil {

		var cdnDomian cadenceDomain

		cdnErr := json.Unmarshal(resp.Body(), &cdnDomian)

		if cdnErr == nil {
			domain = Domain{
				ID:        id,
				Name:      cdnDomian.Name,
				Workflows: nil,
			}
		} else {
			panic(cdnErr.Error())
		}
	} else {
		panic(err.Error())
	}

	return &domain, nil
}
func (r *queryResolver) AllDomains(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *string) ([]*Domain, error) {
	panic("not implemented")
}
func (r *queryResolver) Workflow(ctx context.Context, id string) (*Workflow, error) {
	panic("not implemented")
}
func (r *queryResolver) AllWorkflows(ctx context.Context, page *int, perPage *int, sortField *string, sortOrder *string, filter *string) ([]*Workflow, error) {
	panic("not implemented")
}

type subscriptionResolver struct{ *resolver }

func (r *subscriptionResolver) Workflow(ctx context.Context) (<-chan Workflow, error) {
	panic("not implemented")
}
