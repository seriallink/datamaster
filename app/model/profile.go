package model

import (
	"encoding/json"
)

func init() {
	Register(&Profile{})
}

// Profile represents a user profile record from the "profile" table in the raw dataset.
//
// This struct is used to capture metadata about users who submit reviews,
// including identifiers, location, and contact information.
//
// Fields:
//   - ProfileID: Unique identifier of the user (primary key).
//   - ProfileName: Display name of the user.
//   - Email: Email address associated with the profile.
//   - State: Geographic state or region of the user.
type Profile struct {
	ProfileID   json.Number `json:"profile_id"   gorm:"column:profile_id;primaryKey"`
	ProfileName string      `json:"profile_name" gorm:"column:profile_name"`
	Email       string      `json:"email"        gorm:"column:email"`
	State       string      `json:"state"        gorm:"column:state"`
}

func (Profile) TableName() string {
	return "dm_core.profile"
}
