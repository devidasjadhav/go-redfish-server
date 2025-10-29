package models

import "time"

// TaskService represents the TaskService resource
type TaskService struct {
	Resource
	ServiceEnabled                  bool             `json:"ServiceEnabled,omitempty"`
	CompletedTaskOverWritePolicy    string           `json:"CompletedTaskOverWritePolicy,omitempty"`
	DateTime                        string           `json:"DateTime,omitempty"`
	LifeCycleEventOnTaskStateChange bool             `json:"LifeCycleEventOnTaskStateChange,omitempty"`
	TaskAutoDeleteTimeoutMinutes    int              `json:"TaskAutoDeleteTimeoutMinutes,omitempty"`
	Status                          Status           `json:"Status,omitempty"`
	Tasks                           TaskServiceTasks `json:"Tasks,omitempty"`
}

// TaskServiceTasks represents the Tasks link in TaskService
type TaskServiceTasks struct {
	ODataID string `json:"@odata.id"`
}

// NewTaskService creates a new TaskService instance
func NewTaskService() *TaskService {
	return &TaskService{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#TaskService.TaskService",
			ODataID:      "/redfish/v1/TaskService",
			ODataType:    "#TaskService.v1_2_1.TaskService",
			ID:           "TaskService",
			Name:         "Task Service",
		},
		ServiceEnabled:                  true,
		CompletedTaskOverWritePolicy:    "Manual",
		DateTime:                        time.Now().Format(time.RFC3339),
		LifeCycleEventOnTaskStateChange: true,
		TaskAutoDeleteTimeoutMinutes:    60,
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		Tasks: TaskServiceTasks{
			ODataID: "/redfish/v1/TaskService/Tasks",
		},
	}
}

// Task represents a task resource
type Task struct {
	Resource
	TaskState         string        `json:"TaskState"`
	TaskStatus        string        `json:"TaskStatus,omitempty"`
	StartTime         string        `json:"StartTime,omitempty"`
	EndTime           string        `json:"EndTime,omitempty"`
	PercentComplete   int           `json:"PercentComplete,omitempty"`
	TaskMonitor       string        `json:"TaskMonitor,omitempty"`
	Messages          []Message     `json:"Messages,omitempty"`
	Payload           *TaskPayload  `json:"Payload,omitempty"`
	HidePayload       bool          `json:"HidePayload,omitempty"`
	EstimatedDuration string        `json:"EstimatedDuration,omitempty"`
	SubTasks          *TaskSubTasks `json:"SubTasks,omitempty"`
	Links             TaskLinks     `json:"Links,omitempty"`
}

// TaskPayload represents the payload information for a task
type TaskPayload struct {
	TargetUri     string   `json:"TargetUri,omitempty"`
	HttpOperation string   `json:"HttpOperation,omitempty"`
	HttpHeaders   []string `json:"HttpHeaders,omitempty"`
	JsonBody      string   `json:"JsonBody,omitempty"`
}

// TaskSubTasks represents the SubTasks link in Task
type TaskSubTasks struct {
	ODataID string `json:"@odata.id"`
}

// TaskLinks represents the Links in Task
type TaskLinks struct {
	CreatedResources []ODataID   `json:"CreatedResources,omitempty"`
	Oem              interface{} `json:"Oem,omitempty"`
}

// NewTask creates a new Task instance
func NewTask(id string, operation string, targetUri string) *Task {
	now := time.Now().Format(time.RFC3339)
	return &Task{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#Task.Task",
			ODataID:      ODataID("/redfish/v1/TaskService/Tasks/" + id),
			ODataType:    "#Task.v1_7_4.Task",
			ID:           id,
			Name:         "Task " + id,
		},
		TaskState:       "New",
		TaskStatus:      "OK",
		StartTime:       now,
		PercentComplete: 0,
		TaskMonitor:     "/redfish/v1/TaskService/Tasks/" + id + "/Monitor",
		Messages:        []Message{},
		Payload: &TaskPayload{
			TargetUri:     targetUri,
			HttpOperation: operation,
		},
		HidePayload: false,
		Links: TaskLinks{
			CreatedResources: []ODataID{},
		},
	}
}

// UpdateTaskState updates the task state and related properties
func (t *Task) UpdateTaskState(newState string) {
	t.TaskState = newState

	switch newState {
	case "Running":
		if t.StartTime == "" {
			t.StartTime = time.Now().Format(time.RFC3339)
		}
	case "Completed", "Cancelled", "Exception":
		t.EndTime = time.Now().Format(time.RFC3339)
		if newState == "Completed" {
			t.PercentComplete = 100
		}
	}
}

// AddMessage adds a message to the task
func (t *Task) AddMessage(message Message) {
	t.Messages = append(t.Messages, message)
}

// SetPercentComplete sets the completion percentage
func (t *Task) SetPercentComplete(percent int) {
	if percent >= 0 && percent <= 100 {
		t.PercentComplete = percent
	}
}
