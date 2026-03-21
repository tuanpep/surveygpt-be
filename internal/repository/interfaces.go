package repository

import (
	"context"
	"time"

	"github.com/surveyflow/be/internal/models"
)

// UserRepo provides data access for user records.
type UserRepo interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	SetEmailVerified(ctx context.Context, userID string) error
	Set2FASecret(ctx context.Context, userID, secret string) error
	Enable2FA(ctx context.Context, userID string) error
	Disable2FA(ctx context.Context, userID string) error
}

// OrgRepo provides data access for organizations, memberships, and invitations.
type OrgRepo interface {
	CreateOrg(ctx context.Context, org *models.Organization) error
	GetOrgByID(ctx context.Context, id string) (*models.Organization, error)
	GetOrgBySlug(ctx context.Context, slug string) (*models.Organization, error)
	UpdateOrg(ctx context.Context, org *models.Organization) error
	ListMembers(ctx context.Context, orgID string) ([]MemberWithUser, error)
	GetMembership(ctx context.Context, orgID, userID string) (*models.OrgMembership, error)
	AddMember(ctx context.Context, orgID, userID, invitedBy string, role models.Role) (*models.OrgMembership, error)
	UpdateMemberRole(ctx context.Context, orgID, userID string, role models.Role) error
	RemoveMember(ctx context.Context, orgID, userID string) error
	GetMemberCount(ctx context.Context, orgID string) (int, error)
	CreateInvitation(ctx context.Context, inv *models.Invitation) error
	GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error)
	ListInvitations(ctx context.Context, orgID string) ([]models.Invitation, error)
	AcceptInvitation(ctx context.Context, token string) error
	DeleteInvitation(ctx context.Context, token string) error
	GetMembershipsByUserID(ctx context.Context, userID string) ([]models.OrgMembership, error)
	GetPendingInvitation(ctx context.Context, orgID, email string) (*models.Invitation, error)
	GetUsage(ctx context.Context, orgID string) (*OrgUsage, error)
}

// SurveyRepo provides data access for surveys.
type SurveyRepo interface {
	Create(ctx context.Context, s *models.Survey) error
	GetByID(ctx context.Context, id string) (*models.Survey, error)
	List(ctx context.Context, orgID, status, search, sortBy, sortOrder string, limit int, cursor string) ([]models.Survey, string, int, error)
	Update(ctx context.Context, s *models.Survey) error
	UpdateStatus(ctx context.Context, id string, status models.SurveyStatus) error
	SoftDelete(ctx context.Context, id string) error
	Duplicate(ctx context.Context, id, newTitle string) (*models.Survey, error)
	IncrementViewCount(ctx context.Context, id string) error
	IncrementResponseCount(ctx context.Context, id string) error
	CountByOrg(ctx context.Context, orgID string) (int, error)
}

// ResponseRepo provides data access for responses and answers.
type ResponseRepo interface {
	Create(ctx context.Context, resp *models.Response) error
	GetByID(ctx context.Context, id string) (*models.Response, error)
	List(ctx context.Context, surveyID, status string, dateFrom, dateTo *time.Time, limit int, cursor string) ([]models.Response, string, int, error)
	Complete(ctx context.Context, id string, durationMs int) error
	Delete(ctx context.Context, id string) error
	CountBySurvey(ctx context.Context, surveyID string) (int, error)
	CountCompletedBySurvey(ctx context.Context, surveyID string) (int, error)
	GetAvgDuration(ctx context.Context, surveyID string) (int, error)
	GetResponsesByDay(ctx context.Context, surveyID string, days int) ([]models.DailyCount, error)
	GetResponsesBySource(ctx context.Context, surveyID string) (map[string]int, error)
	GetResponsesByDevice(ctx context.Context, surveyID string) (map[string]int, error)
	CreateAnswers(ctx context.Context, answers []models.Answer) error
	GetAnswersByResponseID(ctx context.Context, responseID string) ([]models.Answer, error)
	GetAnswerCountsByQuestion(ctx context.Context, surveyID, questionID string) (int, error)
	GetChoiceDistribution(ctx context.Context, surveyID, questionID string) (map[string]int, error)
}

// TemplateRepo provides data access for templates.
type TemplateRepo interface {
	GetByID(ctx context.Context, id string) (*models.Template, error)
	List(ctx context.Context, category, orgID string) ([]models.Template, error)
	Create(ctx context.Context, t *models.Template) error
	IncrementUseCount(ctx context.Context, id string) error
}

// AnalyticsRepo provides data access for analytics queries.
type AnalyticsRepo interface {
	GetSummary(ctx context.Context, surveyID string) (*models.AnalyticsSummary, error)
	GetResponsesByDay(ctx context.Context, surveyID string, days int) ([]models.DailyCount, error)
	GetDeviceBreakdown(ctx context.Context, surveyID string) (map[string]int, error)
	GetSourceBreakdown(ctx context.Context, surveyID string) (map[string]int, error)
	GetQuestionStats(ctx context.Context, surveyID, questionID, questionType string) (*models.QuestionStats, error)
	GetCrossTab(ctx context.Context, surveyID, rowQuestionID, colQuestionID string) (*models.CrossTabResult, error)
	GetDropoff(ctx context.Context, surveyID string) ([]models.DropoffStep, error)
}

// EmailListRepo provides data access for email lists.
type EmailListRepo interface {
	Create(ctx context.Context, list *models.EmailList) error
	GetByID(ctx context.Context, id string) (*models.EmailList, error)
	List(ctx context.Context, orgID string) ([]models.EmailList, error)
	Update(ctx context.Context, list *models.EmailList) error
	Delete(ctx context.Context, id string) error
	AddContacts(ctx context.Context, contacts []models.EmailListContact) error
	GetContacts(ctx context.Context, listID string) ([]models.EmailListContact, error)
	RemoveContact(ctx context.Context, contactID string) error
	IncrementContactCount(ctx context.Context, listID string) error
	DecrementContactCount(ctx context.Context, listID string) error
}
