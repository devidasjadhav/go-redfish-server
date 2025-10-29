package models

// EventService represents the EventService resource
type EventService struct {
	Resource
	ServiceEnabled                    bool              `json:"ServiceEnabled,omitempty"`
	DeliveryRetryAttempts             int               `json:"DeliveryRetryAttempts,omitempty"`
	DeliveryRetryIntervalSeconds      int               `json:"DeliveryRetryIntervalSeconds,omitempty"`
	EventFormatTypes                  []string          `json:"EventFormatTypes,omitempty"`
	ExcludeMessageId                  bool              `json:"ExcludeMessageId,omitempty"`
	ExcludeRegistryPrefix             bool              `json:"ExcludeRegistryPrefix,omitempty"`
	IncludeOriginOfConditionSupported bool              `json:"IncludeOriginOfConditionSupported,omitempty"`
	RegistryPrefixes                  []string          `json:"RegistryPrefixes,omitempty"`
	ResourceTypes                     []string          `json:"ResourceTypes,omitempty"`
	ServerSentEventUri                string            `json:"ServerSentEventUri,omitempty"`
	Severities                        []string          `json:"Severities,omitempty"`
	Status                            Status            `json:"Status,omitempty"`
	Actions                           Actions           `json:"Actions,omitempty"`
	Links                             EventServiceLinks `json:"Links,omitempty"`
}

// EventServiceLinks represents the links in the EventService
type EventServiceLinks struct {
	Subscriptions ODataID `json:"Subscriptions,omitempty"`
}

// NewEventService creates a new EventService instance
func NewEventService() *EventService {
	return &EventService{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#EventService.EventService",
			ODataID:      "/redfish/v1/EventService",
			ODataType:    "#EventService.v1_11_0.EventService",
			ID:           "EventService",
			Name:         "Event Service",
		},
		ServiceEnabled:                    true,
		DeliveryRetryAttempts:             3,
		DeliveryRetryIntervalSeconds:      60,
		EventFormatTypes:                  []string{"Event"},
		ExcludeMessageId:                  false,
		ExcludeRegistryPrefix:             false,
		IncludeOriginOfConditionSupported: true,
		RegistryPrefixes:                  []string{"Base", "Task"},
		ResourceTypes:                     []string{"ComputerSystem", "Manager", "Chassis"},
		ServerSentEventUri:                "/redfish/v1/EventService/SSE",
		Severities:                        []string{"OK", "Warning", "Critical"},
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		Actions: Actions{
			Oem: map[string]interface{}{},
		},
		Links: EventServiceLinks{
			Subscriptions: "/redfish/v1/EventService/Subscriptions",
		},
	}
}

// EventSubscription represents an event subscription (EventDestination)
type EventSubscription struct {
	Resource
	Context                  string       `json:"Context,omitempty"`
	DeliveryRetryPolicy      string       `json:"DeliveryRetryPolicy,omitempty"`
	Destination              string       `json:"Destination"`
	EventFormatType          string       `json:"EventFormatType,omitempty"`
	ExcludeMessageIds        []string     `json:"ExcludeMessageIds,omitempty"`
	ExcludeRegistryPrefixes  []string     `json:"ExcludeRegistryPrefixes,omitempty"`
	HttpHeaders              []HttpHeader `json:"HttpHeaders,omitempty"`
	IncludeOriginOfCondition bool         `json:"IncludeOriginOfCondition,omitempty"`
	MessageIds               []string     `json:"MessageIds,omitempty"`
	OriginResources          []ODataID    `json:"OriginResources,omitempty"`
	Protocol                 string       `json:"Protocol"`
	RegistryPrefixes         []string     `json:"RegistryPrefixes,omitempty"`
	ResourceTypes            []string     `json:"ResourceTypes,omitempty"`
	Severities               []string     `json:"Severities,omitempty"`
	Status                   Status       `json:"Status,omitempty"`
	SubordinateResources     bool         `json:"SubordinateResources,omitempty"`
	SubscriptionType         string       `json:"SubscriptionType"`
	Actions                  Actions      `json:"Actions,omitempty"`
}

// HttpHeader represents an HTTP header for event delivery
type HttpHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NewEventSubscription creates a new EventSubscription instance
func NewEventSubscription(id string, destination string, protocol string) *EventSubscription {
	return &EventSubscription{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#EventDestination.EventDestination",
			ODataID:      ODataID("/redfish/v1/EventService/Subscriptions/" + id),
			ODataType:    "#EventDestination.v1_15_1.EventDestination",
			ID:           id,
			Name:         "Event Subscription " + id,
		},
		Destination:              destination,
		Protocol:                 protocol,
		SubscriptionType:         "RedfishEvent",
		EventFormatType:          "Event",
		IncludeOriginOfCondition: false,
		SubordinateResources:     false,
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		Actions: Actions{
			Oem: map[string]interface{}{},
		},
	}
}

// Event represents the payload sent to event subscribers
type Event struct {
	ODataContext string        `json:"@odata.context,omitempty"`
	ODataType    string        `json:"@odata.type,omitempty"`
	ID           string        `json:"Id"`
	Name         string        `json:"Name"`
	Context      string        `json:"Context,omitempty"`
	Events       []EventRecord `json:"Events"`
}

// EventRecord represents a single event in the Events array
type EventRecord struct {
	EventType         string      `json:"EventType,omitempty"`
	EventId           string      `json:"EventId"`
	EventTimestamp    string      `json:"EventTimestamp"`
	Severity          string      `json:"Severity,omitempty"`
	Message           string      `json:"Message,omitempty"`
	MessageId         string      `json:"MessageId"`
	MessageArgs       []string    `json:"MessageArgs,omitempty"`
	MessageSeverity   string      `json:"MessageSeverity,omitempty"`
	OriginOfCondition *ODataID    `json:"OriginOfCondition,omitempty"`
	Resolution        string      `json:"Resolution,omitempty"`
	MemberId          string      `json:"MemberId"`
	Oem               interface{} `json:"Oem,omitempty"`
}

// NewEvent creates a new Event payload
func NewEvent(context string, events []EventRecord) *Event {
	return &Event{
		ODataContext: "/redfish/v1/$metadata#Event.Event",
		ODataType:    "#Event.v1_12_0.Event",
		ID:           "Event",
		Name:         "Event",
		Context:      context,
		Events:       events,
	}
}
