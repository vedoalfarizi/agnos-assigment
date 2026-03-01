package model

import "time"

type Staff struct {
	ID         int       `db:"id"`
	HospitalID int       `db:"hospital_id"`
	Username   string    `db:"username"`
	Password   string    `db:"password"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
