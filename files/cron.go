package files

// Shamelessly lifted from https://github.com/temporalio/temporal/blob/main/service/worker/scheduler/calendar.go
// I'd import it as a dependency, but it's not exported from the package, so this is the
// only way to use it from here.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	schedpb "go.temporal.io/api/schedule/v1"
	"go.temporal.io/server/common/primitives/timestamp"
	"google.golang.org/protobuf/types/known/durationpb"
)

type parseMode int

const (
	// minCalendarYear is the smallest year that can appear in a calendar spec.
	minCalendarYear = 2000
	// maxCalendarYear is the latest year that will be recognized for calendar dates.
	// If you're still using Temporal in 2100 please change this constant and rebuild.
	maxCalendarYear = 2100
)

const (
	// Modes for parsing range strings: all modes accept decimal integers.
	parseModeInt parseMode = iota
	// parseModeYear is like parseModeInt but returns an empty range for the default.
	parseModeYear
	// parseModeMonth also accepts month name prefixes (at least three letters).
	parseModeMonth
	// parseModeDow also accepts day-of-week prefixes (at least two letters).
	parseModeDow
)

var (
	errOutOfRange = errors.New("out of range")

	monthStrings = []string{ //nolint:gochecknoglobals
		"january",
		"february",
		"march",
		"april",
		"may",
		"june",
		"july",
		"august",
		"september",
		"october",
		"november",
		"december",
	}

	dowStrings = []string{ //nolint:gochecknoglobals
		"sunday",
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
	}
)

func parseCronString(cronStr string) (*schedpb.StructuredCalendarSpec, *schedpb.IntervalSpec, string, error) {
	var (
		tzName  string
		comment string
	)

	cronStr = strings.TrimSpace(cronStr)

	// split out timezone
	if strings.HasPrefix(cronStr, "TZ=") || strings.HasPrefix(cronStr, "CRON_TZ=") {
		tz, rest, found := strings.Cut(cronStr, " ")
		if !found {
			return nil, nil, "", errors.New("CronString has time zone but missing fields") //nolint:goerr113
		}

		cronStr = rest

		_, tzName, _ = strings.Cut(tz, "=")
	}

	// split out comment
	cronStr, comment, _ = strings.Cut(cronStr, "#")
	cronStr = strings.TrimSpace(cronStr)
	comment = strings.TrimSpace(comment)

	// handle @every intervals
	if strings.HasPrefix(cronStr, "@every") {
		iv, err := parseCronStringInterval(cronStr)

		return nil, iv, "", err
	}

	// handle @hourly, etc.
	cronStr = handlePredefinedCronStrings(cronStr)

	// split fields
	cal := schedpb.CalendarSpec{Comment: comment}
	fields := strings.Fields(cronStr)

	switch len(fields) {
	case 5: //nolint:gomnd,mnd
		cal.Minute, cal.Hour, cal.DayOfMonth, cal.Month, cal.DayOfWeek = fields[0], fields[1], fields[2], fields[3], fields[4] //nolint:lll
	case 6: //nolint:gomnd,mnd
		cal.Minute, cal.Hour, cal.DayOfMonth, cal.Month, cal.DayOfWeek, cal.Year = fields[0], fields[1], fields[2], fields[3], fields[4], fields[5] //nolint:lll
	case 7: //nolint:gomnd,mnd
		cal.Second, cal.Minute, cal.Hour, cal.DayOfMonth, cal.Month, cal.DayOfWeek, cal.Year = fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6] //nolint:lll
	default:
		return nil, nil, "", errors.New("CronString does not have 5-7 fields") //nolint:goerr113
	}

	structured, err := parseCalendarToStructured(&cal)
	if err != nil {
		return nil, nil, "", err
	}

	return structured, nil, tzName, nil
}

func parseCalendarToStructured(cal *schedpb.CalendarSpec) (*schedpb.StructuredCalendarSpec, error) {
	var errs []string

	makeRangeOrNil := func(s, field, def string, min, max int, parseMode parseMode) []*schedpb.Range {
		r, err := makeRange(s, field, def, min, max, parseMode)
		if err != nil {
			errs = append(errs, err.Error())
		}

		return r
	}

	spec := &schedpb.StructuredCalendarSpec{
		Second:     makeRangeOrNil(cal.Second, "Second", "0", 0, 59, parseModeInt),                         //nolint:protogetter,lll,gomnd,mnd
		Minute:     makeRangeOrNil(cal.Minute, "Minute", "0", 0, 59, parseModeInt),                         //nolint:protogetter,lll,gomnd,mnd
		Hour:       makeRangeOrNil(cal.Hour, "Hour", "0", 0, 23, parseModeInt),                             //nolint:protogetter,lll,gomnd,mnd
		DayOfWeek:  makeRangeOrNil(cal.DayOfWeek, "DayOfWeek", "*", 0, 7, parseModeDow),                    //nolint:protogetter,lll,gomnd,mnd
		DayOfMonth: makeRangeOrNil(cal.DayOfMonth, "DayOfMonth", "*", 1, 31, parseModeInt),                 //nolint:protogetter,lll,gomnd,mnd
		Month:      makeRangeOrNil(cal.Month, "Month", "*", 1, 12, parseModeMonth),                         //nolint:protogetter,lll,gomnd,mnd
		Year:       makeRangeOrNil(cal.Year, "Year", "*", minCalendarYear, maxCalendarYear, parseModeYear), //nolint:protogetter,lll
		Comment:    cal.Comment,                                                                            //nolint:protogetter,lll
	}

	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", ")) //nolint:goerr113
	}

	return spec, nil
}

func parseCronStringInterval(c string) (*schedpb.IntervalSpec, error) {
	// split after @every
	_, interval, found := strings.Cut(c, " ")
	if !found {
		return nil, errors.New("CronString does not have interval after @every") //nolint:goerr113
	}
	// allow @every 14h/3h
	interval, phase, _ := strings.Cut(interval, "/")

	intervalDuration, err := timestamp.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	if phase == "" {
		return &schedpb.IntervalSpec{Interval: durationpb.New(intervalDuration)}, nil
	}

	phaseDuration, err := timestamp.ParseDuration(phase)
	if err != nil {
		return nil, err
	}

	return &schedpb.IntervalSpec{Interval: durationpb.New(intervalDuration), Phase: durationpb.New(phaseDuration)}, nil
}

func handlePredefinedCronStrings(cronStr string) string {
	switch cronStr {
	case "@yearly", "@annually":
		return "0 0 1 1 *"
	case "@monthly":
		return "0 0 1 * *"
	case "@weekly":
		return "0 0 * * 0"
	case "@daily", "@midnight":
		return "0 0 * * *"
	case "@hourly":
		return "0 * * * *"
	default:
		return cronStr
	}
}

func makeRange(str, field, def string, min, max int, parseMode parseMode) ([]*schedpb.Range, error) { //nolint:funlen,gocognit,lll,cyclop
	str = strings.TrimSpace(str)
	if str == "" {
		str = def
	}

	if str == "*" && parseMode == parseModeYear {
		return nil, nil // special case for year: all is represented as empty range list
	}

	var ranges []*schedpb.Range //nolint:prealloc

	for _, part := range strings.Split(str, ",") {
		var err error

		step := 1
		hasStep := false

		if strings.Contains(part, "/") {
			skipParts := strings.Split(part, "/")
			if len(skipParts) != 2 { //nolint:gomnd,mnd
				return nil, fmt.Errorf("%s has too many slashes", field) //nolint:goerr113
			}

			part = skipParts[0]

			step, err = strconv.Atoi(skipParts[1])
			if err != nil {
				return nil, err
			}

			if step < 1 {
				return nil, fmt.Errorf("%s has invalid Step", field) //nolint:goerr113
			}

			hasStep = true
		}

		start, end := min, max

		if part != "*" { //nolint:nestif
			if strings.Contains(part, "-") {
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) != 2 { //nolint:gomnd,mnd
					return nil, fmt.Errorf("%s has too many dashes", field) //nolint:goerr113
				}

				if start, err = parseValue(rangeParts[0], min, max, parseMode); err != nil {
					return nil, fmt.Errorf("%s Start is not in range [%d-%d]", field, min, max) //nolint:goerr113
				}

				if end, err = parseValue(rangeParts[1], start, max, parseMode); err != nil {
					return nil, fmt.Errorf("%s End is before Start or not in range [%d-%d]", field, min, max) //nolint:goerr113
				}
			} else {
				if start, err = parseValue(part, min, max, parseMode); err != nil {
					return nil, fmt.Errorf("%s is not in range [%d-%d]", field, min, max) //nolint:goerr113
				}

				if !hasStep {
					// if / is present, a single value is treated as that value to the
					// end. otherwise a single value is just the single value.
					end = start
				}
			}
		}
		// Special handling for Sunday: Turn "7" into "0", which may require an extra range.
		// Consider some cases:
		// 0-7 or 1-7   can turn into 0-6
		// 3-7          has to turn into 0,3-6
		// 3-7/3        can turn into 3-6/3  (7 doesn't match)
		// 1-7/2        has to turn into 0,1-6/2
		// That is, we can use a single range and just turn the 7 into a 6 only if step == 1
		// and start == 0 or 1. Or if 7 isn't actually included. In other cases, we can add a
		// 0, and then turn the 7 into a 6 in whatever the original range was. If the original
		// range was just 7-7, then we're done.
		if parseMode == parseModeDow && end == 7 { //nolint:gomnd,mnd
			if (7-start)%step == 0 && (step > 1 || step == 1 && start > 1) { //nolint:gomnd,mnd
				ranges = append(ranges, &schedpb.Range{Start: int32(0)})

				if start == 7 { //nolint:gomnd,mnd
					continue
				}
			}

			end = 6
		}

		if start == end {
			end = 0 // use default value so proto is smaller
		}

		if step == 1 {
			step = 0 // use default value so proto is smaller
		}

		ranges = append(ranges, &schedpb.Range{Start: int32(start), End: int32(end), Step: int32(step)}) //nolint:gosec
	}

	return ranges, nil
}

func parseValue(str string, min, max int, parseMode parseMode) (int, error) { //nolint:gocognit,cyclop
	if parseMode == parseModeMonth { //nolint:nestif
		if len(str) >= 3 { //nolint:gomnd,mnd
			str = strings.ToLower(str)
			for i, month := range monthStrings {
				if strings.HasPrefix(month, str) {
					i++
					if i < min || i > max {
						return i, errOutOfRange
					}

					return i, nil
				}
			}
		}
	} else if parseMode == parseModeDow {
		if len(str) >= 2 { //nolint:gomnd,mnd
			str = strings.ToLower(str)
			for i, dow := range dowStrings {
				if strings.HasPrefix(dow, str) {
					if i < min || i > max {
						return i, errOutOfRange
					}

					return i, nil
				}
			}
		}
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		return i, err
	}

	if i < min || i > max {
		return i, errOutOfRange
	}

	return i, nil
}
