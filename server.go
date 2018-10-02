package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	//"github.com/99designs/gqlgen/example/chat"

	"github.com/777or666/testgogql-cadence/models"
	"github.com/99designs/gqlgen/handler"
	gqlopentracing "github.com/99designs/gqlgen/opentracing"
	//"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/cors"
	"sourcegraph.com/sourcegraph/appdash"
	appdashtracer "sourcegraph.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

func main() {
	startAppdashServer()

	router := mux.NewRouter()

	router.HandleFunc("/", handler.Playground("Todo", "/query"))
	router.HandleFunc("/query", handler.GraphQL(models.NewExecutableSchema(models.New()),
		handler.ResolverMiddleware(gqlopentracing.ResolverMiddleware()),
		handler.RequestMiddleware(gqlopentracing.RequestMiddleware()),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		})),
	)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Fatal(http.ListenAndServe(":8085", handler))
}

func startAppdashServer() opentracing.Tracer {
	memStore := appdash.NewMemoryStore()
	store := &appdash.RecentStore{
		MinEvictAge: 5 * time.Minute,
		DeleteStore: memStore,
	}

	url, err := url.Parse("http://localhost:8700")
	if err != nil {
		log.Fatal(err)
	}
	tapp, err := traceapp.New(nil, url)
	if err != nil {
		log.Fatal(err)
	}
	tapp.Store = store
	tapp.Queryer = memStore

	go func() {
		log.Fatal(http.ListenAndServe(":8700", tapp))
	}()
	tapp.Store = store
	tapp.Queryer = memStore

	collector := appdash.NewLocalCollector(store)
	tracer := appdashtracer.NewTracer(collector)
	opentracing.InitGlobalTracer(tracer)

	log.Println("Appdash web UI running on HTTP :8700")
	return tracer
}
