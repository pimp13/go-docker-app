package dto

type UserCreateRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name"  validate:"required,min=2,max=100"`
	Email     string `json:"email"      validate:"required,email"`
	Username  string `json:"username"   validate:"omitempty,min=3,max=50,alphanum"`
	Password  string `json:"password"   validate:"required,min=8"`
	Phone     string `json:"phone"      validate:"omitempty,e164"`
}

type UserCreateResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}
