package timex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/jinzhu/now"
)

const (
	Day                    = "day"
	Week                   = "week"
	Month                  = "month"
	Year                   = "year"
	Custom                 = "custom"
	PeriodPrevious         = "previous"
	PeriodCurrent          = "current"
	TimeZoneAsiaBangkok    = "Asia/Bangkok"
	TimeFormatSlash        = "2006/01/02 15:04:05"
	TimeFormatDash1        = "2006-01-02 15:04:05"
	TimeFormatDash2        = "2006-01-02 15:04"
	DateFormatSlash1       = "2006/01/02"
	DateFormatSlash2       = "02/01/2022"
	DateFormatDash         = "2006-01-02"
	DateFormat             = "20060102"
	DateFormatTime         = "15:04:05"
	DateTimeFormatISO      = "2006-01-02T15:04:05.000Z"
	DateTimeFormatISOShort = "2006-01-02T15:04:05Z"
	DateTimeFormatISO2     = "2012-03-29T10:05:45-06:00"
)

type TimeCurrent struct {
	Start time.Time
	End   time.Time
}

type TimePrevious struct {
	Start time.Time
	End   time.Time
}

type TimeRange struct {
	Start          string
	Stop           string
	StartTimestamp int64
	StopTimestamp  int64
}

type GenRange struct {
	Unit         string
	Every        string
	EveryNum     int
	PrevDuration time.Duration
	Num          int
	NumRange     int
	TimeRange    []TimeRange
}

func IsWorkdayBy(date string, layout string) (bool, error) {
	t, err := ParseBy(date, layout)
	if err != nil {
		return false, err
	}

	return IsWorkday(t), nil
}

func IsWorkday(date time.Time) bool {
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	return true
}

func GetWeekday(date time.Time) string {
	return date.Weekday().String()
}

// ParseStartOfDay date string 2021-09-09 to time.Time
func ParseStartOfDay(date string) (time.Time, error) {
	startTime := fmt.Sprintf("%sT00:00:00.000Z", date)
	return Parse(startTime)
}

// ParseEndOfDay date string 2021-09-09 to time.Time
func ParseEndOfDay(date string) (time.Time, error) {
	endTime := fmt.Sprintf("%sT23:59:59.999Z", date)
	return Parse(endTime)
}

// Parse date string 2021-09-09T00:00:00.000Z to time.Time
func Parse(date string) (time.Time, error) {
	return time.Parse(DateTimeFormatISO, date)
}

func ParseBy(date string, layout string) (time.Time, error) {
	return time.Parse(layout, date)
}

func SupportUnit(data string) bool {
	result, _ := regexp.MatchString("(day|month|year)+", data)
	return result
}

func SupportUnitPrevious(data string) bool {
	result, _ := regexp.MatchString("(day|month|year)-\\d+", data)
	return result
}

func SupportTimeSeries(data string) bool {
	result, _ := regexp.MatchString("(\\d+s|\\d+m|\\d+h|\\d+d|\\d+w|\\d+mo|\\d+y|custom)", data)
	return result
}

func GetTimeZone(zone string) *time.Location {
	tz, err := time.LoadLocation(zone)
	timeZone := time.Local
	if err == nil {
		timeZone = tz
	}
	return timeZone
}

func TimeNowFormat(zone string, format string) string {
	timeZone := GetTimeZone(zone)
	return time.Now().In(timeZone).Format(format)
}

func Previous(filter string) (int, time.Month, int) {
	switch filter {
	case Day:
		return Now().Add(-24 * time.Hour).Date()
	case Month:
		y, m, d := Now().Date()
		month := m - 1
		if month == 0 {
			month = 12
			y -= 1
		}
		return y, month, d
	case Year:
		y, m, d := Now().Date()
		return y - 1, m, d
	}
	return Date()
}

func PrevDay(num int) time.Time {
	return Now().AddDate(0, 0, -num)
}

func PrevMonth(num int) time.Time {
	t := Now()
	return SubMonth(t, num)
}

func SubMonth(t time.Time, num int) time.Time {
	sub := time.Date(t.Year(), (t.Month()+1)-time.Month(num), 0, 0, 00, 00, 00, t.Location())
	return sub
}

func PrevYear(num int) time.Time {
	return Now().AddDate(-num, 0, 0)
}

func GetFebruaryLastOfMonth() {
	month := 2
	feb := time.Date(2016, time.Month(month+1), 0, 0, 0, 0, 0, time.Local)
	fmt.Println(feb.Day()) // 29 days
}

func Now() time.Time {
	timeZone := GetTimeZone(TimeZoneAsiaBangkok)
	return time.Now().In(timeZone)
}

func Date() (int, time.Month, int) {
	return Now().Date()
}

func TimeNow() *now.Now {
	return now.With(Now())
}

func FromTimestampGMT7(timestamp int64) time.Time {
	return FromTimestamp(timestamp, GetTimeZone(TimeZoneAsiaBangkok))
}

func FromTimestamp(timestamp int64, tz *time.Location) time.Time {
	return time.Unix(timestamp, 0).In(tz)
}

func Format(year int, m time.Month, d int) string {
	month := fmt.Sprintf("%d", m)
	day := fmt.Sprintf("%d", d)
	if m < 9 {
		month = fmt.Sprintf("0%d", m)
	}
	if d < 9 {
		day = fmt.Sprintf("0%d", d)
	}
	return fmt.Sprintf("%d-%s-%s", year, month, day)
}

func ToTimeCurrent(filter string) TimeCurrent {
	current := now.With(Now())
	switch filter {
	case Day:
		return TimeCurrent{
			Start: current.BeginningOfDay(),
			End:   current.EndOfDay(),
		}
	case Month:
		return TimeCurrent{
			Start: current.BeginningOfMonth(),
			End:   current.EndOfMonth(),
		}
	case Year:
		return TimeCurrent{
			Start: current.BeginningOfYear(),
			End:   current.EndOfYear(),
		}
	}
	return TimeCurrent{
		Start: current.BeginningOfDay(),
		End:   current.EndOfDay(),
	}
}

func ToTimePrevious(filter string) TimePrevious {
	t, e := now.Parse(Format(Previous(filter)))
	if e != nil {
		t = Now()
	}
	previous := now.With(t)
	switch filter {
	case Day:
		return TimePrevious{
			Start: previous.BeginningOfDay(),
			End:   previous.EndOfDay(),
		}
	case Month:
		return TimePrevious{
			Start: previous.BeginningOfMonth(),
			End:   previous.EndOfMonth(),
		}
	case Year:
		return TimePrevious{
			Start: previous.BeginningOfYear(),
			End:   previous.EndOfYear(),
		}
	}

	current := now.With(Now())
	return TimePrevious{
		Start: current.BeginningOfDay(),
		End:   current.EndOfDay(),
	}
}

// GenerateMinuteRange support every 1m
// How to use:
// timeRange := timex.GenerateMinuteRange(5, 60, 0)
func GenerateMinuteRange(num int, secondOfMinute int, customCurrentTimeUnix int64) GenRange {
	data := []TimeRange{}
	numSecond := num * secondOfMinute
	var sec time.Duration
	var dt time.Time
	// Custom current time
	currentTime := Now()
	if customCurrentTimeUnix > 0 {
		currentTime = time.Unix(customCurrentTimeUnix, 0)
	}
	everyNum := 60 // seconds
	every := "1m"
	for s := numSecond; s >= 1; s -= everyNum {
		sec = time.Duration(s)
		if s == 1 {
			dt = currentTime
		} else {
			dt = currentTime.Add(-(time.Second) * (sec - time.Duration(everyNum)))
		}
		current := dt
		previous := current.Add(-(time.Second) * time.Duration(everyNum))
		prev := now.With(previous)
		start := prev.BeginningOfMinute().Format(DateTimeFormatISO)
		stop := current.Format(DateTimeFormatISO)
		data = append(data, TimeRange{
			Start:          start,
			Stop:           stop,
			StartTimestamp: prev.BeginningOfMinute().Unix(),
			StopTimestamp:  current.Unix(),
		})
	}
	return GenRange{
		Unit:         "m",
		Every:        every,
		EveryNum:     everyNum,
		PrevDuration: -(time.Second),
		Num:          num,
		NumRange:     numSecond,
		TimeRange:    data,
	}
}

// GenerateHourRange support every 2m, 15m, 30m
// How to use:
// timeRange := timex.GenerateHourRange(1, 60, 0)
func GenerateHourRange(num int, minuteOfHour int, customCurrentTimeUnix int64) GenRange {
	data := []TimeRange{}
	numMinute := num * minuteOfHour
	var mint time.Duration
	var dt time.Time
	// Custom current time
	currentTime := Now()
	if customCurrentTimeUnix > 0 {
		currentTime = time.Unix(customCurrentTimeUnix, 0)
	}
	everyNum := 1 // minute
	every := "1m"
	if num == 1 { // 1h  			->  30 axis
		every = "2m"
		everyNum = 2 // minute
	} else if num == 6 { // 6h  	->  24 axis
		every = "15m"
		everyNum = 15 // minute
	} else if num == 12 { // 12h  	->  24 axis
		every = "30m"
		everyNum = 30 // minute
	} else if num == 24 { // 1d  	->  24 axis
		every = "1h"
		everyNum = 60 // 1hour
	} else if num == 48 { // 2d  	->  24 axis
		every = "2h"
		everyNum = 120 // 2hour
	}
	for m := numMinute; m >= 1; m -= everyNum {
		mint = time.Duration(m)
		if m == 1 {
			dt = currentTime
		} else {
			dt = currentTime.Add(-(time.Minute) * (mint - time.Duration(everyNum)))
		}
		prev := now.With(dt)
		start := prev.BeginningOfMinute().Format(DateTimeFormatISO)
		stop := prev.EndOfMinute().Format(DateTimeFormatISO)
		data = append(data, TimeRange{
			Start:          start,
			Stop:           stop,
			StartTimestamp: prev.BeginningOfMinute().Unix(),
			StopTimestamp:  prev.EndOfMinute().Unix(),
		})
	}
	return GenRange{
		Unit:         "h",
		Every:        every,
		EveryNum:     everyNum,
		PrevDuration: -(time.Minute),
		Num:          num,
		NumRange:     numMinute,
		TimeRange:    data,
	}
}

// GenerateDayRange support every 1h
// How to use:
// timeRange := timex.GenerateDayRange(1, 24, 0)
func GenerateDayRange(num int, hourOfDay int, customCurrentTimeUnix int64) GenRange {
	data := []TimeRange{}
	hourNum := hourOfDay * num
	var hh time.Duration
	var dt time.Time
	// Custom current time
	currentTime := Now()
	if customCurrentTimeUnix > 0 {
		currentTime = time.Unix(customCurrentTimeUnix, 0)
	}
	everyNum := 1 // hour
	every := "1h"
	if num >= 2 {
		every = "24h"
		everyNum = 24 // hour
	}
	for h := hourNum; h >= 1; h -= everyNum {
		hh = time.Duration(h)
		if h == 1 {
			dt = currentTime
		} else {
			dt = currentTime.Add(-(time.Hour) * (hh - time.Duration(everyNum)))
		}
		prev := now.With(dt)
		// End of hour
		start := prev.BeginningOfHour().Format(DateTimeFormatISO)
		stop := prev.EndOfHour().Format(DateTimeFormatISO)
		startTimestamp := prev.BeginningOfHour().Unix()
		stopTimestamp := prev.EndOfHour().Unix()
		// End of day
		if num >= 2 {
			start = prev.BeginningOfDay().Format(DateTimeFormatISO)
			stop = prev.EndOfDay().Format(DateTimeFormatISO)
			startTimestamp = prev.BeginningOfDay().Unix()
			stopTimestamp = prev.EndOfDay().Unix()
		}
		data = append(data, TimeRange{
			Start:          start,
			Stop:           stop,
			StartTimestamp: startTimestamp,
			StopTimestamp:  stopTimestamp,
		})
	}
	return GenRange{
		Unit:         "d",
		Every:        every,
		EveryNum:     everyNum,
		PrevDuration: -(time.Hour),
		Num:          num,
		NumRange:     hourNum,
		TimeRange:    data,
	}
}

// GenerateMonthRange support every 1d
// How to use:
// timeRange := timex.GenerateMonthRange(1, 30, 0)
func GenerateMonthRange(num int, daysOfMonth int, customCurrentTimeUnix int64) GenRange {
	data := []TimeRange{}
	numMonth := daysOfMonth * num
	var day time.Duration
	var dt time.Time
	// Custom current time
	currentTime := Now()
	if customCurrentTimeUnix > 0 {
		currentTime = time.Unix(customCurrentTimeUnix, 0)
	}
	everyNum := 1 // day
	every := "1d"
	for d := numMonth; d >= 1; d-- {
		day = time.Duration(d)
		if d == 1 {
			dt = currentTime
		} else {
			dt = currentTime.Add(-(24 * time.Hour) * (day - time.Duration(everyNum)))
		}
		prev := now.With(dt)
		start := prev.BeginningOfDay().Format(DateTimeFormatISO)
		stop := prev.EndOfDay().Format(DateTimeFormatISO)
		data = append(data, TimeRange{
			Start:          start,
			Stop:           stop,
			StartTimestamp: prev.BeginningOfDay().Unix(),
			StopTimestamp:  prev.EndOfDay().Unix(),
		})
	}
	return GenRange{
		Unit:         "mo",
		Every:        every,
		EveryNum:     everyNum,
		PrevDuration: -(24 * time.Hour),
		Num:          num,
		NumRange:     numMonth,
		TimeRange:    data,
	}
}

// GenerateYearRange support every 1mo
// How to use:
// timeRange := timex.GenerateYearRange(1, 12, 0)
func GenerateYearRange(num int, monthOfYear int, customCurrentTimeUnix int64) GenRange {
	data := []TimeRange{}
	numYear := monthOfYear * num
	var dt time.Time
	// Custom current time
	currentTime := Now()
	if customCurrentTimeUnix > 0 {
		currentTime = time.Unix(customCurrentTimeUnix, 0)
	}
	everyNum := 1 // month
	every := "1mo"
	for month := numYear; month >= 1; month-- {
		if month == 1 {
			dt = currentTime
		} else {
			dt = currentTime.AddDate(0, -(month - everyNum), 0)
		}
		prev := now.With(dt)
		start := prev.BeginningOfMonth().Format(DateTimeFormatISO)
		stop := prev.EndOfMonth().Format(DateTimeFormatISO)
		data = append(data, TimeRange{
			Start:          start,
			Stop:           stop,
			StartTimestamp: prev.BeginningOfMonth().Unix(),
			StopTimestamp:  prev.EndOfMonth().Unix(),
		})
	}
	return GenRange{
		Unit:      "y",
		Every:     every,
		EveryNum:  everyNum,
		Num:       num,
		NumRange:  numYear,
		TimeRange: data,
	}
}

func NextDay(day int) time.Time {
	return Now().Add((24 * time.Hour) * time.Duration(day))
}

func NextDayRange(day int) TimeRange {
	prev := now.With(NextDay(day))
	start := prev.BeginningOfDay().Format(DateTimeFormatISO)
	stop := prev.EndOfDay().Format(DateTimeFormatISO)
	return TimeRange{
		Start:          start,
		Stop:           stop,
		StartTimestamp: prev.BeginningOfDay().Unix(),
		StopTimestamp:  prev.EndOfDay().Unix(),
	}
}

func GetYearByTimeSeries(ts string) int {
	nm := strings.Split(ts, "y")
	if len(nm) > 0 {
		month, err := strconv.Atoi(nm[0])
		if err != nil {
			return 0
		}
		return month
	}
	return 0
}

func GetMonthByTimeSeries(ts string) int {
	nm := strings.Split(ts, "mo")
	if len(nm) > 0 {
		month, err := strconv.Atoi(nm[0])
		if err != nil {
			return 0
		}
		return month
	}
	return 0
}

func GetDayByTimeSeries(ts string) int {
	// days case
	nd := strings.Split(ts, "d")
	if len(nd) > 0 {
		day, err := strconv.Atoi(nd[0])
		if err != nil {
			return 0
		}
		return day
	}

	// hours case
	nh := strings.Split(ts, "h")
	if len(nh) > 0 {
		hour, err := strconv.Atoi(nh[0])
		if err != nil {
			return 0
		}
		return hour / 24 // convert to day
	}
	return 0
}

func GetHourByTimeSeries(ts string) int {
	nh := strings.Split(ts, "h")
	if len(nh) > 0 {
		hour, err := strconv.Atoi(nh[0])
		if err != nil {
			return 0
		}
		return hour
	}
	return 0
}

func GetMinuteByTimeSeries(ts string) int {
	nm := strings.Split(ts, "m")
	if len(nm) > 0 {
		month, err := strconv.Atoi(nm[0])
		if err != nil {
			return 0
		}
		return month
	}
	return 0
}

func IsHourTimeSeries(ts string) bool {
	match1, _ := regexp.MatchString("\\d+h$", ts)
	return match1
}

func IsMinuteTimeSeries(ts string) bool {
	match1, _ := regexp.MatchString("\\d+m$", ts)
	return match1
}

func IsDayTimeSeries(ts string) bool {
	match1, _ := regexp.MatchString("\\d+d$", ts)
	match2, _ := regexp.MatchString("\\d+h$", ts)
	return match1 || match2
}

func IsMonthTimeSeries(ts string) bool {
	match, _ := regexp.MatchString("\\d+mo$", ts)
	return match
}

func IsYearTimeSeries(ts string) bool {
	match, _ := regexp.MatchString("\\d+y$", ts)
	return match
}

func GetTimeSeriesByStartEnd(startTime time.Time, endTime time.Time) interface{} {
	diffTime := endTime.Sub(startTime)
	fmt.Println(diffTime)
	return diffTime
}

func Subtract(num int) {
	n := time.Now()
	fmt.Println("Today:", n)

	after := n.AddDate(-1, 0, 0)
	fmt.Println("Subtract 1 Year:", after)

	after = n.AddDate(0, -1, 0)
	fmt.Println("Subtract 1 Month:", after)

	after = n.AddDate(0, 0, -1)
	fmt.Println("Subtract 1 Day:", after)

	after = n.AddDate(-2, -2, -5)
	fmt.Println("Subtract multiple values:", after)

	after = n.Add(-10 * time.Minute)
	fmt.Println("Subtract 10 Minutes:", after)

	after = n.Add(-10 * time.Second)
	fmt.Println("Subtract 10 Second:", after)

	after = n.Add(-10 * time.Hour)
	fmt.Println("Subtract 10 Hour:", after)

	after = n.Add(-10 * time.Millisecond)
	fmt.Println("Subtract 10 Millisecond:", after)

	after = n.Add(-10 * time.Microsecond)
	fmt.Println("Subtract 10 Microsecond:", after)

	after = n.Add(-10 * time.Nanosecond)
	fmt.Println("Subtract 10 Nanosecond:", after)
}
