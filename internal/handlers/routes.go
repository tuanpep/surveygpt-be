package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/surveyflow/be/internal/config"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/middleware"
	"github.com/surveyflow/be/internal/repository"
	"github.com/surveyflow/be/internal/services"
)

// Dependencies holds references to all services and repositories.
type Dependencies struct {
	Config *config.Config
	Pool   *pgxpool.Pool
	Redis  *redis.Client

	// Handlers (set during SetupRoutes).
	Auth         *AuthHandler
	Org          *OrgHandler
	Survey       *SurveyHandler
	Template     *TemplateHandler
	Response     *ResponseHandler
	Analytics    *AnalyticsHandler
	Distribution *DistributionHandler
	Billing      *BillingHandler
	SSE          *SSEHandler
}

// SetupRoutes creates a new Echo instance, registers global middleware,
// and mounts all route groups.
func SetupRoutes(deps *Dependencies) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// --- Global middleware (order matters) ---
	e.Use(middleware.Recovery())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS(deps.Config.FrontendURL))
	e.Use(middleware.Logger())
	e.Use(echomiddleware.BodyLimit("10M"))
	e.Use(middleware.RateLimit(deps.Redis, "global", 100, time.Minute))

	// Health check (unauthenticated).
	healthHandler := NewHealthHandler(deps.Pool, deps.Redis)
	e.GET("/health", healthHandler.Check)
	e.GET("/healthz", healthHandler.Check)
	e.GET("/ready", healthHandler.Ready)

	// Swagger UI (unauthenticated, only in development).
	if deps.Config.App.Env == "development" {
		e.GET("/docs/openapi.yaml", docsYAMLHandler)
		e.GET("/docs", docsHandler)
		e.GET("/docs/*", docsHandler)
	}

	// --- Initialize repos and services ---
	userRepo := repository.NewUserRepo(deps.Pool)
	orgRepo := repository.NewOrgRepo(deps.Pool)
	surveyRepo := repository.NewSurveyRepo(deps.Pool)
	templateRepo := repository.NewTemplateRepo(deps.Pool)
	responseRepo := repository.NewResponseRepo(deps.Pool)
	analyticsRepo := repository.NewAnalyticsRepo(deps.Pool)
	emailListRepo := repository.NewEmailListRepo(deps.Pool)

	authSvc := services.NewAuthService(userRepo, orgRepo, deps.Pool, deps.Config)
	orgSvc := services.NewOrgService(orgRepo, userRepo)
	surveySvc := services.NewSurveyService(surveyRepo, templateRepo, deps.Config)
	responseSvc := services.NewResponseService(responseRepo, surveyRepo, deps.Pool)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo, surveyRepo)
	distributionSvc := services.NewDistributionService(surveyRepo, emailListRepo)
	billingSvc := services.NewBillingService(orgRepo)

	// --- Initialize handlers ---
	deps.Auth = NewAuthHandler(authSvc)
	deps.Org = NewOrgHandler(orgSvc)
	deps.Survey = NewSurveyHandler(surveySvc)
	deps.Template = NewTemplateHandler(surveySvc)
	deps.Response = NewResponseHandler(responseSvc)
	deps.Analytics = NewAnalyticsHandler(analyticsSvc)
	deps.Distribution = NewDistributionHandler(distributionSvc)
	deps.Billing = NewBillingHandler(billingSvc)
	deps.SSE = NewSSEHandler(deps.Redis, surveyRepo)

	// --- Public routes ---
	public := e.Group("/api/v1")
	{
		// Auth
		public.POST("/auth/register", deps.Auth.Register)
		public.POST("/auth/login", deps.Auth.Login)
		public.POST("/auth/login/2fa", deps.Auth.Login2FA)
		public.POST("/auth/refresh", deps.Auth.Refresh)
		public.POST("/auth/forgot-password", deps.Auth.ForgotPassword)
		public.POST("/auth/reset-password", deps.Auth.ResetPassword)

		// OAuth (stubs for V1.0)
		public.GET("/auth/google", stubHandler("auth:google"))
		public.GET("/auth/google/callback", stubHandler("auth:google:callback"))
		public.GET("/auth/microsoft", stubHandler("auth:microsoft"))
		public.GET("/auth/microsoft/callback", stubHandler("auth:microsoft:callback"))

		// Public survey endpoints
		public.GET("/public/surveys/:slug", deps.Survey.GetPublic)
		public.POST("/surveys/:id/responses", deps.Response.Submit)
		public.PATCH("/surveys/:id/responses/:responseId", stubHandler("responses:update"))
	}

	// --- Authenticated routes ---
	auth := public.Group("")
	auth.Use(middleware.Auth(deps.Config.JWT.Secret))
	{
		// Current user
		auth.GET("/me", deps.Auth.GetMe)
		auth.PUT("/me", deps.Auth.UpdateMe)
		auth.PUT("/me/password", deps.Auth.ChangePassword)
		auth.POST("/me/avatar", stubHandler("users:upload-avatar"))
		auth.DELETE("/me/account", stubHandler("users:delete-account"))

		// Two-factor auth (stubs)
		auth.POST("/me/2fa/enable", stubHandler("users:enable-2fa"))
		auth.POST("/me/2fa/disable", stubHandler("users:disable-2fa"))
		auth.POST("/me/2fa/verify", stubHandler("users:verify-2fa"))

		// Organizations
		auth.POST("/organizations", deps.Org.Create)
		auth.GET("/organizations/me", deps.Org.Get)
		auth.PUT("/organizations/me", deps.Org.Update)
		auth.GET("/organizations/:orgId", deps.Org.Get) // also works for any org the user has access to
		auth.PUT("/organizations/:orgId", deps.Org.Update)
		auth.DELETE("/organizations/:orgId", stubHandler("organizations:delete"))

		// Organization members
		auth.GET("/organizations/:orgId/members", deps.Org.ListMembers)
		auth.POST("/organizations/:orgId/members/invite", deps.Org.InviteMember)
		auth.PATCH("/organizations/:orgId/members/:memberId/role", deps.Org.UpdateMemberRole)
		auth.DELETE("/organizations/:orgId/members/:memberId", deps.Org.RemoveMember)

		// Organization invitations
		auth.GET("/organizations/:orgId/invitations", deps.Org.ListInvitations)
		auth.POST("/invitations/:token/accept", deps.Org.AcceptInvitation)
		auth.POST("/invitations/:token/decline", deps.Org.DeclineInvitation)

		// Surveys
		auth.POST("/surveys", deps.Survey.Create)
		auth.GET("/surveys", deps.Survey.List)
		auth.GET("/surveys/:id", deps.Survey.Get)
		auth.PUT("/surveys/:id", deps.Survey.Update)
		auth.DELETE("/surveys/:id", deps.Survey.Delete)
		auth.POST("/surveys/:id/publish", deps.Survey.Publish)
		auth.POST("/surveys/:id/close", deps.Survey.Close)
		auth.POST("/surveys/:id/duplicate", deps.Survey.Duplicate)
		auth.GET("/surveys/:id/qr-code", deps.Distribution.GetQRCode)
		auth.GET("/surveys/:id/embed-code", deps.Distribution.GetEmbedCode)
		auth.GET("/surveys/:id/live", deps.SSE.LiveUpdates)

		// Responses
		auth.GET("/surveys/:id/responses", deps.Response.List)
		auth.GET("/surveys/:id/responses/:responseId", deps.Response.Get)
		auth.DELETE("/surveys/:id/responses/:responseId", deps.Response.Delete)
		auth.POST("/surveys/:id/responses/export", deps.Response.Export)

		// Analytics
		auth.GET("/surveys/:id/analytics/summary", deps.Analytics.GetSummary)
		auth.GET("/surveys/:id/analytics/questions/:questionId", deps.Analytics.GetQuestionStats)
		auth.GET("/surveys/:id/analytics/cross-tab", deps.Analytics.GetCrossTab)
		auth.GET("/surveys/:id/analytics/dropoff", deps.Analytics.GetDropoff)
		auth.POST("/surveys/:id/analytics/ai-insights", deps.Analytics.GetAIInsights)

		// Templates
		auth.GET("/templates", deps.Template.List)
		auth.POST("/templates", deps.Template.Create)
		auth.GET("/templates/:id", deps.Template.Get)
		auth.POST("/templates/:id/duplicate", deps.Template.Duplicate)
		auth.POST("/surveys/from-template/:id", deps.Template.CreateFromTemplate)

		// Files (stubs)
		auth.POST("/files/upload", stubHandler("files:upload"))
		auth.GET("/files", stubHandler("files:list"))
		auth.DELETE("/files/:id", stubHandler("files:delete"))

		// Integrations (stubs)
		auth.GET("/integrations", stubHandler("integrations:list"))
		auth.POST("/integrations", stubHandler("integrations:create"))
		auth.PUT("/integrations/:id", stubHandler("integrations:update"))
		auth.DELETE("/integrations/:id", stubHandler("integrations:delete"))

		// API keys (stubs)
		auth.GET("/api-keys", stubHandler("api-keys:list"))
		auth.POST("/api-keys", stubHandler("api-keys:create"))
		auth.DELETE("/api-keys/:id", stubHandler("api-keys:revoke"))

		// Email lists
		auth.GET("/email-lists", deps.Distribution.ListEmailLists)
		auth.POST("/email-lists", deps.Distribution.CreateEmailList)
		auth.GET("/email-lists/:id", deps.Distribution.GetEmailList)
		auth.PUT("/email-lists/:id", deps.Distribution.UpdateEmailList)
		auth.DELETE("/email-lists/:id", deps.Distribution.DeleteEmailList)
		auth.POST("/email-lists/:id/contacts", deps.Distribution.AddContacts)
		auth.DELETE("/email-lists/:id/contacts/:contactId", deps.Distribution.RemoveContact)

		// Email distribution
		auth.POST("/surveys/:id/send-emails", deps.Distribution.SendSurveyEmails)
		auth.GET("/surveys/:id/email/campaigns", deps.Distribution.ListEmailCampaigns)

		// Webhooks (stubs)
		auth.GET("/webhooks", stubHandler("webhooks:list"))
		auth.POST("/webhooks", stubHandler("webhooks:create"))
		auth.PUT("/webhooks/:id", stubHandler("webhooks:update"))
		auth.DELETE("/webhooks/:id", stubHandler("webhooks:delete"))

		// Billing
		auth.GET("/billing/plan", deps.Billing.GetPlan)
		auth.GET("/billing/plans", deps.Billing.GetPlans)
		auth.POST("/billing/change-plan", deps.Billing.ChangePlan)
		auth.GET("/billing/portal", deps.Billing.GetPortalURL)
		auth.GET("/billing/history", deps.Billing.GetHistory)

		// Audit logs (stubs)
		auth.GET("/audit-logs", stubHandler("audit-logs:list"))

		// Usage
		auth.GET("/usage", deps.Org.GetUsage)
	}

	// --- Webhook routes (no auth, verified via signature) ---
	webhooks := e.Group("/webhooks")
	{
		webhooks.POST("/stripe", deps.Billing.HandleStripeWebhook)
	}

	// --- Global error handler ---
	e.HTTPErrorHandler = customHTTPErrorHandler

	return e
}

// customHTTPErrorHandler maps AppError types to structured JSON responses.
func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if appErr, ok := err.(*apperr.AppError); ok {
		c.JSON(appErr.StatusCode, echo.Map{
			"code":    appErr.Code,
			"message": appErr.Message,
		})
		return
	}

	// Handle Echo HTTP errors.
	if he, ok := err.(*echo.HTTPError); ok {
		c.JSON(he.Code, echo.Map{
			"code":    httpStatusToCode(he.Code),
			"message": he.Message,
		})
		return
	}

	// Fallback for unexpected errors.
	slog.Error("internal error", "error", err, "path", c.Path())
	c.JSON(http.StatusInternalServerError, echo.Map{
		"code":    "INTERNAL_ERROR",
		"message": "an internal error occurred",
	})
}

// httpStatusToCode converts an HTTP status code to a machine-readable error code.
func httpStatusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// stubHandler returns a placeholder handler that responds with "not implemented".
func stubHandler(endpoint string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, echo.Map{
			"code":    "NOT_IMPLEMENTED",
			"message": "endpoint not yet implemented",
			"endpoint": endpoint,
		})
	}
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>SurveyFlow API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/docs/openapi.yaml",
      dom_id: "#swagger-ui",
      layout: "BaseLayout",
    });
  </script>
</body>
</html>`

// docsHandler serves the Swagger UI HTML page.
func docsHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return c.String(http.StatusOK, swaggerHTML)
}

// docsYAMLHandler serves the OpenAPI specification file.
func docsYAMLHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/yaml")
	return c.File("api/openapi.yaml")
}
