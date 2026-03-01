package dto

type StaffCreateRequest struct {
	Username   string `json:"username" validate:"required,min=3"`
	Password   string `json:"password" validate:"required,min=8"`
	HospitalID int    `json:"hospital_id" validate:"required,gt=0"`
}

type StaffCreateResponse struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	HospitalID int    `json:"hospital_id"`
}
