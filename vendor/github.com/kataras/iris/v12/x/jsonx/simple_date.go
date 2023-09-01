package jsonx

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// SimpleDateLayout represents the "year-month-day" Go time format.
const SimpleDateLayout = "2006-01-02"

// SimpleDate holds a json "year-month-day" time.
type SimpleDate time.Time

// SimpleDateFromTime accepts a "t" Time and returns
// a SimpleDate. If format fails, it returns the zero value of time.Time.
func SimpleDateFromTime(t time.Time) SimpleDate {
	date, _ := ParseSimpleDate(t.Format(SimpleDateLayout))
	return date
}

// ParseSimpleDate reads from "s" and returns the SimpleDate time.
func ParseSimpleDate(s string) (SimpleDate, error) {
	if s == "" || s == "null" {
		return SimpleDate{}, nil
	}

	var (
		tt  time.Time
		err error
	)

	tt, err = time.Parse(SimpleDateLayout, s)
	if err != nil {
		return SimpleDate{}, err
	}

	return SimpleDate(tt.UTC()), nil
}

// UnmarshalJSON binds the json "data" to "t" with the `SimpleDateLayout`.
func (t *SimpleDate) UnmarshalJSON(data []byte) error {
	if isNull(data) {
		return nil
	}

	data = trimQuotes(data)
	if len(data) == 0 {
		return nil // as an excepption here, allow empty "" on simple dates, as the server would render it on a response.
	}

	dataStr := string(data)

	tt, err := time.Parse(SimpleDateLayout, dataStr)
	if err != nil {
		return err
	}

	*t = SimpleDate(tt)
	return nil
}

// MarshalJSON returns the json representation of the "t".
func (t SimpleDate) MarshalJSON() ([]byte, error) {
	if s := t.String(); s != "" {
		s = strconv.Quote(s)
		return []byte(s), nil
	}

	return emptyQuoteBytes, nil
}

// IsZero reports whether "t" is zero time.
// It completes the pg.Zeroer interface.
func (t SimpleDate) IsZero() bool {
	return t.ToTime().IsZero()
}

// After reports whether the time instant t is after u.
func (t SimpleDate) After(d2 SimpleDate) bool {
	t1, t2 := t.ToTime(), d2.ToTime()
	return t1.Truncate(24 * time.Hour).After(t2.Truncate(24 * time.Hour))
}

// Before reports whether the time instant t is before u.
func (t SimpleDate) Before(d2 SimpleDate) bool {
	t1, t2 := t.ToTime(), d2.ToTime()
	return t1.Truncate(24 * time.Hour).Before(t2.Truncate(24 * time.Hour))
	// OR: compare year and year's day.
}

// ToTime returns the standard time type.
func (t SimpleDate) ToTime() time.Time {
	return time.Time(t)
}

func (t SimpleDate) Value() (driver.Value, error) {
	return t.String(), nil
}

// String returns the text representation of the date
// formatted based on the `SimpleDateLayout`.
// If date is zero it returns an empty string.
func (t SimpleDate) String() string {
	tt := t.ToTime()
	if tt.IsZero() {
		return ""
	}

	return tt.Format(SimpleDateLayout)
}

// Scan completes the pg and native sql driver.Scanner interface
// reading functionality of a custom type.
func (t *SimpleDate) Scan(src interface{}) error {
	switch v := src.(type) {
	case time.Time: // type was set to timestamp
		if v.IsZero() {
			return nil // don't set zero, ignore it.
		}
		*t = SimpleDate(v)
	case string:
		tt, err := ParseSimpleDate(v)
		if err != nil {
			return err
		}
		*t = tt
	case nil:
		*t = SimpleDate(time.Time{})
	default:
		return fmt.Errorf("SimpleDate: unknown type of: %T", v)
	}

	return nil
}
