package application

import identitydomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"

func ToUserEntity(request *UserRequest, password string) *identitydomain.User {
	entity := identitydomain.NewUser(request.Name, request.Email, password)
	return entity
}

func ToUserResponse(user *identitydomain.User) *UserResponse {
	return &UserResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Active: user.Active,
	}
}
