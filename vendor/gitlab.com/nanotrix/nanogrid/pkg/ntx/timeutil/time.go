package timeutil

import (
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
)

const TicksPerSecond = uint64(10000000)
const NanosecondsPerSecond = uint64(1e9)
const NanosecondsInTicks = uint64(100)

type NowFn func() time.Time

func UnixZero() time.Time {
	return time.Unix(0, 0).UTC()
}

func NowUTC() time.Time {
	return time.Now().UTC()
}

func NowUTCUnixSeconds() uint64 {
	return uint64(time.Now().UTC().Unix())
}

func NowUTCString() string {
	return FormatTime(time.Now().UTC())
}

func DurationToTicks(d time.Duration) uint64 {
	return uint64(d.Nanoseconds()) / NanosecondsInTicks
}

func DurationToXMLDuration(d time.Duration) string {
	return fmt.Sprintf("PT%0.3fS", d.Seconds())
}

const TICKS_MAX_VALUE = math.MaxInt64

func MaxIfZero(t uint64) uint64 {
	if t == 0 {
		return TICKS_MAX_VALUE
	} else {
		return t
	}
}

func TicksToTime(t uint64) time.Time {
	return time.Time{}.Add(TicksToDuration(t))
}

func TicksToDuration(t uint64) time.Duration {
	return time.Duration(uint64(t) * NanosecondsInTicks)
}

func TicksToXMLDuration(t uint64) string {
	return DurationToXMLDuration(TicksToDuration(t))
}

func ParseTime(str string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, str)
}

func ValidTimeIfNotBlank(str string) error {
	if str == "" {
		return nil
	} else {
		return ValidTime(str)
	}
}

func ValidTime(str string) error {
	_, err := time.Parse(time.RFC3339Nano, str)
	return errors.Wrapf(err, "Unable to parse date '%v'", str)
}

func MustParseTime(str string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, str); err != nil {
		panic(errors.Wrapf(err, "Unable to parse date '%v' - should be checked before this stage in code", str))
	} else {
		return t
	}
}

func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func FormatUnixSeconds(s uint64) string {
	return time.Unix(int64(s), 0).UTC().Format(time.RFC3339)
}
