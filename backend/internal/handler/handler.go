package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"employee-portal/backend/internal/config"
	"employee-portal/backend/internal/domain"
	"employee-portal/backend/internal/middleware"
	"employee-portal/backend/internal/response"
	"employee-portal/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg            config.Config
	authService    *service.AuthService
	profileService *service.ProfileService
	adminService   *service.AdminService
	logger         *slog.Logger
}

func New(cfg config.Config, auth *service.AuthService, profile *service.ProfileService, admin *service.AdminService, logger *slog.Logger) *Handler {
	return &Handler{cfg: cfg, authService: auth, profileService: profile, adminService: admin, logger: logger}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})
	router.GET("/openapi.json", h.openapi)
	router.GET("/swagger", h.swagger)
	router.Static("/uploads", h.cfg.UploadDir)

	api := router.Group("/api/v1")
	api.POST("/auth/login", h.login)
	api.GET("/master-data/skills", h.listSkills)

	protected := api.Group("")
	protected.Use(middleware.Auth(h.authService, h.logger))
	protected.GET("/profile", h.getCurrentProfile)
	protected.PATCH("/profile/contact", h.updateContact)
	protected.PATCH("/profile/skills", h.updateSkills)
	protected.POST("/profile/avatar", h.updateAvatar)
	protected.GET("/profile/change-logs", h.history)
	protected.POST("/demo/profile/reset", h.resetDemo)

	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRole("admin", h.logger))
	admin.GET("/employees", h.listEmployees)
	admin.POST("/employees", h.createEmployee)
	admin.PATCH("/employees/:id/status", h.setEmployeeStatus)
}

func (h *Handler) login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, h.logger, validationError("Invalid login payload"))
		return
	}
	result, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) getCurrentProfile(c *gin.Context) {
	profile, err := h.profileService.GetMe(middleware.UserID(c))
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, profile)
}

func (h *Handler) listSkills(c *gin.Context) {
	skills, err := h.profileService.ListSkills()
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, skills)
}

func (h *Handler) updateContact(c *gin.Context) {
	var req struct {
		MobilePhone  string `json:"mobilePhone"`
		ContactEmail string `json:"contactEmail"`
		Address      string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, h.logger, validationError("Invalid contact payload"))
		return
	}
	profile, err := h.profileService.UpdateContact(middleware.UserID(c), req.MobilePhone, req.ContactEmail, req.Address)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, profile)
}

func (h *Handler) updateSkills(c *gin.Context) {
	var req struct {
		SkillIDs []uint `json:"skillIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, h.logger, validationError("Invalid skills payload"))
		return
	}
	profile, err := h.profileService.UpdateSkills(middleware.UserID(c), req.SkillIDs)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, profile)
}

func (h *Handler) updateAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		response.Fail(c, h.logger, validationError("Avatar file is required"))
		return
	}
	if err := h.profileService.ValidateAvatar(file); err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	filename := fmt.Sprintf("avatar-%d-%d%s", middleware.UserID(c), time.Now().UnixNano(), filepath.Ext(file.Filename))
	path := filepath.Join(h.cfg.UploadDir, filename)
	if err := c.SaveUploadedFile(file, path); err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	profile, err := h.profileService.UpdateAvatar(middleware.UserID(c), "/uploads/"+filename)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.Created(c, profile)
}

func (h *Handler) history(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", strconv.Itoa(h.cfg.HistoryRetentionDays)))
	logs, err := h.profileService.ListHistory(middleware.UserID(c), days)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, logs)
}

func (h *Handler) resetDemo(c *gin.Context) {
	profile, err := h.profileService.ResetDemo(middleware.UserID(c))
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, profile)
}

func (h *Handler) listEmployees(c *gin.Context) {
	employees, err := h.adminService.ListEmployees()
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, employees)
}

func (h *Handler) createEmployee(c *gin.Context) {
	var req service.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, h.logger, validationError("Invalid employee payload"))
		return
	}
	employee, err := h.adminService.CreateEmployee(req)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.Created(c, employee)
}

func (h *Handler) setEmployeeStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.Fail(c, h.logger, validationError("Invalid employee id"))
		return
	}
	var req struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, h.logger, validationError("Invalid status payload"))
		return
	}
	employee, err := h.adminService.SetEmployeeActive(uint(id), req.IsActive)
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	response.OK(c, employee)
}

func (h *Handler) openapi(c *gin.Context) {
	c.File("docs/openapi.json")
}

func (h *Handler) swagger(c *gin.Context) {
	html, err := os.ReadFile("docs/swagger.html")
	if err != nil {
		response.Fail(c, h.logger, err)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", html)
}

func BuildRouter(cfg config.Config, h *Handler, logger *slog.Logger) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery(), middleware.RequestLogger(logger))
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if isAllowedOrigin(origin, cfg.CORSAllowedOrigin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	h.RegisterRoutes(router)
	return router
}

func validationError(message string) error {
	return domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", message)
}

func isAllowedOrigin(origin, allowed string) bool {
	if origin == "" {
		return false
	}
	for _, item := range strings.Split(allowed, ",") {
		if strings.TrimSpace(item) == origin {
			return true
		}
	}
	return false
}
