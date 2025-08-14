// in file: /pkg/models/sin.go
package models

import (
	"time"
)

// Sin represents the data structure for a code sin.
type Sin struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Count       int       `json:"count"`
	CreatedAt   time.Time `json:"created_at"`
	Tags        []string  `json:"tags"` // Just a standard string slice
	Severity    *int      `json:"severity,omitempty"`
}
