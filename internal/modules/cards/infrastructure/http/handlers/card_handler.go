package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type CardHandler struct {
	listCards  usecase.ListCards
	getCard    usecase.GetCard
	createCard usecase.CreateCard
	updateCard usecase.UpdateCard
	deactivate usecase.DeactivateCard
}

func NewCardHandler(
	listCards usecase.ListCards,
	getCard usecase.GetCard,
	createCard usecase.CreateCard,
	updateCard usecase.UpdateCard,
	deactivate usecase.DeactivateCard,
) *CardHandler {
	return &CardHandler{
		listCards:  listCards,
		getCard:    getCard,
		createCard: createCard,
		updateCard: updateCard,
		deactivate: deactivate,
	}
}

func (h *CardHandler) List(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	out, err := h.listCards.Execute(c.UserContext(), userID, paginationFromCtx(c))
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CardHandler) Get(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	id, err := vos.ParseCardID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.ErrInvalidCardID.Error()})
	}
	out, err := h.getCard.Execute(c.UserContext(), userID, id)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CardHandler) Create(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	var req dtos.CardRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}
	out, err := h.createCard.Execute(c.UserContext(), userID, req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

func (h *CardHandler) Update(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	id, err := vos.ParseCardID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.ErrInvalidCardID.Error()})
	}
	var req dtos.CardRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}
	out, err := h.updateCard.Execute(c.UserContext(), userID, id, req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CardHandler) Deactivate(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	id, err := vos.ParseCardID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.ErrInvalidCardID.Error()})
	}
	if err := h.deactivate.Execute(c.UserContext(), userID, id); err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// paginationFromCtx extrai page e size dos query params, aplicando defaults
// quando os valores estão ausentes ou são inválidos. NewPagination cuida do
// clamp de tamanho máximo.
func paginationFromCtx(c *fiber.Ctx) dtos.Pagination {
	pageRaw := c.Query("page")
	sizeRaw := c.Query("size")
	if pageRaw == "" && sizeRaw == "" {
		return dtos.Pagination{}
	}

	page := 0
	if pageRaw != "" {
		parsed, err := strconv.Atoi(pageRaw)
		if err != nil {
			page = -1
		}
		if err == nil {
			page = parsed
		}
	}

	size := 0
	if sizeRaw != "" {
		parsed, err := strconv.Atoi(sizeRaw)
		if err != nil {
			size = -1
		}
		if err == nil {
			size = parsed
		}
	}

	return dtos.NewPagination(page, size)
}

func userIDFromCtx(c *fiber.Ctx) (identityvo.UserID, error) {
	identity, err := identitycontext.FromContext(c.UserContext())
	if err != nil {
		return "", err
	}
	return identityvo.ParseUserID(identity.UserID)
}
