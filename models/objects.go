package models

// Account status values stored in User.Status. Only UserStatusActive may log in.
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

// RoleAdmin is the role seeded by seeders.SeedRoles.
const RoleAdmin = "admin"

func IsValidUserStatus(status string) bool {
	return status == UserStatusActive || status == UserStatusInactive
}

type PocketBaseAuthResponse struct {
	Token  string                 `json:"token"`
	Record map[string]interface{} `json:"record"`
}

type GetRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

func (p *GetRequest) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}

	if p.PageSize == 0 {
		p.PageSize = 10
	}
}

type SearchRequest struct {
	GetRequest
	SearchQuery string `json:"searchQuery,omitempty"`
}

type HeartBeatRequest struct {
	DeviceID  string `json:"device_id"`
	Timestamp string `json:"timestamp"`
}
