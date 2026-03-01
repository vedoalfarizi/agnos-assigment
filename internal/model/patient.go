package model

import "time"

type Patient struct {
	ID           int        `db:"id"`
	HospitalID   int        `db:"hospital_id"`
	FirstNameTh  *string    `db:"first_name_th"`
	MiddleNameTh *string    `db:"middle_name_th"`
	LastNameTh   *string    `db:"last_name_th"`
	FirstNameEn  *string    `db:"first_name_en"`
	MiddleNameEn *string    `db:"middle_name_en"`
	LastNameEn   *string    `db:"last_name_en"`
	NationalID   *string    `db:"national_id"`
	PassportID   *string    `db:"passport_id"`
	DateOfBirth  *time.Time `db:"date_of_birth"`
	PhoneNumber  *string    `db:"phone_number"`
	Email        *string    `db:"email"`
	Gender       *string    `db:"gender"`
	HospitalName *string    `db:"hospital_name"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}
