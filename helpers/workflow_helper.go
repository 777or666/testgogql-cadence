package helpers

//**типы для чтения конфигураций****
type WorkflowConfiguration struct {
	WorkflowName     string              `yaml:"workflowname"`
	WorkflowActivity WorkflowActivityMap `yaml:"activity"`
}

type WorkflowActivityConfig struct {
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

//**************************************
