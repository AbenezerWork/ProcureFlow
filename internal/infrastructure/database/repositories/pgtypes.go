package repositories

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func optionalDate(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}

	formatted := value.Time.UTC().Format("2006-01-02")
	return &formatted
}

func optionalNumeric(value pgtype.Numeric) *string {
	if !value.Valid {
		return nil
	}

	raw, err := value.Value()
	if err != nil || raw == nil {
		return nil
	}

	text, ok := raw.(string)
	if !ok {
		return nil
	}

	return &text
}

func requiredNumeric(value pgtype.Numeric) string {
	raw, err := value.Value()
	if err != nil || raw == nil {
		return ""
	}

	text, ok := raw.(string)
	if !ok {
		return ""
	}

	return text
}

func parseNumeric(value *string) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}

	var numeric pgtype.Numeric
	if err := numeric.Scan(*value); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("scan numeric: %w", err)
	}

	return numeric, nil
}

func parseDate(value *string) (pgtype.Date, error) {
	if value == nil {
		return pgtype.Date{}, nil
	}

	parsed, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("parse date: %w", err)
	}

	var date pgtype.Date
	if err := date.Scan(parsed); err != nil {
		return pgtype.Date{}, fmt.Errorf("scan date: %w", err)
	}

	return date, nil
}
