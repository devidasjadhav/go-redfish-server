package models

// Common Redfish objects and types used across multiple schemas

// ODataContext represents the @odata.context annotation
type ODataContext string

// ODataID represents the @odata.id annotation
type ODataID string

// Link represents a reference to another resource
type Link struct {
	ODataID ODataID `json:"@odata.id"`
}

// ODataType represents the @odata.type annotation
type ODataType string

// Status represents the health and state of a resource
type Status struct {
	State  string `json:"State,omitempty"`  // Enabled, Disabled, StandbyOffline, etc.
	Health string `json:"Health,omitempty"` // OK, Warning, Critical
}

// Location represents the location of a resource
type Location struct {
	Info       string `json:"Info,omitempty"`
	InfoFormat string `json:"InfoFormat,omitempty"`
}

// Identifier represents additional identifiers for a resource
type Identifier struct {
	DurableName       string `json:"DurableName,omitempty"`
	DurableNameFormat string `json:"DurableNameFormat,omitempty"`
}

// IPv4Address represents an IPv4 address configuration
type IPv4Address struct {
	Address       string `json:"Address,omitempty"`
	SubnetMask    string `json:"SubnetMask,omitempty"`
	Gateway       string `json:"Gateway,omitempty"`
	AddressOrigin string `json:"AddressOrigin,omitempty"` // Static, DHCP, etc.
}

// IPv6Address represents an IPv6 address configuration
type IPv6Address struct {
	Address       string `json:"Address,omitempty"`
	PrefixLength  int    `json:"PrefixLength,omitempty"`
	AddressOrigin string `json:"AddressOrigin,omitempty"` // Static, DHCP, SLAAC, etc.
	AddressState  string `json:"AddressState,omitempty"`  // Preferred, Deprecated, etc.
}

// Actions represents the available actions for a resource
type Actions struct {
	Oem map[string]interface{} `json:"Oem,omitempty"`
}

// Links represents the links to related resources
type Links struct {
	Oem map[string]interface{} `json:"Oem,omitempty"`
}

// Oem represents OEM-specific extensions
type Oem struct {
	// This will be extended with specific OEM implementations
}

// Resource represents the common properties all Redfish resources share
type Resource struct {
	ODataContext ODataContext `json:"@odata.context,omitempty"`
	ODataID      ODataID      `json:"@odata.id,omitempty"`
	ODataType    ODataType    `json:"@odata.type,omitempty"`
	ID           string       `json:"Id"`
	Name         string       `json:"Name"`
	Description  string       `json:"Description,omitempty"`
	Oem          *Oem         `json:"Oem,omitempty"`
}

// Collection represents a collection of resources
type Collection struct {
	ODataContext      ODataContext `json:"@odata.context,omitempty"`
	ODataID           ODataID      `json:"@odata.id,omitempty"`
	ODataType         ODataType    `json:"@odata.type,omitempty"`
	Name              string       `json:"Name"`
	Members           []Link       `json:"Members"`
	MembersODataCount int          `json:"Members@odata.count"`
	Oem               *Oem         `json:"Oem,omitempty"`
}

// Message represents an error message
type Message struct {
	MessageID  string `json:"MessageId"`
	Message    string `json:"Message,omitempty"`
	Severity   string `json:"Severity,omitempty"` // OK, Warning, Critical
	Resolution string `json:"Resolution,omitempty"`
}

// RedfishError represents a Redfish error response
type RedfishError struct {
	Error struct {
		Code    string    `json:"code"`
		Message string    `json:"message"`
		Details []Message `json:"@Message.ExtendedInfo,omitempty"`
	} `json:"error"`
}
