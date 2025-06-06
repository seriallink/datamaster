package model

import (
	"github.com/jaswdr/faker"
)

type Profile struct {
	ProfileID string `gorm:"column:profile_id;primaryKey"`
	Email     string `gorm:"column:email"`
	State     string `gorm:"column:state"`
}

func (Profile) TableName() string {
	return "dm_core.profile"
}

func FakeProfile(f faker.Faker) Profile {
	return Profile{
		ProfileID: f.RandomStringWithLength(10),
		Email:     f.Internet().Email(),
		State:     f.Address().State(),
	}
}
