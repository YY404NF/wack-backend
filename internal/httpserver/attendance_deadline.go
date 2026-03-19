package httpserver

import (
	"errors"
	"time"

	"wack-backend/internal/model"
)

func attendanceSessionDate(startDate string, lesson model.CourseGroupLesson, location *time.Location) (time.Time, error) {
	if startDate == "" {
		return time.Time{}, errors.New("missing current term start date")
	}
	termStart, err := time.ParseInLocation("2006-01-02", startDate, location)
	if err != nil {
		return time.Time{}, err
	}
	dayOffset := (lesson.WeekNo-1)*7 + (lesson.Weekday - 1)
	return termStart.AddDate(0, 0, dayOffset), nil
}

func (h *apiHandler) attendanceWindow(lesson model.CourseGroupLesson, now time.Time) (time.Time, time.Time, error) {
	setting, err := h.systemSettings.GetSystemSetting()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sessionDate, err := attendanceSessionDate(setting.CurrentTermStartDate, lesson, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	clockRange, err := sectionClockRangeWithSchedule(lesson.Section, currentScheduleName(setting))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), clockRange.startHour, clockRange.startMinute, 0, 0, now.Location()).Add(-attendanceEntryLeadTime)
	end := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), clockRange.endHour, clockRange.endMinute, 0, 0, now.Location()).Add(attendanceEntryGraceTime)
	return start, end, nil
}

func (h *apiHandler) attendanceDeadline(lesson model.CourseGroupLesson, now time.Time) (time.Time, error) {
	_, end, err := h.attendanceWindow(lesson, now)
	if err != nil {
		return time.Time{}, err
	}
	return end, nil
}

func (h *apiHandler) withinDeadline(lesson model.CourseGroupLesson, now time.Time) bool {
	start, deadline, err := h.attendanceWindow(lesson, now)
	if err != nil {
		return false
	}
	return !now.Before(start) && !now.After(deadline)
}
