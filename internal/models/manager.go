package models

// Manager represents a management controller
type Manager struct {
	Resource
	ManagerType           string         `json:"ManagerType,omitempty"` // BMC, EnclosureManager, etc.
	FirmwareVersion       string         `json:"FirmwareVersion,omitempty"`
	Status                Status         `json:"Status,omitempty"`
	PowerState            string         `json:"PowerState,omitempty"`
	ServiceIdentification string         `json:"ServiceIdentification,omitempty"`
	UUID                  string         `json:"UUID,omitempty"`
	Model                 string         `json:"Model,omitempty"`
	DateTime              string         `json:"DateTime,omitempty"` // ISO 8601 format
	DateTimeLocalOffset   string         `json:"DateTimeLocalOffset,omitempty"`
	NetworkProtocol       ODataID        `json:"NetworkProtocol,omitempty"`
	EthernetInterfaces    ODataID        `json:"EthernetInterfaces,omitempty"`
	SerialInterfaces      ODataID        `json:"SerialInterfaces,omitempty"`
	LogServices           ODataID        `json:"LogServices,omitempty"`
	VirtualMedia          ODataID        `json:"VirtualMedia,omitempty"`
	Links                 ManagerLinks   `json:"Links,omitempty"`
	Actions               ManagerActions `json:"Actions,omitempty"`
}

// ManagerLinks represents links to related resources
type ManagerLinks struct {
	ManagerForServers []ODataID `json:"ManagerForServers,omitempty"`
	ManagerForChassis []ODataID `json:"ManagerForChassis,omitempty"`
	ManagerInChassis  ODataID   `json:"ManagerInChassis,omitempty"`
	Oem               Oem       `json:"Oem,omitempty"`
}

// ManagerActions represents available actions
type ManagerActions struct {
	ManagerReset struct {
		Target string `json:"target"`
		Title  string `json:"title,omitempty"`
	} `json:"#Manager.Reset,omitempty"`
	ManagerForceFailover struct {
		Target string `json:"target"`
		Title  string `json:"title,omitempty"`
	} `json:"#Manager.ForceFailover,omitempty"`
	Oem Oem `json:"Oem,omitempty"`
}

// NewManager creates a new Manager instance
func NewManager(id string) *Manager {
	return &Manager{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#Manager.Manager",
			ODataID:      ODataID("/redfish/v1/Managers/" + id),
			ODataType:    "#Manager.v1_20_0.Manager",
			ID:           id,
			Name:         "Manager",
		},
		ManagerType:     "BMC",
		FirmwareVersion: "1.0.0",
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		PowerState:            "On",
		ServiceIdentification: "BMC",
		UUID:                  "00000000-0000-0000-0000-000000000001",
		Model:                 "Baseboard Management Controller",
		DateTime:              "2025-10-29T18:48:45+00:00",
		DateTimeLocalOffset:   "+00:00",
		NetworkProtocol:       ODataID("/redfish/v1/Managers/" + id + "/NetworkProtocol"),
		EthernetInterfaces:    ODataID("/redfish/v1/Managers/" + id + "/EthernetInterfaces"),
		LogServices:           ODataID("/redfish/v1/Managers/" + id + "/LogServices"),
		Links: ManagerLinks{
			ManagerForServers: []ODataID{ODataID("/redfish/v1/Systems/1")},
			ManagerForChassis: []ODataID{ODataID("/redfish/v1/Chassis/1")},
		},
		Actions: ManagerActions{
			ManagerReset: struct {
				Target string `json:"target"`
				Title  string `json:"title,omitempty"`
			}{
				Target: "/redfish/v1/Managers/" + id + "/Actions/Manager.Reset",
				Title:  "Reset Manager",
			},
		},
	}
}

// ManagerCollection represents a collection of managers
type ManagerCollection struct {
	Collection
}

// NewManagerCollection creates a new ManagerCollection instance
func NewManagerCollection() *ManagerCollection {
	return &ManagerCollection{
		Collection: Collection{
			ODataContext:      "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
			ODataID:           "/redfish/v1/Managers",
			ODataType:         "#ManagerCollection.ManagerCollection",
			Name:              "Manager Collection",
			Members:           []ODataID{"/redfish/v1/Managers/1"},
			MembersODataCount: 1,
		},
	}
}
