package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var (
	errInvalidPage     = errors.New("parâmetro page inválido")
	errInvalidPageSize = errors.New("parâmetro pageSize inválido")
)

type CategoryHandler struct {
	listCategories usecase.ListCategories
	getCategory    usecase.GetCategory
	createCategory usecase.CreateCategory
	updateCategory usecase.UpdateCategory
	deleteCategory usecase.DeleteCategory
}

func NewCategoryHandler(
	listCategories usecase.ListCategories,
	getCategory usecase.GetCategory,
	createCategory usecase.CreateCategory,
	updateCategory usecase.UpdateCategory,
	deleteCategory usecase.DeleteCategory,
) *CategoryHandler {
	return &CategoryHandler{
		listCategories: listCategories,
		getCategory:    getCategory,
		createCategory: createCategory,
		updateCategory: updateCategory,
		deleteCategory: deleteCategory,
	}
}

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	q, err := listQueryFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}
	out, err := h.listCategories.Execute(c.UserContext(), userID, q)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CategoryHandler) Get(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	out, err := h.getCategory.Execute(c.UserContext(), userID, c.Params("id"))
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	var req dtos.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}
	out, err := h.createCategory.Execute(c.UserContext(), userID, req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	var req dtos.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}
	out, err := h.updateCategory.Execute(c.UserContext(), userID, c.Params("id"), req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.deleteCategory.Execute(c.UserContext(), userID, c.Params("id")); err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

func userIDFromCtx(c *fiber.Ctx) (identityvo.UserID, error) {
	identity, err := identitycontext.FromContext(c.UserContext())
	if err != nil {
		return "", err
	}
	return identityvo.ParseUserID(identity.UserID)
}

func listQueryFromCtx(c *fiber.Ctx) (dtos.ListCategoriesQuery, error) {
	q := dtos.ListCategoriesQuery{
		Name:     strings.TrimSpace(c.Query("name")),
		Scope:    strings.TrimSpace(c.Query("scope")),
		ParentID: strings.TrimSpace(c.Query("parentId")),
	}
	if raw := c.Query("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return dtos.ListCategoriesQuery{}, errInvalidPage
		}
		q.Page = v
	}
	if raw := c.Query("pageSize"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return dtos.ListCategoriesQuery{}, errInvalidPageSize
		}
		q.PageSize = v
	}
	return q, nil
}
