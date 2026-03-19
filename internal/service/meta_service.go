package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type MetaService struct {
	db *gorm.DB
}

type SectionMetaItem struct {
	Section        int    `json:"section"`
	Label          string `json:"label"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	CheckStartTime string `json:"check_start_time"`
	CheckEndTime   string `json:"check_end_time"`
}

type WeekMetaItem struct {
	WeekNo    int    `json:"week_no"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsCurrent bool   `json:"is_current"`
}

func NewMetaService(db *gorm.DB) *MetaService {
	return &MetaService{db: db}
}

func (s *MetaService) ListTerms() ([]model.Term, error) {
	var terms []model.Term
	if err := s.db.Order("term_start_date DESC, id DESC").Find(&terms).Error; err != nil {
		return nil, err
	}
	return terms, nil
}

func (s *MetaService) GetTerm(termID uint64) (model.Term, error) {
	var term model.Term
	if err := s.db.First(&term, termID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.Term{}, ErrTermNotFound
		}
		return model.Term{}, err
	}
	return term, nil
}

func (s *MetaService) GetLatestTerm() (*model.Term, error) {
	var term model.Term
	if err := s.db.Order("term_start_date DESC, id DESC").First(&term).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &term, nil
}

func (s *MetaService) Sections(now time.Time) []SectionMetaItem {
	schedule := inferScheduleByDate(now)
	ranges := map[int]struct {
		label string
		start string
		end   string
	}{
		1: {label: "上午1-2节"},
		2: {label: "上午3-4节"},
		3: {label: "下午5-6节"},
		4: {label: "下午7-8节"},
		5: {label: "晚上9-10节"},
	}

	items := make([]SectionMetaItem, 0, len(ranges))
	for section := 1; section <= 5; section++ {
		clockRange, err := sectionClockRangeWithSchedule(section, schedule)
		if err != nil {
			continue
		}
		base := ranges[section]
		items = append(items, SectionMetaItem{
			Section:        section,
			Label:          base.label,
			StartTime:      formatClock(clockRange.startHour, clockRange.startMinute),
			EndTime:        formatClock(clockRange.endHour, clockRange.endMinute),
			CheckStartTime: formatClockOffset(clockRange.startHour, clockRange.startMinute, -10*time.Minute),
			CheckEndTime:   formatClockOffset(clockRange.endHour, clockRange.endMinute, 15*time.Minute),
		})
	}
	return items
}

func (s *MetaService) TermWeeks(term model.Term, now time.Time) ([]WeekMetaItem, int, error) {
	start, err := time.ParseInLocation("2006-01-02", term.TermStartDate, now.Location())
	if err != nil {
		return nil, 0, err
	}

	currentWeek, ok := academicWeek(term.TermStartDate, now)
	if !ok {
		currentWeek = 0
	}

	maxWeek, err := s.maxWeekForTerm(term.Name)
	if err != nil {
		return nil, 0, err
	}
	totalWeeks := maxInt(maxWeek, currentWeek, 20)
	if totalWeeks <= 0 {
		totalWeeks = 20
	}

	weeks := make([]WeekMetaItem, 0, totalWeeks)
	for weekNo := 1; weekNo <= totalWeeks; weekNo++ {
		weekStart := start.AddDate(0, 0, (weekNo-1)*7)
		weekEnd := weekStart.AddDate(0, 0, 6)
		weeks = append(weeks, WeekMetaItem{
			WeekNo:    weekNo,
			StartDate: weekStart.Format("2006-01-02"),
			EndDate:   weekEnd.Format("2006-01-02"),
			IsCurrent: currentWeek > 0 && weekNo == currentWeek,
		})
	}

	return weeks, currentWeek, nil
}

func (s *MetaService) maxWeekForTerm(termName string) (int, error) {
	type result struct {
		MaxWeek int
	}

	var row result
	err := s.db.Table("course_group_lesson").
		Select("COALESCE(MAX(course_group_lesson.week_no), 0) AS max_week").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN term ON term.id = course_group.term_id").
		Where("course_group_lesson.status = 1 AND course_group.status = 1 AND term.name = ?", termName).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.MaxWeek, nil
}

func formatClock(hour, minute int) string {
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func formatClockOffset(hour, minute int, offset time.Duration) string {
	base := time.Date(2000, time.January, 1, hour, minute, 0, 0, time.Local).Add(offset)
	return base.Format("15:04")
}

func maxInt(values ...int) int {
	best := 0
	for _, value := range values {
		if value > best {
			best = value
		}
	}
	return best
}
