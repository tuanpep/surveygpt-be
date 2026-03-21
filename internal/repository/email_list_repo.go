package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
)

// emailListRepo provides data access for email lists.
type emailListRepo struct {
	pool *pgxpool.Pool
}

// NewEmailListRepo creates a new EmailListRepo.
func NewEmailListRepo(pool *pgxpool.Pool) EmailListRepo {
	return &emailListRepo{pool: pool}
}

// Create creates a new email list.
func (r *emailListRepo) Create(ctx context.Context, list *models.EmailList) error {
	query := `
		INSERT INTO email_lists (org_id, name, status, contact_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		list.OrgID, list.Name, list.Status, list.ContactCount,
	).Scan(&list.ID, &list.CreatedAt, &list.UpdatedAt)
}

// GetByID returns an email list by ID.
func (r *emailListRepo) GetByID(ctx context.Context, id string) (*models.EmailList, error) {
	query := `
		SELECT id, org_id, name, status, contact_count, created_at, updated_at
		FROM email_lists
		WHERE id = $1 AND deleted_at IS NULL`

	list := &models.EmailList{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&list.ID, &list.OrgID, &list.Name, &list.Status,
		&list.ContactCount, &list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// List returns email lists for an organization.
func (r *emailListRepo) List(ctx context.Context, orgID string) ([]models.EmailList, error) {
	query := `
		SELECT id, org_id, name, status, contact_count, created_at, updated_at
		FROM email_lists
		WHERE org_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.EmailList
	for rows.Next() {
		var list models.EmailList
		if err := rows.Scan(
			&list.ID, &list.OrgID, &list.Name, &list.Status,
			&list.ContactCount, &list.CreatedAt, &list.UpdatedAt,
		); err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// Update updates an email list.
func (r *emailListRepo) Update(ctx context.Context, list *models.EmailList) error {
	query := `
		UPDATE email_lists
		SET name = $2, status = $3, contact_count = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	return r.pool.QueryRow(ctx, query,
		list.ID, list.Name, list.Status, list.ContactCount,
	).Scan(&list.UpdatedAt)
}

// Delete soft deletes an email list.
func (r *emailListRepo) Delete(ctx context.Context, id string) error {
	query := `UPDATE email_lists SET deleted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// AddContacts adds contacts to an email list.
func (r *emailListRepo) AddContacts(ctx context.Context, contacts []models.EmailListContact) error {
	query := `
		INSERT INTO email_contacts
		(email_list_id, email, first_name, last_name, metadata, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	for _, c := range contacts {
		_, err := r.pool.Exec(ctx, query, c.ListID, c.Email, c.FirstName, c.LastName, c.Metadata, c.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetContacts returns contacts for an email list.
func (r *emailListRepo) GetContacts(ctx context.Context, listID string) ([]models.EmailListContact, error) {
	query := `
		SELECT id, email_list_id, email, first_name, last_name, metadata, status, created_at
		FROM email_contacts
		WHERE email_list_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.EmailListContact
	for rows.Next() {
		var c models.EmailListContact
		if err := rows.Scan(
			&c.ID, &c.ListID, &c.Email, &c.FirstName, &c.LastName,
			&c.Metadata, &c.Status, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

// RemoveContact removes a contact from an email list.
func (r *emailListRepo) RemoveContact(ctx context.Context, contactID string) error {
	query := `UPDATE email_contacts SET deleted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, contactID)
	return err
}

// IncrementContactCount increments the contact count for a list.
func (r *emailListRepo) IncrementContactCount(ctx context.Context, listID string) error {
	query := `UPDATE email_lists SET contact_count = contact_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, listID)
	return err
}

// DecrementContactCount decrements the contact count for a list.
func (r *emailListRepo) DecrementContactCount(ctx context.Context, listID string) error {
	query := `UPDATE email_lists SET contact_count = GREATEST(0, contact_count - 1) WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, listID)
	return err
}
