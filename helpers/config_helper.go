package helpers

type EmailConfig struct {
	Emailidentity string
	Emailfrom     string
	Emailusername string
	Emailpassword string
	Emailhost     string
	Emailport     string
}

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

type WorkflowInput struct {
	UserData     WorkflowInputUser   `json:"user"`
	WorkflowData WorkflowInputObject `json:"workflowdata"`
}
