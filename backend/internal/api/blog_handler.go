package api

import (
	"testbox/internal/service"

	"github.com/gofiber/fiber/v2"
)

type BlogHandler struct {
	service service.BlogService
}

func NewBlogHandler(service service.BlogService) *BlogHandler {
	return &BlogHandler{service: service}
}

type CreateBlogRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
}

type UpdateBlogRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
}

// CreateBlog creates a new blog post
// @Summary Create a new blog post
// @Description Creates a new blog post with title, content, and tags
// @Tags blogs
// @Accept json
// @Produce json
// @Success 201 {object} models.BlogPost
// @Router /api/blogs [post]
func (h *BlogHandler) CreateBlog(c *fiber.Ctx) error {
	var req CreateBlogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	blog, err := h.service.CreateBlogPost(req.Title, req.Content, req.Tags)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(blog)
}

// GetBlog retrieves a blog post by ID
// @Summary Get blog post by ID
// @Description Retrieves a specific blog post by its ID
// @Tags blogs
// @Produce json
// @Param id path string true "Blog Post ID"
// @Success 200 {object} models.BlogPost
// @Router /api/blogs/{id} [get]
func (h *BlogHandler) GetBlog(c *fiber.Ctx) error {
	id := c.Params("id")

	blog, err := h.service.GetBlogPost(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(blog)
}

// GetAllBlogs retrieves all blog posts
// @Summary Get all blog posts
// @Description Retrieves all blog posts ordered by creation date (newest first)
// @Tags blogs
// @Produce json
// @Success 200 {array} models.BlogPost
// @Router /api/blogs [get]
func (h *BlogHandler) GetAllBlogs(c *fiber.Ctx) error {
	blogs, err := h.service.GetAllBlogPosts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(blogs)
}

// UpdateBlog updates an existing blog post
// @Summary Update blog post
// @Description Updates title, content, and tags of a blog post
// @Tags blogs
// @Accept json
// @Produce json
// @Param id path string true "Blog Post ID"
// @Success 200 {object} models.BlogPost
// @Router /api/blogs/{id} [put]
func (h *BlogHandler) UpdateBlog(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateBlogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	blog, err := h.service.UpdateBlogPost(id, req.Title, req.Content, req.Tags)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(blog)
}

// DeleteBlog deletes a blog post
// @Summary Delete blog post
// @Description Deletes a blog post by ID
// @Tags blogs
// @Param id path string true "Blog Post ID"
// @Success 204
// @Router /api/blogs/{id} [delete]
func (h *BlogHandler) DeleteBlog(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.DeleteBlogPost(id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
