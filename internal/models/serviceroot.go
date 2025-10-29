package models

// ServiceRoot represents the root of the Redfish service
type ServiceRoot struct {
	Resource
	RedfishVersion string           `json:"RedfishVersion"`
	UUID           string           `json:"UUID,omitempty"`
	Systems        Link             `json:"Systems,omitempty"`
	Chassis        Link             `json:"Chassis,omitempty"`
	Managers       Link             `json:"Managers,omitempty"`
	Tasks          Link             `json:"Tasks,omitempty"`
	SessionService Link             `json:"SessionService,omitempty"`
	AccountService Link             `json:"AccountService,omitempty"`
	EventService   Link             `json:"EventService,omitempty"`
	Registries     Link             `json:"Registries,omitempty"`
	JsonSchemas    Link             `json:"JsonSchemas,omitempty"`
	UpdateService  Link             `json:"UpdateService,omitempty"`
	Links          ServiceRootLinks `json:"Links,omitempty"`
}

// ServiceRootLinks represents the links in the ServiceRoot
type ServiceRootLinks struct {
	Sessions Link `json:"Sessions,omitempty"`
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
		Systems:        Link{ODataID: "/redfish/v1/Systems"},
		Chassis:        Link{ODataID: "/redfish/v1/Chassis"},
		Managers:       Link{ODataID: "/redfish/v1/Managers"},
		Tasks:          Link{ODataID: "/redfish/v1/TaskService"},
		SessionService: Link{ODataID: "/redfish/v1/SessionService"},
		AccountService: Link{ODataID: "/redfish/v1/AccountService"},
		EventService:   Link{ODataID: "/redfish/v1/EventService"},
		Registries:     Link{ODataID: "/redfish/v1/Registries"},
		JsonSchemas:    Link{ODataID: "/redfish/v1/JsonSchemas"},
		Links: ServiceRootLinks{
			Sessions: Link{ODataID: "/redfish/v1/SessionService/Sessions"},
		},
	}
}
