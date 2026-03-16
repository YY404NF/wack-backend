package httpserver

import (
	"errors"
	"time"

	"wack-backend/internal/model"
)

const (
	attendanceEntryLeadTime  = 10 * time.Minute
	attendanceEntryGraceTime = 15 * time.Minute
)

type sectionClockRange struct {
	startHour   int
	startMinute int
	endHour     int
	endMinute   int
}

var sectionClockRanges = map[string]map[int]sectionClockRange{
	"summer": {
		1: {startHour: 8, startMinute: 0, endHour: 9, endMinute: 40},
		2: {startHour: 9, startMinute: 55, endHour: 11, endMinute: 35},
		3: {startHour: 14, startMinute: 30, endHour: 16, endMinute: 10},
		4: {startHour: 16, startMinute: 25, endHour: 18, endMinute: 5},
		5: {startHour: 19, startMinute: 0, endHour: 20, endMinute: 40},
	},
	"autumn": {
		1: {startHour: 8, startMinute: 0, endHour: 9, endMinute: 40},
		2: {startHour: 9, startMinute: 55, endHour: 11, endMinute: 35},
		3: {startHour: 14, startMinute: 0, endHour: 15, endMinute: 40},
		4: {startHour: 15, startMinute: 55, endHour: 17, endMinute: 35},
		5: {startHour: 19, startMinute: 0, endHour: 20, endMinute: 40},
	},
}

func academicWeek(startDate string, now time.Time) (int, bool) {
	if startDate == "" {
		return 0, false
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, now.Location())
	if err != nil {
		return 0, false
	}
	diff := now.Sub(start)
	week := int(diff.Hours()/(24*7)) + 1
	if week < 1 {
		week = 1
	}
	return week, true
}

func currentScheduleName(setting model.SystemSetting) string {
	if setting.CurrentSchedule == "summer" {
		return "summer"
	}
	return "autumn"
}

func sectionVisibleNow(section int, schedule string, now time.Time) bool {
	target, err := sectionClockRangeWithSchedule(section, schedule)
	if err != nil {
		return false
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), target.startHour, target.startMinute, 0, 0, now.Location()).Add(-attendanceEntryLeadTime)
	end := time.Date(now.Year(), now.Month(), now.Day(), target.endHour, target.endMinute, 0, 0, now.Location()).Add(attendanceEntryGraceTime)
	return !now.Before(start) && !now.After(end)
}

func sectionClockRangeWithSchedule(section int, schedule string) (sectionClockRange, error) {
	rangeBySection, ok := sectionClockRanges[schedule]
	if !ok {
		return sectionClockRange{}, errors.New("invalid schedule")
	}
	target, ok := rangeBySection[section]
	if !ok {
		return sectionClockRange{}, errors.New("invalid section")
	}
	return target, nil
}

func sectionEndTimeWithSchedule(section int, schedule string) (int, int, error) {
	target, err := sectionClockRangeWithSchedule(section, schedule)
	if err != nil {
		return 0, 0, err
	}
	return target.endHour, target.endMinute, nil
}
