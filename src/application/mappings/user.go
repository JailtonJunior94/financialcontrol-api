package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

func ToUserEntity(r *requests.UserRequest, password string) (e *entities.User) {
	entity := new(entities.User)
	entity.NewUser(r.Name, r.Email, password)

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
