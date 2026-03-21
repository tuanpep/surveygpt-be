package services

import (
	"context"
	"time"

	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/repository"
)

// BillingService handles billing and subscription management.
type BillingService struct {
	orgRepo repository.OrgRepo
}

// NewBillingService creates a new BillingService.
func NewBillingService(orgRepo repository.OrgRepo) *BillingService {
	return &BillingService{
		orgRepo: orgRepo,
	}
}

// --- Types ---

type PlanInfo struct {
	PlanID          string    `json:"plan_id"`
	PlanName        string    `json:"plan_name"`
	BillingPeriod   string    `json:"billing_period"` // monthly, yearly
	Status          string    `json:"status"`          // active, canceled, past_due
	TrialEndsAt     *time.Time `json:"trial_ends_at,omitempty"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty"`
	Limits          PlanLimits `json:"limits"`
	Usage           PlanUsage  `json:"usage"`
}

type PlanLimits struct {
	Surveys       int `json:"surveys"`
	Responses     int `json:"responses"`
	Members       int `json:"members"`
	EmailCampaigns int `json:"email_campaigns"`
	AICredits     int `json:"ai_credits"`
}

type PlanUsage struct {
	Surveys       int `json:"surveys"`
	Responses     int `json:"responses"`
	Members       int `json:"members"`
	EmailCampaigns int `json:"email_campaigns"`
	AICredits     int `json:"ai_credits"`
}

type Plan struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	PriceMonthly  int        `json:"price_monthly"`
	PriceYearly   int        `json:"price_yearly"`
	Features      []string   `json:"features"`
	Limits        PlanLimits `json:"limits"`
	IsPopular     bool       `json:"is_popular"`
}

type Invoice struct {
	ID          string    `json:"id"`
	Number      string    `json:"number"`
	Amount      int       `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`    // paid, open, void
	IssuedAt    time.Time `json:"issued_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	PDFURL      string    `json:"pdf_url,omitempty"`
}

type ChangePlanInput struct {
	PlanID         string `json:"plan_id" validate:"required"`
	BillingPeriod  string `json:"billing_period"` // monthly, yearly
	SuccessURL     string `json:"success_url"`
	CancelURL      string `json:"cancel_url"`
}

type ChangePlanOutput struct {
	CheckoutURL string `json:"checkout_url"`
}

// --- Methods ---

// GetPlan returns the current plan info for an organization.
func (s *BillingService) GetPlan(ctx context.Context, orgID string) (*PlanInfo, error) {
	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if org == nil {
		return nil, apperr.NotFound("organization")
	}

	// Get usage.
	usage, err := s.orgRepo.GetUsage(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Map plan to limits.
	limits := s.getLimitsForPlan(org.Plan)

	// Build plan info.
	planInfo := &PlanInfo{
		PlanID:        org.Plan,
		PlanName:      s.getPlanName(org.Plan),
		BillingPeriod: "monthly", // Default for MVP
		Status:        "active",
		Limits:        limits,
		Usage: PlanUsage{
			Surveys:        usage.Surveys,
			Responses:      usage.Responses,
			Members:        usage.Members,
			EmailCampaigns: 0, // Track separately
			AICredits:      usage.AICredits.Used,
		},
	}

	// Calculate next billing date.
	if org.CurrentPeriodEnd != nil {
		planInfo.NextBillingDate = org.CurrentPeriodEnd
	}

	return planInfo, nil
}

// GetPlans returns all available plans.
func (s *BillingService) GetPlans(ctx context.Context) []Plan {
	return []Plan{
		{
			ID:          "free",
			Name:        "Free",
			Description: "Perfect for trying out SurveyFlow",
			PriceMonthly: 0,
			PriceYearly:  0,
			Features: []string{
				"3 surveys",
				"100 responses per month",
				"1 team member",
				"Basic analytics",
				"Email support",
			},
			Limits: PlanLimits{
				Surveys:        3,
				Responses:      100,
				Members:        1,
				EmailCampaigns: 0,
				AICredits:      0,
			},
		},
		{
			ID:          "starter",
			Name:        "Starter",
			Description: "For small teams and growing businesses",
			PriceMonthly: 2900,
			PriceYearly:  29000,
			Features: []string{
				"Unlimited surveys",
				"1,000 responses per month",
				"5 team members",
				"Advanced analytics",
				"Email distribution",
				"Custom branding",
				"Priority support",
			},
			Limits: PlanLimits{
				Surveys:        -1, // unlimited
				Responses:      1000,
				Members:        5,
				EmailCampaigns: 10,
				AICredits:      0,
			},
			IsPopular: true,
		},
		{
			ID:          "professional",
			Name:        "Professional",
			Description: "For organizations that need more power",
			PriceMonthly: 9900,
			PriceYearly:  99000,
			Features: []string{
				"Everything in Starter",
				"10,000 responses per month",
				"20 team members",
				"AI-powered insights",
				"Custom themes",
				"API access",
				"Webhooks",
				"SSO",
			},
			Limits: PlanLimits{
				Surveys:        -1,
				Responses:      10000,
				Members:        20,
				EmailCampaigns: -1,
				AICredits:      100,
			},
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			Description: "For large organizations with custom needs",
			PriceMonthly: -1, // contact sales
			PriceYearly:  -1,
			Features: []string{
				"Everything in Professional",
				"Unlimited responses",
				"Unlimited team members",
				"Unlimited AI credits",
				"Custom integrations",
				"Dedicated account manager",
				"SLA guarantee",
				"On-premise option",
			},
			Limits: PlanLimits{
				Surveys:        -1,
				Responses:      -1,
				Members:        -1,
				EmailCampaigns: -1,
				AICredits:      -1,
			},
		},
	}
}

// ChangePlan initiates a plan change.
func (s *BillingService) ChangePlan(ctx context.Context, orgID string, input ChangePlanInput) (*ChangePlanOutput, error) {
	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if org == nil {
		return nil, apperr.NotFound("organization")
	}

	// For MVP, return placeholder.
	// In production, would create Stripe checkout session.
	checkoutURL := "https://checkout.stripe.com/pay-placeholder"

	return &ChangePlanOutput{
		CheckoutURL: checkoutURL,
	}, nil
}

// GetPortalURL returns the Stripe customer portal URL.
func (s *BillingService) GetPortalURL(ctx context.Context, orgID string) (string, error) {
	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return "", apperr.Internal(err)
	}
	if org == nil {
		return "", apperr.NotFound("organization")
	}

	// For MVP, return placeholder.
	// In production, would create Stripe portal session.
	return "https://portal.stripe.com/placeholder", nil
}

// GetHistory returns invoice history.
func (s *BillingService) GetHistory(ctx context.Context, orgID string) ([]Invoice, error) {
	// For MVP, return empty list.
	// In production, would fetch from Stripe.
	return []Invoice{}, nil
}

// HandleStripeWebhook handles incoming Stripe webhooks.
func (s *BillingService) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	// For MVP, placeholder.
	// In production:
	// 1. Verify signature with Stripe webhook secret
	// 2. Parse event
	// 3. Switch on event type:
	//    - checkout.session.completed: Update org plan
	//    - customer.subscription.updated: Update plan/limits
	//    - invoice.paid: Record payment
	//    - customer.subscription.deleted: Cancel plan

	return nil
}

// --- Helpers ---

func (s *BillingService) getPlanName(planID string) string {
	names := map[string]string{
		"free":         "Free",
		"starter":      "Starter",
		"professional": "Professional",
		"enterprise":   "Enterprise",
	}
	if name, ok := names[planID]; ok {
		return name
	}
	return "Unknown"
}

func (s *BillingService) getLimitsForPlan(planID string) PlanLimits {
	limits := map[string]PlanLimits{
		"free": {
			Surveys:        3,
			Responses:      100,
			Members:        1,
			EmailCampaigns: 0,
			AICredits:      0,
		},
		"starter": {
			Surveys:        -1,
			Responses:      1000,
			Members:        5,
			EmailCampaigns: 10,
			AICredits:      0,
		},
		"professional": {
			Surveys:        -1,
			Responses:      10000,
			Members:        20,
			EmailCampaigns: -1,
			AICredits:      100,
		},
		"enterprise": {
			Surveys:        -1,
			Responses:      -1,
			Members:        -1,
			EmailCampaigns: -1,
			AICredits:      -1,
		},
	}
	if limits, ok := limits[planID]; ok {
		return limits
	}
	return limits["free"]
}
