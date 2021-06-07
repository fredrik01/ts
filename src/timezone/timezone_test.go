package timezone

import (
	"testing"
	"time"
)

func TestInTimezone(t *testing.T) {
	tzTime := InTimezone(time.Now())
	if tzTime.Hour() != time.Now().Hour() {
		t.Errorf("Hour should be the same when no timezone is set")
	}
}
