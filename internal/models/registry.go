package models

// MessageRegistry represents a message registry containing message definitions
type MessageRegistry struct {
	Resource
	Language string                     `json:"Language"`
	Messages map[string]RegistryMessage `json:"Messages"`
}

// RegistryMessage represents a single message in a message registry
type RegistryMessage struct {
	Description     string      `json:"Description"`
	Message         string      `json:"Message"`
	NumberOfArgs    int         `json:"NumberOfArgs"`
	Severity        string      `json:"Severity,omitempty"`
	MessageSeverity string      `json:"MessageSeverity,omitempty"`
	Resolution      string      `json:"Resolution"`
	ParamTypes      []string    `json:"ParamTypes,omitempty"`
	ArgDescriptions []string    `json:"ArgDescriptions,omitempty"`
	LongDescription string      `json:"LongDescription,omitempty"`
	Deprecated      string      `json:"Deprecated,omitempty"`
	ClearsAll       bool        `json:"ClearsAll,omitempty"`
	ClearsIf        string      `json:"ClearsIf,omitempty"`
	ClearsMessage   []string    `json:"ClearsMessage,omitempty"`
	Oem             interface{} `json:"Oem,omitempty"`
}

// MessageRegistryFile represents a registry file locator resource
type MessageRegistryFile struct {
	Resource
	Languages []string               `json:"Languages"`
	Registry  string                 `json:"Registry"`
	Location  []RegistryFileLocation `json:"Location"`
}

// RegistryFileLocation represents location information for a registry file
type RegistryFileLocation struct {
	Language       string `json:"Language"`
	Uri            string `json:"Uri,omitempty"`
	ArchiveUri     string `json:"ArchiveUri,omitempty"`
	ArchiveFile    string `json:"ArchiveFile,omitempty"`
	PublicationUri string `json:"PublicationUri,omitempty"`
}

// NewMessageRegistry creates a new MessageRegistry instance
func NewMessageRegistry(language string) *MessageRegistry {
	return &MessageRegistry{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#MessageRegistry.MessageRegistry",
			ODataID:      "/redfish/v1/Registries/Base.1.0.0",
			ODataType:    "#MessageRegistry.v1_7_0.MessageRegistry",
			ID:           "Base.1.0.0",
			Name:         "Base Message Registry",
		},
		Language: language,
		Messages: map[string]RegistryMessage{
			"Success": {
				Description:     "Indicates a successful operation",
				Message:         "Successfully Completed Request",
				NumberOfArgs:    0,
				MessageSeverity: "OK",
				Severity:        "OK",
				Resolution:      "No action required",
			},
			"InternalError": {
				Description:     "Indicates an internal error",
				Message:         "Internal Server Error",
				NumberOfArgs:    0,
				MessageSeverity: "Critical",
				Severity:        "Critical",
				Resolution:      "Contact system administrator",
			},
			"ResourceNotFound": {
				Description:     "The requested resource was not found",
				Message:         "The requested resource %1 was not found",
				NumberOfArgs:    1,
				MessageSeverity: "Warning",
				Severity:        "Warning",
				Resolution:      "Check the URI and try again",
				ParamTypes:      []string{"string"},
				ArgDescriptions: []string{"URI of the resource"},
			},
			"PropertyValueNotInList": {
				Description:     "The property value is not in the list of acceptable values",
				Message:         "The value %1 for the property %2 is not in the list of acceptable values",
				NumberOfArgs:    2,
				MessageSeverity: "Warning",
				Severity:        "Warning",
				Resolution:      "Choose a value from the enumeration list",
				ParamTypes:      []string{"string", "string"},
				ArgDescriptions: []string{"Property value", "Property name"},
			},
		},
	}
}

// NewMessageRegistryFile creates a new MessageRegistryFile instance
func NewMessageRegistryFile(id string, registry string) *MessageRegistryFile {
	return &MessageRegistryFile{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#MessageRegistryFile.MessageRegistryFile",
			ODataID:      ODataID("/redfish/v1/Registries/" + id),
			ODataType:    "#MessageRegistryFile.v1_1_5.MessageRegistryFile",
			ID:           id,
			Name:         registry + " Message Registry File",
			Description:  registry + " Message Registry File locations",
		},
		Languages: []string{"en"},
		Registry:  registry,
		Location: []RegistryFileLocation{
			{
				Language:       "en",
				Uri:            "/redfish/v1/Registries/" + id + ".json",
				PublicationUri: "https://www.dmtf.org/sites/default/files/standards/documents/DSP8011_" + registry + ".json",
			},
		},
	}
}

// OEM represents OEM-specific extensions
type OEM struct {
	Contoso *ContosoOEM `json:"Contoso,omitempty"`
}

// ContosoOEM represents Contoso-specific OEM extensions
type ContosoOEM struct {
	VendorID         string                 `json:"VendorId,omitempty"`
	ProductID        string                 `json:"ProductId,omitempty"`
	SerialNumber     string                 `json:"SerialNumber,omitempty"`
	FirmwareVersion  string                 `json:"FirmwareVersion,omitempty"`
	CustomProperties map[string]interface{} `json:"CustomProperties,omitempty"`
}

// NewContosoOEM creates a new Contoso OEM extension
func NewContosoOEM() *ContosoOEM {
	return &ContosoOEM{
		VendorID:        "CONTOSO",
		ProductID:       "SERVER-001",
		SerialNumber:    "CN123456789",
		FirmwareVersion: "1.2.3",
		CustomProperties: map[string]interface{}{
			"PowerEfficiency":      95.5,
			"TemperatureThreshold": 75,
			"CustomFeatureEnabled": true,
		},
	}
}
