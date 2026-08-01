package catalog

import "time"

type Team struct {
	ID        int64
	Slug      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
