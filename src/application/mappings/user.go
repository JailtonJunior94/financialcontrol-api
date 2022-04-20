package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

func ToUserEntity(r *requests.UserRequest, password string) (e *entities.User) {
	entity := entities.NewUser(r.Name, r.Email, r.Password)
	return entity
}

func ToUserResponse(e *entities.User) (r *responses.UserResponse) {
	return &responses.UserResponse{
		ID:     e.ID,
		Name:   e.Name,
		Email:  e.Email,
		Active: e.Active,
	}
}
