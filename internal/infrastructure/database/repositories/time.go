package repositories

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	t := value.Time.UTC()
	return &t
}

func requiredTime(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}
