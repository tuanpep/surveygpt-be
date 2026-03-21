# SurveyFlow API Documentation

Base URL: `http://localhost:8080/api/v1`

All authenticated endpoints require an `Authorization: Bearer <jwt_token>` header.

---

## Table of Contents

- [Authentication](#authentication)
- [Users](#users)
- [Organizations](#organizations)
- [Surveys](#surveys)
- [Responses](#responses)
- [Analytics](#analytics)
- [Templates](#templates)
- [Distribution](#distribution)
- [Billing](#billing)
- [Files](#files)
- [Integrations](#integrations)
- [API Keys](#api-keys)
- [Webhooks](#webhooks)
- [Audit Logs](#audit-logs)
- [Health](#health)

---

## Common Response Format

```json
{
  "data": { ... },
  "message": "optional message"
}
```

### Error Response

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable description"
}
```

### Paginated Response

```json
{
  "data": [...],
  "next_cursor": "base64-cursor-or-empty",
  "total": 42
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid input |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Duplicate resource |
| `VALIDATION_FAILED` | 422 | Request validation failed |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |
| `NOT_IMPLEMENTED` | 501 | Endpoint not yet available |

---

## Authentication

### POST /auth/register

Register a new user and create an organization.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | yes | Valid email (max 255 chars) |
| `password` | string | yes | Min 8, max 128 chars |
| `name` | string | yes | Display name (max 255 chars) |

**Response** `201 Created`

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "two_factor_enabled": false,
      "org_id": "uuid",
      "org_name": "John Doe's Workspace",
      "role": "owner",
      "created_at": "2026-01-01T00:00:00Z"
    },
    "tokens": {
      "access_token": "eyJhbG...",
      "refresh_token": "eyJhbG...",
      "expires_at": "2026-01-01T00:15:00Z"
    }
  }
}
```

---

### POST /auth/login

Authenticate and receive tokens.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `email` | string | yes |
| `password` | string | yes |

**Response** `200 OK`

Same as register response.

---

### POST /auth/login/2fa

Verify a two-factor authentication code.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `email` | string | yes |
| `code` | string | yes |

**Response** `200 OK` — Token pair.

---

### POST /auth/refresh

Exchange a refresh token for a new access token.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `refresh_token` | string | yes |

**Response** `200 OK`

```json
{
  "data": {
    "access_token": "eyJhbG...",
    "refresh_token": "eyJhbG...",
    "expires_at": "2026-01-01T00:15:00Z"
  }
}
```

---

### POST /auth/forgot-password

Request a password reset email. Always returns 200 even if the email doesn't exist.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `email` | string | yes |

**Response** `200 OK`

```json
{
  "message": "if an account with that email exists, a password reset link has been sent"
}
```

---

### POST /auth/reset-password

Reset password using a token from the forgot-password email.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `token` | string | yes |
| `new_password` | string | yes |

**Response** `200 OK`

```json
{
  "message": "password has been reset successfully"
}
```

---

## Users

### GET /me

Get the current authenticated user.

**Response** `200 OK`

```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "two_factor_enabled": false,
    "org_id": "uuid",
    "org_name": "John Doe's Workspace",
    "role": "owner",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

---

### PUT /me

Update the current user's profile.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `name` | string | no |
| `avatar_url` | string | no |

**Response** `200 OK` — Updated user object.

---

### PUT /me/password

Change the current user's password.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `current_password` | string | yes |
| `new_password` | string | yes |

**Response** `200 OK`

```json
{
  "message": "password changed successfully"
}
```

---

### POST /me/avatar `STUB`

Upload user avatar.

**Response** `501 Not Implemented`

---

### DELETE /me/account `STUB`

Delete user account.

**Response** `501 Not Implemented`

---

### POST /me/2fa/enable `STUB`

Enable two-factor authentication.

**Response** `501 Not Implemented`

---

### POST /me/2fa/disable `STUB`

Disable two-factor authentication.

**Response** `501 Not Implemented`

---

### POST /me/2fa/verify `STUB`

Verify 2FA code during setup.

**Response** `501 Not Implemented`

---

## Organizations

### POST /organizations

Create a new organization.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `name` | string | yes |
| `slug` | string | yes |

**Response** `201 Created`

```json
{
  "data": {
    "id": "uuid",
    "name": "My Org",
    "slug": "my-org",
    "plan": "Free",
    "response_limit": 100,
    "survey_limit": 10,
    "member_limit": 5,
    "ai_credits": 50,
    "settings": {},
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```

---

### GET /organizations/me

Get the current user's organization.

**Response** `200 OK` — Same as above.

---

### PUT /organizations/me

Update the current organization.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `name` | string | no |
| `slug` | string | no |
| `billing_email` | string | no |
| `settings` | object | no |

**Response** `200 OK` — Updated organization.

---

### GET /organizations/:orgId

Get an organization by ID (user must be a member).

**Response** `200 OK` — Organization object.

---

### PUT /organizations/:orgId

Update an organization by ID.

**Request Body:** Same as `PUT /organizations/me`.

**Response** `200 OK` — Updated organization.

---

### DELETE /organizations/:orgId `STUB`

Delete an organization.

**Response** `501 Not Implemented`

---

### GET /organizations/:orgId/members

List organization members.

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": "uuid",
      "org_id": "uuid",
      "user_id": "uuid",
      "role": "owner",
      "invited_by": "uuid",
      "joined_at": "2026-01-01T00:00:00Z",
      "user": {
        "id": "uuid",
        "email": "user@example.com",
        "name": "John Doe",
        "avatar_url": null
      }
    }
  ]
}
```

---

### POST /organizations/:orgId/members/invite

Invite a new member to the organization.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | yes | Valid email |
| `role` | string | yes | `owner`, `admin`, `member`, or `viewer` |

**Response** `201 Created`

```json
{
  "data": {
    "id": "uuid",
    "org_id": "uuid",
    "email": "invitee@example.com",
    "role": "member",
    "invited_by": "uuid",
    "token": "invite-token-string",
    "expires_at": "2026-01-08T00:00:00Z",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

**Errors:**
- `422` — Invalid email or role
- `409` — User is already a member or has a pending invitation
- `402` — Organization member limit reached

---

### PATCH /organizations/:orgId/members/:memberId/role

Update a member's role.

**Request Body:**

| Field | Type | Required |
|-------|------|----------|
| `role` | string | yes |

**Response** `200 OK`

```json
{
  "message": "role updated successfully"
}
```

**Errors:**
- `400` — Cannot change the organization owner's role

---

### DELETE /organizations/:orgId/members/:memberId

Remove a member from the organization.

**Response** `200 OK`

```json
{
  "message": "member removed successfully"
}
```

**Errors:**
- `400` — Cannot remove the organization owner

---

### GET /organizations/:orgId/invitations

List pending invitations.

**Response** `200 OK` — Array of invitation objects.

---

### POST /invitations/:token/accept

Accept an organization invitation.

**Response** `200 OK` — Membership object.

---

### POST /invitations/:token/decline

Decline an organization invitation.

**Response** `200 OK`

```json
{
  "message": "invitation declined"
}
```

---

### GET /usage

Get current organization usage statistics.

**Response** `200 OK`

```json
{
  "data": {
    "surveys": 5,
    "responses": 42,
    "members": 3,
    "ai_credits": {
      "limit": 50,
      "used": 10
    }
  }
}
```

---

## Surveys

### POST /surveys

Create a new survey.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | yes | Survey title (max 500 chars) |
| `description` | string | no | Survey description |
| `ui_mode` | string | no | `classic`, `minimal`, `cards`, or `conversational` |

**Response** `201 Created`

```json
{
  "data": {
    "id": "uuid",
    "org_id": "uuid",
    "created_by": "uuid",
    "title": "Customer Feedback",
    "description": "",
    "status": "draft",
    "ui_mode": "classic",
    "structure": { "questions": [], "blocks": [], "flow": [] },
    "settings": {},
    "theme": {},
    "response_count": 0,
    "view_count": 0,
    "published_at": null,
    "closed_at": null,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```

---

### GET /surveys

List surveys for the current organization.

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | — | Filter: `draft`, `published`, `closed` |
| `search` | string | — | Search by title |
| `sort_by` | string | `created_at` | Sort field |
| `sort_dir` | string | `desc` | `asc` or `desc` |
| `cursor` | string | — | Pagination cursor |
| `limit` | int | 20 | Max 100 |

**Response** `200 OK` — Paginated list of survey objects.

---

### GET /surveys/:id

Get a survey by ID.

**Response** `200 OK` — Survey object.

---

### PUT /surveys/:id

Update a survey. All fields are optional (partial update).

**Request Body:**

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | New title |
| `description` | string | New description |
| `ui_mode` | string | New UI mode |
| `status` | string | New status |
| `structure` | object | Survey structure with questions, blocks, flow |
| `settings` | object | Survey settings |
| `theme` | object | Survey theme customization |

**Structure Example:**

```json
{
  "structure": {
    "questions": [
      {
        "id": "q1",
        "title": "How satisfied are you?",
        "type": "rating",
        "required": true,
        "options": [
          { "id": "o1", "label": "Very satisfied" },
          { "id": "o2", "label": "Satisfied" },
          { "id": "o3", "label": "Neutral" },
          { "id": "o4", "label": "Dissatisfied" },
          { "id": "o5", "label": "Very dissatisfied" }
        ]
      }
    ],
    "blocks": [],
    "flow": []
  }
}
```

**Response** `200 OK` — Updated survey object.

---

### DELETE /surveys/:id

Soft-delete a survey.

**Response** `200 OK`

```json
{
  "message": "survey deleted successfully"
}
```

---

### POST /surveys/:id/publish

Publish a survey (must have at least one question).

**Response** `200 OK` — Survey object with `status: "published"`.

**Errors:**
- `400` — Survey has no questions or is not in draft status

---

### POST /surveys/:id/close

Close a published survey (stops accepting responses).

**Response** `200 OK` — Survey object with `status: "closed"`.

---

### POST /surveys/:id/duplicate

Create a copy of a survey with "Copy of" prefix.

**Response** `201 Created` — New survey object.

---

### GET /public/surveys/:slug

Get a published survey for public viewing (no auth required).

**Response** `200 OK`

```json
{
  "data": {
    "id": "uuid",
    "title": "Customer Feedback",
    "description": "",
    "status": "published",
    "ui_mode": "classic",
    "structure": { ... },
    "settings": {},
    "theme": {},
    "share_url": "http://localhost:5173/s/uuid"
  }
}
```

**Errors:**
- `404` — Survey not found
- `400` — Survey is not published

---

## Responses

### POST /surveys/:id/responses

Submit a survey response (public, no auth required).

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `answers` | array | yes | Array of answer objects |
| `metadata` | object | no | Response metadata |

**Answer Object:**

| Field | Type | Description |
|-------|------|-------------|
| `question_id` | string | Question ID to answer |
| `value` | any | Answer value (string, number, array, etc.) |

**Metadata Object:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `duration` | int | — | Time spent in milliseconds |
| `source` | string | `direct` | Response source |
| `device` | string | — | Device type |
| `browser` | string | — | Browser name |
| `os` | string | — | Operating system |
| `country` | string | — | Country code |
| `ip_hash` | string | — | Hashed IP address |
| `language` | string | `en` | Language code |
| `embedded_data` | object | — | Custom embedded data |

**Response** `201 Created`

```json
{
  "data": {
    "success": true,
    "response_id": "uuid",
    "redirect_url": "/s/survey-uuid/thank-you"
  }
}
```

**Errors:**
- `400` — Survey not published, missing required answers, or invalid body
- `404` — Survey not found

---

### GET /surveys/:id/responses

List survey responses (authenticated).

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | — | Filter: `in_progress`, `completed`, `disqualified`, `partial` |
| `cursor` | string | — | Pagination cursor |
| `limit` | int | 20 | Max 100 |
| `date_from` | string | — | Start date (YYYY-MM-DD) |
| `date_to` | string | — | End date (YYYY-MM-DD) |

**Response** `200 OK` — Paginated list of response objects.

---

### GET /surveys/:id/responses/:responseId

Get a single response with answers.

**Response** `200 OK`

```json
{
  "data": {
    "id": "uuid",
    "survey_id": "uuid",
    "status": "completed",
    "duration_ms": 5000,
    "source": "direct",
    "language": "en",
    "started_at": "2026-01-01T00:00:00Z",
    "completed_at": "2026-01-01T00:00:05Z",
    "answers": [
      {
        "id": "uuid",
        "response_id": "uuid",
        "question_id": "q1",
        "value": { ... },
        "created_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### DELETE /surveys/:id/responses/:responseId

Delete a response.

**Response** `200 OK`

```json
{
  "message": "response deleted successfully"
}
```

---

### POST /surveys/:id/responses/export

Export survey responses.

**Request Body:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `json` | `json`, `csv`, or `xlsx` |
| `status` | string | — | Filter by response status |
| `date_from` | string | — | Start date (YYYY-MM-DD) |
| `date_to` | string | — | End date (YYYY-MM-DD) |

**Response** `200 OK`

```json
{
  "data": "...",
  "format": "json"
}
```

---

## Analytics

### GET /surveys/:id/analytics/summary

Get overall analytics for a survey.

**Response** `200 OK`

```json
{
  "data": {
    "total_responses": 42,
    "completed_responses": 38,
    "completion_rate": 90.48,
    "avg_duration_ms": 5000,
    "total_views": 150,
    "conversion_rate": 28.0,
    "first_response_at": "2026-01-01T00:00:00Z",
    "last_response_at": "2026-01-05T00:00:00Z",
    "daily_counts": [
      { "date": "2026-01-01", "count": 10 },
      { "date": "2026-01-02", "count": 15 }
    ],
    "device_breakdown": { "desktop": 30, "mobile": 12 },
    "browser_breakdown": { "Chrome": 25, "Firefox": 10, "Safari": 7 },
    "country_breakdown": { "US": 20, "GB": 10, "DE": 8 },
    "source_breakdown": { "direct": 25, "email": 10, "social": 7 }
  }
}
```

---

### GET /surveys/:id/analytics/questions/:questionId

Get statistics for a specific question.

**Response** `200 OK`

```json
{
  "data": {
    "question_id": "q1",
    "question_type": "rating",
    "question_title": "How satisfied are you?",
    "total_answers": 40,
    "skipped_count": 2,
    "choice_stats": [
      { "choice_id": "o1", "label": "Very satisfied", "count": 10, "percentage": 25.0 },
      { "choice_id": "o2", "label": "Satisfied", "count": 15, "percentage": 37.5 }
    ],
    "numeric_stats": {
      "mean": 3.5,
      "median": 4.0,
      "min": 1.0,
      "max": 5.0,
      "std_dev": 1.2
    },
    "text_stats": {
      "total_words": 200,
      "avg_word_count": 10.0,
      "top_keywords": ["great", "easy", "fast"],
      "sentiment_score": 0.75
    }
  }
}
```

---

### GET /surveys/:id/analytics/cross-tab

Get cross-tabulation between two questions.

**Query Parameters:**

| Param | Type | Required |
|-------|------|----------|
| `row_question_id` | string | yes |
| `col_question_id` | string | yes |

**Response** `200 OK`

```json
{
  "data": {
    "row_question_id": "q1",
    "column_question_id": "q2",
    "headers": ["Option A", "Option B", "Option C"],
    "rows": [
      { "label": "Choice 1", "values": [5, 3, 2], "total": 10 },
      { "label": "Choice 2", "values": [8, 1, 1], "total": 10 }
    ]
  }
}
```

**Errors:**
- `400` — Missing `row_question_id` or `col_question_id`

---

### GET /surveys/:id/analytics/dropoff

Get question-by-question dropoff rates.

**Response** `200 OK`

```json
{
  "data": {
    "steps": [
      {
        "step_id": "q1",
        "step_type": "question",
        "step_label": "How satisfied are you?",
        "views": 150,
        "dropoffs": 20,
        "dropoff_rate": 13.33
      },
      {
        "step_id": "q2",
        "step_type": "question",
        "step_label": "Would you recommend us?",
        "views": 130,
        "dropoffs": 50,
        "dropoff_rate": 38.46
      }
    ]
  }
}
```

---

### POST /surveys/:id/analytics/ai-insights

Get AI-generated insights about survey responses.

**Response** `200 OK`

```json
{
  "data": {
    "summary": "Overall satisfaction is high with room for improvement in support.",
    "key_findings": ["92% of respondents are satisfied", "Main concern is response time"],
    "recommendations": ["Improve support response time", "Add follow-up questions"],
    "sentiment": "positive",
    "themes": [
      { "name": "Support Speed", "description": "Response time concerns", "confidence": 0.85, "mentions": 15 },
      { "name": "Product Quality", "description": "Positive feedback on features", "confidence": 0.92, "mentions": 25 }
    ],
    "anomalies": [
      { "type": "low_response", "description": "Question 3 has unusually low response rate", "severity": "medium", "question_id": "q3" }
    ],
    "generated_at": "2026-01-01T00:00:00Z"
  }
}
```

---

## Templates

### GET /templates

List available templates (system + custom).

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `category` | string | — | Filter by category |

**Response** `200 OK` — Array of template objects.

---

### GET /templates/:id

Get a template by ID.

**Response** `200 OK` — Template object.

---

### POST /templates

Save a survey as a custom template.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `survey_id` | string | yes | Survey to save as template |
| `name` | string | yes | Template name |
| `category` | string | no | Defaults to `custom` |

**Response** `201 Created` — Template object.

---

### POST /templates/:id/duplicate

Create a new survey from a template.

**Response** `201 Created` — Survey object.

---

### POST /surveys/from-template/:id

Create a survey from a template (alternative path).

**Response** `201 Created` — Survey object.

---

## Distribution

### GET /surveys/:id/qr-code

Generate a QR code for a survey.

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `size` | string | — | QR code size in pixels |
| `error_level` | string | — | `L`, `M`, `Q`, or `H` |

**Response** `200 OK`

```json
{
  "data": {
    "svg": "<svg>...</svg>",
    "png_url": "https://api.qrserver.com/v1/..."
  }
}
```

---

### GET /surveys/:id/embed-code

Get embed code for a survey.

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | — | `popup`, `embedded`, or `fullpage` |
| `trigger` | string | — | `auto`, `button`, or `custom` |
| `width` | string | — | Embed width |
| `height` | string | — | Embed height |

**Response** `200 OK`

```json
{
  "data": {
    "html": "<div id=\"surveyflow-embed\">...",
    "js": "<script>..."
  }
}
```

---

### GET /email-lists

List email lists for the organization.

**Response** `200 OK` — Array of email list objects.

---

### POST /email-lists

Create an email list with optional contacts.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | List name |
| `contacts` | array | no | Initial contacts |

**Contact Object:**

| Field | Type | Required |
|-------|------|----------|
| `email` | string | yes |
| `first_name` | string | no |
| `last_name` | string | no |
| `metadata` | object | no |

**Response** `201 Created` — Email list object.

---

### GET /email-lists/:id `STUB`

Get email list details.

**Response** `501 Not Implemented`

---

### PUT /email-lists/:id `STUB`

Update an email list.

**Response** `501 Not Implemented`

---

### DELETE /email-lists/:id `STUB`

Delete an email list.

**Response** `501 Not Implemented`

---

### POST /email-lists/:id/contacts `STUB`

Add contacts to an email list.

**Response** `501 Not Implemented`

---

### DELETE /email-lists/:id/contacts/:contactId `STUB`

Remove a contact from an email list.

**Response** `501 Not Implemented`

---

### POST /surveys/:id/send-emails

Send survey via email to an email list.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `list_id` | string | yes | Email list ID |
| `subject` | string | yes | Email subject |
| `body` | string | no | Email body (HTML) |
| `schedule_at` | string | no | ISO 8601 timestamp for scheduling |

**Response** `200 OK`

```json
{
  "data": {
    "job_id": "uuid",
    "count": 150,
    "status": "scheduled"
  }
}
```

---

### GET /surveys/:id/email/campaigns `STUB`

List email campaigns for a survey.

**Response** `501 Not Implemented`

---

## Billing

### GET /billing/plan

Get the current organization's billing plan.

**Response** `200 OK`

```json
{
  "data": {
    "plan_id": "free",
    "plan_name": "Free",
    "billing_period": "monthly",
    "status": "active",
    "trial_ends_at": null,
    "next_billing_date": null,
    "limits": {
      "surveys": 10,
      "responses": 100,
      "members": 5,
      "email_campaigns": 0,
      "ai_credits": 50
    },
    "usage": {
      "surveys": 3,
      "responses": 42,
      "members": 1,
      "email_campaigns": 0,
      "ai_credits": 5
    }
  }
}
```

---

### GET /billing/plans

List all available billing plans.

**Response** `200 OK` — Array of plan objects.

---

### POST /billing/change-plan

Initiate a plan change (creates a Stripe checkout session).

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `plan_id` | string | yes | Target plan ID |
| `billing_period` | string | no | `monthly` or `yearly` |
| `success_url` | string | no | Redirect URL on success |
| `cancel_url` | string | no | Redirect URL on cancel |

**Response** `200 OK`

```json
{
  "data": {
    "checkout_url": "https://checkout.stripe.com/..."
  }
}
```

---

### GET /billing/portal

Get the Stripe customer portal URL for managing subscriptions.

**Response** `200 OK`

```json
{
  "data": {
    "url": "https://billing.stripe.com/session/..."
  }
}
```

---

### GET /billing/history

Get invoice/payment history.

**Response** `200 OK` — Array of invoice objects.

---

## Files

### POST /files/upload `STUB`

Upload a file.

**Response** `501 Not Implemented`

---

### GET /files `STUB`

List uploaded files.

**Response** `501 Not Implemented`

---

### DELETE /files/:id `STUB`

Delete a file.

**Response** `501 Not Implemented`

---

## Integrations

### GET /integrations `STUB`

List integrations.

**Response** `501 Not Implemented`

---

### POST /integrations `STUB`

Create an integration.

**Response** `501 Not Implemented`

---

### PUT /integrations/:id `STUB`

Update an integration.

**Response** `501 Not Implemented`

---

### DELETE /integrations/:id `STUB`

Delete an integration.

**Response** `501 Not Implemented`

---

## API Keys

### GET /api-keys `STUB`

List API keys.

**Response** `501 Not Implemented`

---

### POST /api-keys `STUB`

Create an API key.

**Response** `501 Not Implemented`

---

### DELETE /api-keys/:id `STUB`

Revoke an API key.

**Response** `501 Not Implemented`

---

## Webhooks

### GET /webhooks `STUB`

List webhooks.

**Response** `501 Not Implemented`

---

### POST /webhooks `STUB`

Create a webhook.

**Response** `501 Not Implemented`

---

### PUT /webhooks/:id `STUB`

Update a webhook.

**Response** `501 Not Implemented`

---

### DELETE /webhooks/:id `STUB`

Delete a webhook.

**Response** `501 Not Implemented`

---

### POST /webhooks/stripe

Receive Stripe webhook events (no auth required, verified via `Stripe-Signature` header).

**Response** `200 OK`

```json
{
  "data": {
    "received": true
  }
}
```

---

## Audit Logs

### GET /audit-logs `STUB`

List audit log entries.

**Response** `501 Not Implemented`

---

## Health

### GET /health

Health check endpoint (unauthenticated).

**Response** `200 OK`

```json
{
  "status": "ok",
  "deps": {
    "database": "ok",
    "redis": "ok"
  }
}
```

Returns `"degraded"` if any dependency is down.

---

### GET /healthz

Alias for `/health`.

---

### GET /ready

Readiness check (unauthenticated).

**Response** `200 OK`

```json
{
  "status": "ready"
}
```

Returns `503` if the database is not reachable.

---

## OAuth

### GET /auth/google `STUB`

Initiate Google OAuth flow.

**Response** `501 Not Implemented`

---

### GET /auth/google/callback `STUB`

Google OAuth callback.

**Response** `501 Not Implemented`

---

### GET /auth/microsoft `STUB`

Initiate Microsoft OAuth flow.

**Response** `501 Not Implemented`

---

### GET /auth/microsoft/callback `STUB`

Microsoft OAuth callback.

**Response** `501 Not Implemented`

---

## Real-time Updates

### GET /surveys/:id/live

Server-Sent Events stream for live response updates (authenticated).

**Response** `200 OK` with `Content-Type: text/event-stream`

Events:
- `connected` — Sent on initial connection
- `update` — New response submitted or survey updated

---

## Rate Limiting

All endpoints are subject to a global rate limit of **100 requests per minute per IP**.

Rate limit headers are included in every response:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704060800
```

When rate limited, the API returns `429 Too Many Requests`:

```json
{
  "code": "RATE_LIMITED",
  "message": "too many requests, please try again later"
}
```

---

## CORS

The API supports CORS for the configured frontend URL. The following headers are allowed:

- `Authorization`
- `Content-Type`
- `X-Request-ID`
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`
