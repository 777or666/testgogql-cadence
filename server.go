package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/777or666/testgogql-cadence/models"
	"github.com/99designs/gqlgen/handler"
	//gqlopentracing "github.com/99designs/gqlgen/opentracing"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/cors"
	"sourcegraph.com/sourcegraph/appdash"
	appdashtracer "sourcegraph.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash/traceapp"

	"github.com/777or666/testgogql-cadence/helpers"

	"go.uber.org/cadence/worker"

	"github.com/777or666/testgogql-cadence/axibpmActivities"
	"github.com/777or666/testgogql-cadence/axibpmWorkflows"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

const (
	configFile = "config/appsettings.yaml"
)

type Configuration struct {
	UrlRestService     string `yaml:"urlrestservice"`
	ApplicationName    string `yaml:"appname"`
	PrefixWorkflowFunc string `yaml:"prefixworkflowfunc"`
	EmailIdentity      string `yaml:"emailidentity"`
	EmailFrom          string `yaml:"emailfrom"`
	EmailUserName      string `yaml:"emailusername"`
	EmailPassword      string `yaml:"emailpassword"`
	EmailHost          string `yaml:"emailhost"`
	EmailPort          string `yaml:"emailport"`
}

//****************************************

// Регистрируем все воркфлоу и активности
func init() {
	workflow.Register(axibpmWorkflows.TestWorkflow)
	activity.Register(axibpmActivities.TestActivity)
	activity.Register(axibpmActivities.EmailSenderActivity)
}

//****************************************

func main() {
	startAppdashServer()

	//Чтение файла конфигурации
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Ошибка чтения файла: %v, Error: %v", configFile, err))
	}

	var Config Configuration

	if err := yaml.Unmarshal(configData, &Config); err != nil {
		panic(fmt.Sprintf("Ошибка инициализации конфигурации: %v", err))
	}
	//*******************************

	emailconf := helpers.EmailConfig{
		Emailidentity: Config.EmailIdentity,
		Emailfrom:     Config.EmailFrom,
		Emailusername: Config.EmailUserName,
		Emailpassword: Config.EmailPassword,
		Emailhost:     Config.EmailHost,
		Emailport:     Config.EmailPort,
	}

	//Запуск воркера бизнес-процессов
	var h helpers.SampleHelper
	h.SetupServiceConfig()

	startWorkers(&h, Config.ApplicationName)
	//**********************************

	//Запуск системы graphql
	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("AXI-BPM CADENCE", "/query"))
	router.HandleFunc("/query", handler.GraphQL(
		models.NewExecutableSchema(
			models.New(
				Config.UrlRestService,
				Config.ApplicationName,
				Config.PrefixWorkflowFunc,
				&emailconf,
				&h,
			)),
		//handler.ResolverMiddleware(gqlopentracing.ResolverMiddleware()),
		//handler.RequestMiddleware(gqlopentracing.RequestMiddleware()),
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
	//**************************************
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

// Запуск воркера
func startWorkers(h *helpers.SampleHelper, ApplicationName string) {

	// Конфигурация воркера
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}
