package model

import (
	"encoding/json"
)

func init() {
	Register(&Profile{})
}

type Profile struct {
	ProfileID   json.Number `json:"profile_id"   gorm:"column:profile_id;primaryKey"`
	ProfileName string      `json:"profile_name" gorm:"column:profile_name"`
	Email       string      `json:"email"        gorm:"column:email"`
	State       string      `json:"state"        gorm:"column:state"`
}

func (Profile) TableName() string {
	return "dm_core.profile"
}
