package timex_test

import (
	"go-fiber-stater-kit/pkg/timex"
	"testing"
	"time"
)

func TestIsWorkdayBy(t *testing.T) {
	// Given
	date := "2023-02-14"

	// When
	w, _ := timex.IsWorkdayBy(date, timex.DateFormatDash)

	// Then
	if !w {
		t.Error("It's not a workday.")
	}
}

func TestFormatSlash(t *testing.T) {
	// Given
	d := time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)

	// When
	f := d.Format(timex.DateFormatSlash2)

	// Then
	if f != "02/01/2022" {
		t.Error("Format fail!")
	}
}

func TestDateFormat(t *testing.T) {
	// Given
	d := time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)

	// When
	f := d.Format(timex.DateFormat)

	// Then
	if f != "20220102" {
		t.Error("Format fail!")
	}
}

func TestTimeFormatDash2(t *testing.T) {
	// Given
	d := time.Date(2022, 1, 2, 10, 11, 12, 0, time.UTC)

	// When
	f := d.Format(timex.TimeFormatDash2)

	// Then
	if f != "2022-01-02 10:11" {
		t.Error("Format fail!")
	}
}

func TestDateFormatDash(t *testing.T) {
	// Given
	d := time.Date(2022, 1, 2, 10, 11, 12, 0, time.UTC)

	// When
	f := d.Format(timex.DateFormatDash)

	// Then
	if f != "2022-01-02" {
		t.Error("Format fail!")
	}
}

func TestParse(t *testing.T) {
	// Given
	dt := "3333-3-09T00:00:00.000Z"

	// When
	actual, err := timex.Parse(dt)

	// Then
	if err == nil && !actual.IsZero() {
		t.Error("Parse is not error, actual is", actual)
	}
}
