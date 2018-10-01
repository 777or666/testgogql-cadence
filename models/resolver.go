//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"
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

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) StartWorkflow(ctx context.Context, id string, name string, taskList string, input *string) (Workflow, error) {
	panic("not implemented")
}
func (r *mutationResolver) CancelWorkflow(ctx context.Context, id string) (Workflow, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Domain(ctx context.Context, id string) (*Domain, error) {
	panic("not implemented")
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

type subscriptionResolver struct{ *Resolver }

func (r *subscriptionResolver) Workflow(ctx context.Context) (<-chan Workflow, error) {
	panic("not implemented")
}
