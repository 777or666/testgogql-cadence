package helpers

//КОНФИГУРАЦИЯ Е-МЭИЛ
type EmailConfig struct {
	Emailidentity string
	Emailfrom     string
	Emailusername string
	Emailpassword string
	Emailhost     string
	Emailport     string
}

//КОНФИГУРАЦИЯ ФАЙЛОВ YAML НАСТРОЕК WORKFLOW
type WorkflowConfiguration struct {
	WorkflowName     string              `yaml:"workflowname"`
	WorkflowActivity WorkflowActivityMap `yaml:"activity"`
}

type WorkflowActivityConfig struct {
	ActivityId             string                      `yaml:"activityid"`
	Description            string                      `yaml:"description"`
	Operation              string                      `yaml:"operation"`
	Roles                  []string                    `yaml:"roles"`
	ScheduleToStartTimeout int                         `yaml:"scheduletostarttimeout"`
	StartToCloseTimeout    int                         `yaml:"starttoclosetimeout"`
	HeartbeatTimeout       int                         `yaml:"heartbeattimeout"`
	ActivityReminders      WorkflowActivityReminderMap `yaml:"activityreminders"`
}

type WorkflowActivityReminder struct {
	ReminderTime int    `yaml:"remindertime"`
	ReminderText string `yaml:"remindertext"`
}

type WorkflowActivityReminderMap map[int]WorkflowActivityReminder

type WorkflowActivityMap map[int]WorkflowActivityConfig

//КОНФИГУРАЦИЯ INPUT В РЕЗОЛВЕРАХ + КОНФИГУРАЦИЯ ИЗ ФАЙЛОВ
type WorkflowInputUser struct {
	Useremail  string `json:"useremail"`
	Username   string `json:"username"`
	Department string `json:"department"`
}

type WorkflowInputObject struct {
	ObjectId   string `json:"objectId"`
	ObjectHref string `json:"objectHref"`
	ObjectName string `json:"objectName"`
	ObjectType string `json:"objectType"`
	Comment    string `json:"comment"`
}

type WorkflowInputSettings struct {
	WorkflowId                      string `json:"workflowId"`
	ExecutionStartToCloseTimeout    int    `json:"executionStartToCloseTimeout"`
	DecisionTaskStartToCloseTimeout int    `json:"decisionTaskStartToCloseTimeout"`
}

type WorkflowInputEmailAdresses struct {
	EmailResponsible  []string `json:"emailResponsible"`
	EmailParticipants []string `json:"emailParticipants"`
}

type WorkflowInput struct {
	UserData            WorkflowInputUser          `json:"user"`
	WorkflowData        WorkflowInputObject        `json:"workflowdata"`
	WorkflowSettings    WorkflowInputSettings      `json:"workflowsettings"`
	WorkflowEmails      WorkflowInputEmailAdresses `json:"workflowemails"`
	WorkflowEmailConfig EmailConfig                `json:"emailconfig"`
	WorkflowConfig      WorkflowConfiguration      `json:"workflowconfiguration"`
}
