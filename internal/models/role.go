package models

// Role represents a user role
type Role struct {
	Resource
	RoleId             string   `json:"RoleId"`
	AssignedPrivileges []string `json:"AssignedPrivileges"`
	IsPredefined       bool     `json:"IsPredefined"`
	OemPrivileges      []string `json:"OemPrivileges,omitempty"`
}

// NewRole creates a new Role instance
func NewRole(id, name string, privileges []string, predefined bool) *Role {
	return &Role{
		Resource: Resource{
			ODataContext: "/redfish/v1/$metadata#Role.Role",
			ODataID:      ODataID("/redfish/v1/AccountService/Roles/" + id),
			ODataType:    "#Role.v1_2_0.Role",
			ID:           id,
			Name:         name,
		},
		RoleId:             id,
		AssignedPrivileges: privileges,
		IsPredefined:       predefined,
	}
}

// RoleCollection represents a collection of roles
type RoleCollection struct {
	Collection
}

// NewRoleCollection creates a new RoleCollection instance
func NewRoleCollection() *RoleCollection {
	return &RoleCollection{
		Collection: Collection{
			ODataContext: "/redfish/v1/$metadata#RoleCollection.RoleCollection",
			ODataID:      "/redfish/v1/AccountService/Roles",
			ODataType:    "#RoleCollection.RoleCollection",
			Name:         "Role Collection",
			Members: []Link{
				Link{ODataID: "/redfish/v1/AccountService/Roles/Administrator"},
				Link{ODataID: "/redfish/v1/AccountService/Roles/Operator"},
				Link{ODataID: "/redfish/v1/AccountService/Roles/ReadOnly"},
			},
			MembersODataCount: 3,
		},
	}
}
