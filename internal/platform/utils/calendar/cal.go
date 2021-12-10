package calendar

import (
	"github.com/go-pg/pg/v9"
	rp "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/repository"
	m "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/models"
	"time"

	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// HolidayFn calculates the occurrence of a holiday for the given year.
// This is useful for holidays like Easter that depend on complex rules.
type HolidayFn func(year int, loc *time.Location) (month time.Month, day int)

// Holiday : struct for Holiday VN
type Holiday struct {
	Month   time.Month   `json:"month"`
	Weekday time.Weekday `json:"weekday"`
	Day     int          `json:"day"`
	Offset  int          `json:"offset"`
	Year    int          `json:"year"`
	Func    HolidayFn    `json:"func"`

	// last values used to calculate month and day with Func
	lastYear int
	lastLoc  *time.Location
}

// GetHolidays : get holiday from json file
func GetHolidays(holidayDates []time.Time) []Holiday {
	var holidays []Holiday
	for _, hldDate := range holidayDates {
		year, month, day := hldDate.Year(), hldDate.Month(), hldDate.Day()
		holiday := Holiday{
			Month: month,
			Day:   day,
			Year:  year,
		}

		holidays = append(holidays, holiday)
	}

	return holidays
}

func (h *Holiday) matches(date time.Time) bool {
	if h.Func != nil && (date.Year() != h.lastYear || date.Location() != h.lastLoc) {
		h.Month, h.Day = h.Func(date.Year(), date.Location())
		h.lastYear = date.Year()
		h.lastLoc = date.Location()
	}

	if h.Month > 0 {
		if date.Month() != h.Month {
			return false
		}
		if h.Year > 0 && date.Year() != h.Year {
			return false
		}
		if h.Day > 0 {
			return date.Day() == h.Day
		}
		if h.Weekday > 0 && h.Offset != 0 {
			return IsWeekdayN(date, h.Weekday, h.Offset)
		}
	} else if h.Offset > 0 {
		return date.YearDay() == h.Offset
	}
	return false
}

// Calendar : struct for calendar
type Calendar struct {
	holidays [13][]Holiday // 0 for offset based holidays, 1-12 for month based
}

// NewCalendar creates a new Calendar with no holidays defined
// and work days of Monday through Friday.
func NewCalendar(holidayDates []time.Time) *Calendar {
	c := &Calendar{}
	for i := range c.holidays {
		c.holidays[i] = make([]Holiday, 0, 2)
	}

	holidaysVN := GetHolidays(holidayDates)
	for _, hld := range holidaysVN {
		c.AddHoliday(hld)
	}

	return c
}

// AddHoliday adds a holiday to the calendar's list.
func (c *Calendar) AddHoliday(h ...Holiday) {
	for _, hd := range h {
		c.holidays[hd.Month] = append(c.holidays[hd.Month], hd)
	}
}

// IsHoliday reports whether a given date is a holiday. It does not account
// for the observation of holidays on alternate days.
func (c *Calendar) IsHoliday(date time.Time) bool {
	idx := date.Month()
	for i := range c.holidays[idx] {
		if c.holidays[idx][i].matches(date) {
			return true
		}
	}
	for i := range c.holidays[0] {
		if c.holidays[0][i].matches(date) {
			return true
		}
	}
	return false
}

// IsWeekend reports whether the given date falls on a weekend.
func IsWeekend(date time.Time) bool {
	day := date.Weekday()
	return day == time.Saturday || day == time.Sunday
}

func IsSaturday(date time.Time) bool {
	day := date.Weekday()
	return day == time.Saturday
}

func IsSunday(date time.Time) bool {
	day := date.Weekday()
	return day == time.Sunday
}

// IsWeekdayN reports whether the given date is the nth occurrence of the
// day in the month.
// The value of n affects the direction of counting:
//   n > 0: counting begins at the first day of the month.
//   n == 0: the result is always false.
//   n < 0: counting begins at the end of the month.
func IsWeekdayN(date time.Time, day time.Weekday, n int) bool {
	cday := date.Weekday()
	if cday != day || n == 0 {
		return false
	}

	if n > 0 {
		return (date.Day()-1)/7 == (n - 1)
	}

	n = -n
	last := time.Date(date.Year(), date.Month()+1,
		1, 12, 0, 0, 0, date.Location())
	lastCount := 0
	for {
		last = last.AddDate(0, 0, -1)
		if last.Weekday() == day {
			lastCount++
		}
		if lastCount == n || last.Month() != date.Month() {
			break
		}
	}
	return lastCount == n && last.Month() == date.Month() &&
		last.Day() == date.Day()

}

// ParseTime : Parse string to timestamp or Date
// Params    : layout, value
// Returns   : Date or Time
func ParseTime(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}

	return t
}

// isBetweenTime : check between start and end
func isBetweenTime(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

// CalculateHour : Calculate hour
// Params        : leaveRequestType, from, to
// Return        : hour
func CalculateHour(
	organizationId int,
	holidayRepo rp.HolidayRepository,
	leaveRequestType int,
	t1 time.Time,
	t2 time.Time,
	subtractDayOffType int,
	extraTime float64,
) float64 {
	var hour float64
	lunchBreakStart := ParseTime(cf.FormatDateNoSec, t2.Format(cf.FormatDateDatabase)+" "+cf.BreakLunchStart)
	lunchBreakEnd := ParseTime(cf.FormatDateNoSec, t2.Format(cf.FormatDateDatabase)+" "+cf.BreakLunchEnd)
	c := getCalendar(organizationId, holidayRepo, t1.Year())

	switch leaveRequestType {
	case cf.FullDayOff:
		if subtractDayOffType == cf.Other {
			hour = 0
			break
		}
		if !utils.CompareEqualDate(t2, t1) {
			count := 0
			var dates []time.Time

			diff := int(t2.Sub(t1).Hours()/24) + 1
			for i := 0; i < diff; i++ {
				dates = append(dates, t1.AddDate(0, 0, i))
			}

			for _, date := range dates {
				cld := getCalendar(organizationId, holidayRepo, date.Year())
				if !cld.IsHoliday(date) && !IsWeekend(date) {
					count++
				}
			}

			hour = float64(count * 8)
			break
		}

		if c.IsHoliday(t1) || IsWeekend(t1) {
			hour = 0
			break
		}
		hour = 8
	case cf.MorningOff, cf.AfternoonOff:
		if subtractDayOffType == cf.Other {
			hour = 0
			break
		}
		if !c.IsHoliday(t1) && !IsWeekend(t1) {
			hour = 4
			break
		}

		hour = 0
	case cf.LateForWork:
		if !c.IsHoliday(t1) && !IsWeekend(t1) {
			if subtractDayOffType == cf.Subtract {
				if lunchBreakStart.Sub(t2) >= 0 {
					hour = t2.Sub(t1).Hours()
					break
				} else if isBetweenTime(lunchBreakStart, lunchBreakEnd, t2) {
					hour = lunchBreakStart.Sub(t1).Hours()
					break
				} else {
					hour = lunchBreakStart.Sub(t1).Hours() + t2.Sub(lunchBreakEnd).Hours()
					break
				}
			} else if subtractDayOffType == cf.ExtraWork {
				if lunchBreakStart.Sub(t2) >= 0 {
					hour = t2.Sub(t1).Hours()
				} else if isBetweenTime(lunchBreakStart, lunchBreakEnd, t2) {
					hour = lunchBreakStart.Sub(t1).Hours()
				} else {
					hour = lunchBreakStart.Sub(t1).Hours() + t2.Sub(lunchBreakEnd).Hours()
				}

				hour -= extraTime
				break
			} else if subtractDayOffType == cf.Other {
				hour = 0
				break
			}
		}

		hour = 0
	case cf.LeaveEarly:
		if !c.IsHoliday(t1) && !IsWeekend(t1) {
			if subtractDayOffType == cf.Other {
				hour = 0
				break
			}
			if t1.Sub(lunchBreakEnd) >= 0 {
				hour = t2.Sub(t1).Hours()
				break
			} else if isBetweenTime(lunchBreakStart, lunchBreakEnd, t1) {
				hour = t2.Sub(lunchBreakEnd).Hours()
				break
			} else {
				hour = lunchBreakStart.Sub(t1).Hours() + t2.Sub(lunchBreakEnd).Hours()
				break
			}
		}

		hour = 0
	case cf.GoOutside:
		if !c.IsHoliday(t1) && !IsWeekend(t1) {
			if subtractDayOffType == cf.Subtract {
				hour = calculateHourGoOutSide(lunchBreakStart, lunchBreakEnd, t1, t2, cf.NotWorkAtNoon)
				break
			} else if subtractDayOffType == cf.ExtraWork {
				hour = calculateHourGoOutSide(lunchBreakStart, lunchBreakEnd, t1, t2, cf.NotWorkAtNoon) - extraTime
				break
			} else if subtractDayOffType == cf.Other {
				hour = 0
				break
			}
		}

		hour = 0
	default:
		hour = 0
	}

	return hour
}

func CalculateHourBonusOvertime(
	c *Calendar,
	from time.Time,
	to time.Time,
	overtimeWeight m.OvertimeWeight,
	workAtNoon int,
) (float64, float64, float64) {
	lunchBreakStart := ParseTime(cf.FormatDateNoSec, to.Format(cf.FormatDateDatabase)+" "+cf.BreakLunchStart)
	lunchBreakEnd := ParseTime(cf.FormatDateNoSec, to.Format(cf.FormatDateDatabase)+" "+cf.BreakLunchEnd)

	hour := calculateHourGoOutSide(lunchBreakStart, lunchBreakEnd, from, to, workAtNoon)
	if c.IsHoliday(to) {
		return overtimeWeight.HolidayWeight * hour, hour, overtimeWeight.HolidayWeight
	}

	if IsWeekend(to) {
		return overtimeWeight.WeekendWeight * hour, hour, overtimeWeight.WeekendWeight
	}

	return overtimeWeight.NormalDayWeight * hour, hour, overtimeWeight.NormalDayWeight
}

func calculateHourGoOutSide(lunchBreakStart time.Time, lunchBreakEnd time.Time, from time.Time, to time.Time, workAtNoon int) float64 {
	var hour float64
	if workAtNoon == cf.WorkAtNoon {
		hour = to.Sub(from).Hours()
	} else {
		diffFromHour := lunchBreakStart.Sub(from).Hours()
		diffToHour := to.Sub(lunchBreakEnd).Hours()

		if lunchBreakStart.Sub(to) >= 0 || from.Sub(lunchBreakEnd) >= 0 {
			hour = to.Sub(from).Hours()
		} else if diffFromHour >= 0 && isBetweenTime(lunchBreakStart, lunchBreakEnd, to) {
			hour = diffFromHour
		} else if diffToHour >= 0 && isBetweenTime(lunchBreakStart, lunchBreakEnd, from) {
			hour = diffToHour
		} else if diffFromHour >= 0 && diffToHour >= 0 {
			hour = diffToHour + diffFromHour
		} else {
			hour = 0
		}
	}

	return hour
}

func getCalendar(organizationId int, holidayRepo rp.HolidayRepository, year int) *Calendar {
	records, err := holidayRepo.SelectHolidays(organizationId, year, "holiday_date")
	if err != nil && err.Error() != pg.ErrNoRows.Error() {
		panic(err)
	}

	var holidayDates []time.Time
	for _, record := range records {
		holidayDates = append(holidayDates, record.HolidayDate)
	}
	c := NewCalendar(holidayDates)

	return c
}
