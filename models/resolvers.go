//go:generate gorunpkg github.com/99designs/gqlgen

package models

import (
	context "context"
	"math/rand"
	"sync"
	"time"

	"../workflow"

	"cadence_samples/common"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
)

var h common.SampleHelper

type resolver struct {
	Rooms map[string]*Chatroom
	mu    sync.Mutex // nolint: structcheck
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

func New() Config {
	h.SetupServiceConfig()
	startWorkers(&h)
	return Config{
		Resolvers: &resolver{
			Rooms: map[string]*Chatroom{},
		},
	}
}

type Chatroom struct {
	Name      string
	Messages  []Message
	Observers map[string]chan Message
}

type mutationResolver struct{ *resolver }

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, workflow.ApplicationName, workerOptions)
}

//Запуск тестовой задачи
func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "helloworld_" + uuid.New(),
		TaskList:                        workflow.ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, workflow.Workflow, "Согласовать ТКП")
}

func (r *mutationResolver) Post(ctx context.Context, text string, username string, roomName string) (Message, error) {
	r.mu.Lock()
	room := r.Rooms[roomName]
	if room == nil {
		room = &Chatroom{Name: roomName, Observers: map[string]chan Message{}}
		r.Rooms[roomName] = room
	}
	r.mu.Unlock()

	message := Message{
		ID:        randString(8),
		CreatedAt: time.Now(),
		Text:      text,
		CreatedBy: username,
	}

	room.Messages = append(room.Messages, message)
	r.mu.Lock()
	for _, observer := range room.Observers {
		observer <- message
	}
	r.mu.Unlock()

	//Запускаем задачу
	startWorkflow(&h)

	return message, nil
}

type queryResolver struct{ *resolver }

func (r *queryResolver) Room(ctx context.Context, name string) (*Chatroom, error) {
	r.mu.Lock()
	room := r.Rooms[name]
	if room == nil {
		room = &Chatroom{Name: name, Observers: map[string]chan Message{}}
		r.Rooms[name] = room
	}
	r.mu.Unlock()

	return room, nil
}

type subscriptionResolver struct{ *resolver }

func (r *subscriptionResolver) MessageAdded(ctx context.Context, roomName string) (<-chan Message, error) {
	r.mu.Lock()
	room := r.Rooms[roomName]
	if room == nil {
		room = &Chatroom{Name: roomName, Observers: map[string]chan Message{}}
		r.Rooms[roomName] = room
	}
	r.mu.Unlock()

	id := randString(8)
	events := make(chan Message, 1)

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(room.Observers, id)
		r.mu.Unlock()
	}()

	r.mu.Lock()
	room.Observers[id] = events
	r.mu.Unlock()

	return events, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
