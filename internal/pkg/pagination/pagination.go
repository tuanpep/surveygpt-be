package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	// DefaultLimit is the default number of items per page.
	DefaultLimit = 20

	// MaxLimit is the maximum allowed number of items per page.
	MaxLimit = 100
)

// CursorRequest holds pagination parameters for cursor-based pagination.
type CursorRequest struct {
	// Limit is the maximum number of items to return.
	Limit int `json:"limit" query:"limit"`

	// Cursor is an opaque base64-encoded cursor from a previous response.
	Cursor string `json:"cursor" query:"cursor"`
}

// Normalize validates and applies defaults to the cursor request.
func (r *CursorRequest) Normalize() {
	if r.Limit <= 0 {
		r.Limit = DefaultLimit
	}
	if r.Limit > MaxLimit {
		r.Limit = MaxLimit
	}
}

// ParseFromRequest extracts pagination parameters from an HTTP request.
func ParseFromRequest(req *http.Request) CursorRequest {
	limit := DefaultLimit
	if v := req.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return CursorRequest{
		Limit:  limit,
		Cursor: req.URL.Query().Get("cursor"),
	}
}

// CursorResponse wraps a paginated result set with cursor metadata.
type CursorResponse struct {
	Data       any    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	Total      int    `json:"total,omitempty"`
}

// cursorData is the internal structure encoded/decoded in the cursor string.
type cursorData struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"` // Unix nano
}

// EncodeCursor creates an opaque, base64-encoded cursor from an ID and timestamp.
func EncodeCursor(id string, createdAt int64) (string, error) {
	data := cursorData{
		ID:        id,
		CreatedAt: createdAt,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}

	return base64.URLEncoding.EncodeToString(raw), nil
}

// DecodeCursor decodes a base64-encoded cursor into its ID and timestamp components.
// Returns an empty string and zero timestamp if the cursor is empty or invalid.
func DecodeCursor(cursor string) (id string, createdAt int64, err error) {
	if cursor == "" {
		return "", 0, nil
	}

	raw, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", 0, fmt.Errorf("decode cursor: %w", err)
	}

	var data cursorData
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", 0, fmt.Errorf("unmarshal cursor: %w", err)
	}

	return data.ID, data.CreatedAt, nil
}
