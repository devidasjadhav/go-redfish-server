package models

// ServiceRoot represents the root of the Redfish service
type ServiceRoot struct {
	Resource
	RedfishVersion string           `json:"RedfishVersion"`
	UUID           string           `json:"UUID,omitempty"`
	Systems        ODataID          `json:"Systems,omitempty"`
	Chassis        ODataID          `json:"Chassis,omitempty"`
	Managers       ODataID          `json:"Managers,omitempty"`
	Tasks          ODataID          `json:"Tasks,omitempty"`
	SessionService ODataID          `json:"SessionService,omitempty"`
	AccountService ODataID          `json:"AccountService,omitempty"`
	EventService   ODataID          `json:"EventService,omitempty"`
	Registries     ODataID          `json:"Registries,omitempty"`
	JsonSchemas    ODataID          `json:"JsonSchemas,omitempty"`
	UpdateService  ODataID          `json:"UpdateService,omitempty"`
	Links          ServiceRootLinks `json:"Links,omitempty"`
}

// ServiceRootLinks represents the links in the ServiceRoot
type ServiceRootLinks struct {
	Sessions ODataID `json:"Sessions,omitempty"`
}

// NewServiceRoot creates a new ServiceRoot instance
func NewServiceRoot() *ServiceRoot {
	return &ServiceRoot{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
			ODataID:      "/redfish/v1/",
			ODataType:    "#ServiceRoot.v1_15_0.ServiceRoot",
			ID:           "RootService",
			Name:         "Root Service",
		},
		RedfishVersion: "1.15.0",
		UUID:           "00000000-0000-0000-0000-000000000000",
		Systems:        "/redfish/v1/Systems",
		Chassis:        "/redfish/v1/Chassis",
		Managers:       "/redfish/v1/Managers",
		Tasks:          "/redfish/v1/TaskService",
		SessionService: "/redfish/v1/SessionService",
		AccountService: "/redfish/v1/AccountService",
		EventService:   "/redfish/v1/EventService",
		Registries:     "/redfish/v1/Registries",
		JsonSchemas:    "/redfish/v1/JsonSchemas",
		Links: ServiceRootLinks{
			Sessions: "/redfish/v1/SessionService/Sessions",
		},
	}
}
