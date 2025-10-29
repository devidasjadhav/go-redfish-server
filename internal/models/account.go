package models

// AccountService represents the account service
type AccountService struct {
	Resource
	ServiceEnabled                  bool    `json:"ServiceEnabled"`
	Accounts                        ODataID `json:"Accounts,omitempty"`
	Roles                           ODataID `json:"Roles,omitempty"`
	PrivilegeMap                    ODataID `json:"PrivilegeMap,omitempty"`
	Status                          Status  `json:"Status,omitempty"`
	MinPasswordLength               int     `json:"MinPasswordLength,omitempty"`
	MaxPasswordLength               int     `json:"MaxPasswordLength,omitempty"`
	AccountLockoutThreshold         int     `json:"AccountLockoutThreshold,omitempty"`
	AccountLockoutDuration          int     `json:"AccountLockoutDuration,omitempty"`
	AccountLockoutCounterResetAfter int     `json:"AccountLockoutCounterResetAfter,omitempty"`
}

// NewAccountService creates a new AccountService instance
func NewAccountService() *AccountService {
	return &AccountService{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#AccountService.AccountService",
			ODataID:      "/redfish/v1/AccountService",
			ODataType:    "#AccountService.v1_15_0.AccountService",
			ID:           "AccountService",
			Name:         "Account Service",
		},
		ServiceEnabled: true,
		Accounts:       "/redfish/v1/AccountService/Accounts",
		Roles:          "/redfish/v1/AccountService/Roles",
		PrivilegeMap:   "/redfish/v1/AccountService/PrivilegeMap",
		Status: Status{
			State:  "Enabled",
			Health: "OK",
		},
		MinPasswordLength:               8,
		MaxPasswordLength:               64,
		AccountLockoutThreshold:         5,
		AccountLockoutDuration:          300,  // 5 minutes
		AccountLockoutCounterResetAfter: 1800, // 30 minutes
	}
}

// ManagerAccount represents a user account
type ManagerAccount struct {
	Resource
	UserName     string       `json:"UserName"`
	Password     string       `json:"Password,omitempty"` // Never returned in responses
	RoleId       string       `json:"RoleId"`
	AccountTypes []string     `json:"AccountTypes,omitempty"` // Redfish, SNMP, etc.
	Enabled      bool         `json:"Enabled"`
	Locked       bool         `json:"Locked,omitempty"`
	Links        AccountLinks `json:"Links,omitempty"`
}

// AccountLinks represents links for an account
type AccountLinks struct {
	Role ODataID `json:"Role,omitempty"`
}

// NewManagerAccount creates a new ManagerAccount instance
func NewManagerAccount(username, roleId string, enabled bool) *ManagerAccount {
	return &ManagerAccount{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#ManagerAccount.ManagerAccount",
			ODataID:      ODataID("/redfish/v1/AccountService/Accounts/" + username),
			ODataType:    "#ManagerAccount.v1_13_0.ManagerAccount",
			ID:           username,
			Name:         "User Account",
		},
		UserName:     username,
		RoleId:       roleId,
		AccountTypes: []string{"Redfish"},
		Enabled:      enabled,
		Locked:       false,
		Links: AccountLinks{
			Role: ODataID("/redfish/v1/AccountService/Roles/" + roleId),
		},
	}
}

// ManagerAccountCollection represents a collection of accounts
type ManagerAccountCollection struct {
	Collection
}

// NewManagerAccountCollection creates a new ManagerAccountCollection instance
func NewManagerAccountCollection() *ManagerAccountCollection {
	return &ManagerAccountCollection{
		Collection: Collection{
			ODataContext:      "/redfish/v1/$metadata#ManagerAccountCollection.ManagerAccountCollection",
			ODataID:           "/redfish/v1/AccountService/Accounts",
			ODataType:         "#ManagerAccountCollection.ManagerAccountCollection",
			Name:              "Accounts Collection",
			Members:           []ODataID{"/redfish/v1/AccountService/Accounts/admin", "/redfish/v1/AccountService/Accounts/operator"},
			MembersODataCount: 2,
		},
	}
}
