package models

// Chassis represents a physical or virtual chassis
type Chassis struct {
	Resource
	ChassisType        string       `json:"ChassisType"` // Rack, Blade, Enclosure, etc.
	Manufacturer       string       `json:"Manufacturer,omitempty"`
	Model              string       `json:"Model,omitempty"`
	SKU                string       `json:"SKU,omitempty"`
	SerialNumber       string       `json:"SerialNumber,omitempty"`
	PartNumber         string       `json:"PartNumber,omitempty"`
	AssetTag           string       `json:"AssetTag,omitempty"`
	Status             Status       `json:"Status,omitempty"`
	PowerState         string       `json:"PowerState,omitempty"`         // On, Off, PoweringOn, etc.
	EnvironmentalClass string       `json:"EnvironmentalClass,omitempty"` // A1-A4
	HeightMm           float64      `json:"HeightMm,omitempty"`
	WidthMm            float64      `json:"WidthMm,omitempty"`
	DepthMm            float64      `json:"DepthMm,omitempty"`
	WeightKg           float64      `json:"WeightKg,omitempty"`
	Power              ODataID      `json:"Power,omitempty"`
	Thermal            ODataID      `json:"Thermal,omitempty"`
	NetworkAdapters    ODataID      `json:"NetworkAdapters,omitempty"`
	Drives             ODataID      `json:"Drives,omitempty"`
	PCIeDevices        ODataID      `json:"PCIeDevices,omitempty"`
	Links              ChassisLinks `json:"Links,omitempty"`
}

// ChassisLinks represents links to related resources
type ChassisLinks struct {
	ComputerSystems []ODataID `json:"ComputerSystems,omitempty"`
	ContainedBy     ODataID   `json:"ContainedBy,omitempty"`
	Contains        []ODataID `json:"Contains,omitempty"`
	CooledBy        []ODataID `json:"CooledBy,omitempty"`
	ManagedBy       []ODataID `json:"ManagedBy,omitempty"`
	PoweredBy       []ODataID `json:"PoweredBy,omitempty"`
	Oem             Oem       `json:"Oem,omitempty"`
}

// NewChassis creates a new Chassis instance
func NewChassis(id string) *Chassis {
	return &Chassis{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#Chassis.Chassis",
			ODataID:      ODataID("/redfish/v1/Chassis/" + id),
			ODataType:    "#Chassis.v1_23_0.Chassis",
			ID:           id,
			Name:         "Chassis",
		},
		ChassisType: "Rack",
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		PowerState: "On",
		HeightMm:   44.0,  // 1U height
		WidthMm:    482.6, // Standard rack width
		DepthMm:    711.2, // Standard depth
		WeightKg:   15.0,
		Power:      ODataID("/redfish/v1/Chassis/" + id + "/Power"),
		Thermal:    ODataID("/redfish/v1/Chassis/" + id + "/Thermal"),
		Links: ChassisLinks{
			ComputerSystems: []ODataID{ODataID("/redfish/v1/Systems/1")},
			ManagedBy:       []ODataID{ODataID("/redfish/v1/Managers/1")},
		},
	}
}

// ChassisCollection represents a collection of chassis
type ChassisCollection struct {
	Collection
}

// NewChassisCollection creates a new ChassisCollection instance
func NewChassisCollection() *ChassisCollection {
	return &ChassisCollection{
		Collection: Collection{
			ODataContext:      "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
			ODataID:           "/redfish/v1/Chassis",
			ODataType:         "#ChassisCollection.ChassisCollection",
			Name:              "Chassis Collection",
			Members:           []ODataID{"/redfish/v1/Chassis/1"},
			MembersODataCount: 1,
		},
	}
}
