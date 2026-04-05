// @Package carbon
// @Description a simple, semantic and developer-friendly time package for golang
// @Page github.com/golang-module/carbon
// @Developer gouguoyin
// @Blog www.gouguoyin.com
// @Email 245629560@qq.com

// Package carbon is a simple, semantic and developer-friendly time package for golang.
package carbon

import (
	"github.com/dromara/carbon/v2"
)

type StdTime = carbon.StdTime
type Weekday = carbon.Weekday
type Location = carbon.Location
type Duration = carbon.Duration
type Carbon = carbon.Carbon
type Default = carbon.Default
type Language = carbon.Language

// NewCarbon returns a new Carbon instance.
func NewCarbon(time ...StdTime) *Carbon {
	return carbon.NewCarbon(time...)
}

func NewLanguage() *Language {
	return carbon.NewLanguage()
}

func Now(timezone ...string) *Carbon {
	return carbon.Now(timezone...)
}

func Tomorrow(timezone ...string) *Carbon {
	return carbon.Tomorrow(timezone...)
}

func Yesterday(timezone ...string) *Carbon {
	return carbon.Yesterday(timezone...)
}

// Parse parses a standard time string as a Carbon instance.
func Parse(value string, timezone ...string) *Carbon {
	return carbon.Parse(value, timezone...)
}

// ParseByLayout parses a time string as a Carbon instance by a confirmed layout
func ParseByLayout(value, layout string, timezone ...string) *Carbon {
	return carbon.ParseByLayout(value, layout, timezone...)
}

// ParseByFormat parses a time string as a Carbon instance by a confirmed format.
func ParseByFormat(value, format string, timezone ...string) *Carbon {
	return carbon.ParseByFormat(value, format, timezone...)
}

// ParseByLayouts parses a time string as a Carbon instance by multiple fuzzy layouts.
//
// Note: it doesn't support parsing by timestamp layouts.
func ParseByLayouts(value string, layouts []string, timezone ...string) *Carbon {
	return carbon.ParseByLayouts(value, layouts, timezone...)
}

// ParseByFormats parses a time string as a Carbon instance by multiple fuzzy formats.
//
// Note: it doesn't support parsing by timestamp formats.
func ParseByFormats(value string, formats []string, timezone ...string) *Carbon {
	return carbon.ParseByFormats(value, formats, timezone...)
}

// ParseWithLayouts parses a time string as a Carbon instance by multiple fuzzy layouts.
//
// Deprecated: it will be removed in the future, use ParseByLayouts instead.
func ParseWithLayouts(value string, layouts []string, timezone ...string) *Carbon {
	return carbon.ParseWithLayouts(value, layouts, timezone...)
}

// ParseWithFormats parses a time string as a Carbon instance by multiple fuzzy formats.
//
// Deprecated: it will be removed in the future, use ParseByFormats instead.
func ParseWithFormats(value string, formats []string, timezone ...string) *Carbon {
	return carbon.ParseWithFormats(value, formats, timezone...)
}

// CreateFromStdTime creates a Carbon instance from standard time.Time.
func CreateFromStdTime(stdTime StdTime, timezone ...string) *Carbon {
	return carbon.CreateFromStdTime(stdTime, timezone...)
}

// CreateFromTimestamp creates a Carbon instance from a given timestamp with second precision.
func CreateFromTimestamp(timestamp int64, timezone ...string) *Carbon {
	return carbon.CreateFromTimestamp(timestamp, timezone...)
}

// CreateFromTimestampMilli creates a Carbon instance from a given timestamp with millisecond precision.
func CreateFromTimestampMilli(timestampMilli int64, timezone ...string) *Carbon {
	return carbon.CreateFromTimestampMilli(timestampMilli, timezone...)
}

// CreateFromTimestampMicro creates a Carbon instance from a given timestamp with microsecond precision.
func CreateFromTimestampMicro(timestampMicro int64, timezone ...string) *Carbon {
	return carbon.CreateFromTimestampMicro(timestampMicro, timezone...)
}

// CreateFromTimestampNano creates a Carbon instance from a given timestamp with nanosecond precision.
func CreateFromTimestampNano(timestampNano int64, timezone ...string) *Carbon {
	return carbon.CreateFromTimestampNano(timestampNano, timezone...)
}

// CreateFromDateTime creates a Carbon instance from a given date and time.
func CreateFromDateTime(year, month, day, hour, minute, second int, timezone ...string) *Carbon {
	return carbon.CreateFromDateTime(year, month, day, hour, minute, second, timezone...)
}

// CreateFromDateTimeMilli creates a Carbon instance from a given date, time and millisecond.
func CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond int, timezone ...string) *Carbon {
	return carbon.CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond, timezone...)
}

// CreateFromDateTimeMicro creates a Carbon instance from a given date, time and microsecond.
func CreateFromDateTimeMicro(year, month, day, hour, minute, second, microsecond int, timezone ...string) *Carbon {
	return carbon.CreateFromDateTimeMicro(year, month, day, hour, minute, second, microsecond, timezone...)
}

// CreateFromDateTimeNano creates a Carbon instance from a given date, time and nanosecond.
func CreateFromDateTimeNano(year, month, day, hour, minute, second, nanosecond int, timezone ...string) *Carbon {
	return carbon.CreateFromDateTimeNano(year, month, day, hour, minute, second, nanosecond, timezone...)
}

// CreateFromDate creates a Carbon instance from a given date.
func CreateFromDate(year, month, day int, timezone ...string) *Carbon {
	return carbon.CreateFromDate(year, month, day, timezone...)
}

// CreateFromDateMilli creates a Carbon instance from a given date and millisecond.
func CreateFromDateMilli(year, month, day, millisecond int, timezone ...string) *Carbon {
	return carbon.CreateFromDateMilli(year, month, day, millisecond, timezone...)
}

// CreateFromDateMicro creates a Carbon instance from a given date and microsecond.
func CreateFromDateMicro(year, month, day, microsecond int, timezone ...string) *Carbon {
	return carbon.CreateFromDateMicro(year, month, day, microsecond, timezone...)
}

// CreateFromDateNano creates a Carbon instance from a given date and nanosecond.
func CreateFromDateNano(year, month, day, nanosecond int, timezone ...string) *Carbon {
	return carbon.CreateFromDateNano(year, month, day, nanosecond, timezone...)
}

// CreateFromTime creates a Carbon instance from a given time(year, month and day are taken from the current time).
func CreateFromTime(hour, minute, second int, timezone ...string) *Carbon {
	return carbon.CreateFromTime(hour, minute, second, timezone...)
}

// CreateFromTimeMilli creates a Carbon instance from a given time and millisecond(year, month and day are taken from the current time).
func CreateFromTimeMilli(hour, minute, second, millisecond int, timezone ...string) *Carbon {
	return carbon.CreateFromTimeMilli(hour, minute, second, millisecond, timezone...)
}

// CreateFromTimeMicro creates a Carbon instance from a given time and microsecond(year, month and day are taken from the current time).
func CreateFromTimeMicro(hour, minute, second, microsecond int, timezone ...string) *Carbon {
	return carbon.CreateFromTimeMicro(hour, minute, second, microsecond, timezone...)
}

// CreateFromTimeNano creates a Carbon instance from a given time and nanosecond(year, month and day are taken from the current time).
func CreateFromTimeNano(hour, minute, second, nanosecond int, timezone ...string) *Carbon {
	return carbon.CreateFromTimeNano(hour, minute, second, nanosecond, timezone...)
}

// SetLayout sets globally default layout.
func SetLayout(layout string) *Carbon {
	c := carbon.SetLayout(layout)
	if !c.HasError() {
		DefaultLayout = layout
	}
	return c
}

// SetFormat sets globally default format.
func SetFormat(format string) *Carbon {
	c := carbon.SetFormat(format)
	if !c.HasError() {
		DefaultLayout = c.CurrentLayout()
	}
	return c
}

// SetTimezone sets globally default timezone.
func SetTimezone(name string) *Carbon {
	c := carbon.SetTimezone(name)
	if !c.HasError() {
		DefaultTimezone = name
	}
	return c
}

// SetLocation sets globally default location.
func SetLocation(loc *Location) *Carbon {
	c := carbon.SetLocation(loc)
	if !c.HasError() {
		DefaultTimezone = loc.String()
	}
	return c
}

// SetLocale sets globally default locale.
func SetLocale(locale string) *Carbon {
	c := carbon.SetLocale(locale)
	if !c.HasError() {
		DefaultLocale = locale
	}
	return c
}

// SetWeekStartsAt sets globally default start day of the week.
func SetWeekStartsAt(weekday Weekday) *Carbon {
	c := carbon.SetWeekStartsAt(weekday)
	if !c.HasError() {
		DefaultWeekStartsAt = weekday
	}
	return c
}

// SetWeekendDays sets globally default weekend days of the week.
func SetWeekendDays(weekDays []Weekday) *Carbon {
	c := carbon.SetWeekendDays(weekDays)
	if !c.HasError() {
		DefaultWeekendDays = weekDays
	}
	return c
}

// SetTestNow sets a test Carbon instance for now, remember to clear after use.
func SetTestNow(c *Carbon) {
	carbon.SetTestNow(c)
}

// ClearTestNow clears the test Carbon instance for now.
func ClearTestNow() {
	carbon.ClearTestNow()
}

// CleanTestNow cleans the test Carbon instance for now.
// Deprecated: it will be removed in the future, use ClearTestNow instead.
func CleanTestNow() {
	carbon.CleanTestNow()
}

// IsTestNow reports whether is testing time.
func IsTestNow() bool {
	return carbon.IsTestNow()
}

// SetDefault sets default.
func SetDefault(d Default) {
	if d.Layout != "" {
		DefaultLayout = d.Layout
	}
	if d.Timezone != "" {
		DefaultTimezone = d.Timezone
	}
	if d.Locale != "" {
		DefaultLocale = d.Locale
	}
	if d.WeekStartsAt.String() != "" {
		DefaultWeekStartsAt = d.WeekStartsAt
	}
	if len(d.WeekendDays) > 0 {
		DefaultWeekendDays = d.WeekendDays
	}
}

// ResetDefault resets default.
func ResetDefault() {
	carbon.ResetDefault()
}

// ZeroValue returns the zero value of Carbon instance.
func ZeroValue() *Carbon {
	return carbon.ZeroValue()
}

// EpochValue returns the unix epoch value of Carbon instance.
func EpochValue() *Carbon {
	return carbon.EpochValue()
}

// MaxValue returns the maximum value of Carbon instance.
func MaxValue() *Carbon {
	return carbon.MaxValue()
}

// MinValue returns the minimum value of Carbon instance.
func MinValue() *Carbon {
	return carbon.MinValue()
}

// MaxDuration returns the maximum value of duration instance.
func MaxDuration() Duration {
	return carbon.MaxDuration()
}

// MinDuration returns the minimum value of duration instance.
func MinDuration() Duration {
	return carbon.MinDuration()
}

// Max returns the maximum Carbon instance from some given Carbon instances.
func Max(c1 *Carbon, c2 ...*Carbon) (c *Carbon) {
	return carbon.Max(c1, c2...)
}

// Min returns the minimum Carbon instance from some given Carbon instances.
func Min(c1 *Carbon, c2 ...*Carbon) (c *Carbon) {
	return carbon.Min(c1, c2...)
}
