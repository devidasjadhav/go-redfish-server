package models

// ComputerSystem represents a computer system (physical or virtual)
type ComputerSystem struct {
	Resource
	SystemType         string                `json:"SystemType,omitempty"` // Physical, Virtual, etc.
	AssetTag           string                `json:"AssetTag,omitempty"`
	Manufacturer       string                `json:"Manufacturer,omitempty"`
	Model              string                `json:"Model,omitempty"`
	SKU                string                `json:"SKU,omitempty"`
	SerialNumber       string                `json:"SerialNumber,omitempty"`
	PartNumber         string                `json:"PartNumber,omitempty"`
	UUID               string                `json:"UUID,omitempty"`
	HostName           string                `json:"HostName,omitempty"`
	Status             Status                `json:"Status,omitempty"`
	PowerState         string                `json:"PowerState,omitempty"` // On, Off, PoweringOn, etc.
	Boot               Boot                  `json:"Boot,omitempty"`
	BiosVersion        string                `json:"BiosVersion,omitempty"`
	ProcessorSummary   ProcessorSummary      `json:"ProcessorSummary,omitempty"`
	MemorySummary      MemorySummary         `json:"MemorySummary,omitempty"`
	Storage            StorageSummary        `json:"Storage,omitempty"`
	Processors         ODataID               `json:"Processors,omitempty"`
	Memory             ODataID               `json:"Memory,omitempty"`
	StorageControllers ODataID               `json:"StorageControllers,omitempty"`
	NetworkInterfaces  ODataID               `json:"NetworkInterfaces,omitempty"`
	EthernetInterfaces ODataID               `json:"EthernetInterfaces,omitempty"`
	LogServices        ODataID               `json:"LogServices,omitempty"`
	Links              ComputerSystemLinks   `json:"Links,omitempty"`
	Actions            ComputerSystemActions `json:"Actions,omitempty"`
	Oem                *OEM                  `json:"Oem,omitempty"`
}

// Boot represents boot configuration
type Boot struct {
	BootSourceOverrideEnabled    string `json:"BootSourceOverrideEnabled,omitempty"` // Once, Continuous, Disabled
	BootSourceOverrideTarget     string `json:"BootSourceOverrideTarget,omitempty"`  // None, Pxe, etc.
	BootSourceOverrideMode       string `json:"BootSourceOverrideMode,omitempty"`    // Legacy, UEFI
	UefiTargetBootSourceOverride string `json:"UefiTargetBootSourceOverride,omitempty"`
}

// ProcessorSummary represents processor information
type ProcessorSummary struct {
	Count  int    `json:"Count,omitempty"`
	Model  string `json:"Model,omitempty"`
	Status Status `json:"Status,omitempty"`
}

// MemorySummary represents memory information
type MemorySummary struct {
	TotalSystemMemoryGiB float64 `json:"TotalSystemMemoryGiB,omitempty"`
	Status               Status  `json:"Status,omitempty"`
}

// StorageSummary represents storage information
type StorageSummary struct {
	Controllers ODataID `json:"Controllers,omitempty"`
}

// NetworkInterfaces represents network interface information
type NetworkInterfaces struct {
	ODataID ODataID `json:"@odata.id,omitempty"`
}

// EthernetInterfaces represents ethernet interface information
type EthernetInterfaces struct {
	ODataID ODataID `json:"@odata.id,omitempty"`
}

// ComputerSystemLinks represents links to related resources
type ComputerSystemLinks struct {
	Chassis   []ODataID `json:"Chassis,omitempty"`
	ManagedBy []ODataID `json:"ManagedBy,omitempty"`
	Oem       Oem       `json:"Oem,omitempty"`
}

// ComputerSystemActions represents available actions
type ComputerSystemActions struct {
	ComputerSystemReset struct {
		Target string `json:"target"`
		Title  string `json:"title,omitempty"`
	} `json:"#ComputerSystem.Reset,omitempty"`
	Oem Oem `json:"Oem,omitempty"`
}

// NewComputerSystem creates a new ComputerSystem instance
func NewComputerSystem(id string) *ComputerSystem {
	return &ComputerSystem{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
			ODataID:      ODataID("/redfish/v1/Systems/" + id),
			ODataType:    "#ComputerSystem.v1_20_0.ComputerSystem",
			ID:           id,
			Name:         "Computer System",
		},
		SystemType: "Physical",
		PowerState: "On",
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		Boot: Boot{
			BootSourceOverrideEnabled: "Once",
			BootSourceOverrideTarget:  "None",
		},
		ProcessorSummary: ProcessorSummary{
			Count: 1,
			Status: Status{
				State:  "Enabled",
				Health: "OK",
			},
		},
		MemorySummary: MemorySummary{
			TotalSystemMemoryGiB: 16.0,
			Status: Status{
				State:  "Enabled",
				Health: "OK",
			},
		},
		Processors:  ODataID("/redfish/v1/Systems/" + id + "/Processors"),
		Memory:      ODataID("/redfish/v1/Systems/" + id + "/Memory"),
		LogServices: ODataID("/redfish/v1/Systems/" + id + "/LogServices"),
		Links: ComputerSystemLinks{
			ManagedBy: []ODataID{"/redfish/v1/Managers/1"},
		},
		Actions: ComputerSystemActions{
			ComputerSystemReset: struct {
				Target string `json:"target"`
				Title  string `json:"title,omitempty"`
			}{
				Target: "/redfish/v1/Systems/" + id + "/Actions/ComputerSystem.Reset",
				Title:  "Reset Computer System",
			},
		},
		Oem: &OEM{
			Contoso: NewContosoOEM(),
		},
	}
}

// ComputerSystemCollection represents a collection of computer systems
type ComputerSystemCollection struct {
	Collection
}

// NewComputerSystemCollection creates a new ComputerSystemCollection instance
func NewComputerSystemCollection() *ComputerSystemCollection {
	return &ComputerSystemCollection{
		Collection: Collection{
			ODataContext:      "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
			ODataID:           "/redfish/v1/Systems",
			ODataType:         "#ComputerSystemCollection.ComputerSystemCollection",
			Name:              "Computer System Collection",
			Members:           []ODataID{"/redfish/v1/Systems/1"},
			MembersODataCount: 1,
		},
	}
}
