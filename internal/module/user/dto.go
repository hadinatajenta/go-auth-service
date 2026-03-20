package user

type UserProfileResponse struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Roles     []string `json:"roles,omitempty"`
}

type UserCreateRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required,max=50"`
	LastName  string `json:"last_name" binding:"required,max=50"`
}

type UserUpdateRequest struct {
	FirstName string `json:"first_name" binding:"max=50"`
	LastName  string `json:"last_name" binding:"max=50"`
}

type UserResponse struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Roles     []string `json:"roles"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserRoleRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
}

type RoleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
