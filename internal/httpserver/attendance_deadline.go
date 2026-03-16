package httpserver

import (
	"errors"
	"time"

	"wack-backend/internal/model"
)

func sectionEndTime(section int) (int, int, error) {
	switch section {
	case 1:
		return 9, 35, nil
	case 2:
		return 11, 35, nil
	case 3:
		return 15, 35, nil
	case 4:
		return 17, 35, nil
	case 5:
		return 21, 35, nil
	default:
		return 0, 0, errors.New("invalid section")
	}
}

func (h *apiHandler) attendanceDeadline(session model.CourseSession, now time.Time) (time.Time, error) {
	currentWeekday := int(now.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7
	}
	weekdayDelta := session.Weekday - currentWeekday
	if weekdayDelta < 0 {
		weekdayDelta += 7
	}
	baseDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, weekdayDelta)
	hour, minute, err := sectionEndTime(session.Section)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, minute, 0, 0, now.Location()).Add(30 * time.Minute), nil
}

func (h *apiHandler) withinDeadline(session model.CourseSession, now time.Time) bool {
	deadline, err := h.attendanceDeadline(session, now)
	if err != nil {
		return false
	}
	return now.Before(deadline)
}
