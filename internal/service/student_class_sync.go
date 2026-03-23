package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

func syncStudentClassMembership(tx *gorm.DB, studentID uint64, oldClassID, newClassID *uint64) error {
	if sameUint64PointerValue(oldClassID, newClassID) {
		return nil
	}

	term, err := resolveActiveTermForDB(tx, time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if oldClassID != nil {
		if err := removeStudentFromClassCourseGroups(tx, term.ID, studentID, *oldClassID); err != nil {
			return err
		}
	}
	if newClassID != nil {
		if err := addStudentToClassCourseGroups(tx, term.ID, studentID, *newClassID); err != nil {
			return err
		}
	}
	return nil
}

func resolveActiveTermForDB(db *gorm.DB, now time.Time) (model.Term, error) {
	var term model.Term
	today := now.Format("2006-01-02")
	err := db.
		Where("term_start_date <= ?", today).
		Order("term_start_date DESC, id DESC").
		First(&term).Error
	switch {
	case err == nil:
		return term, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return model.Term{}, err
	}

	err = db.
		Order("term_start_date ASC, id ASC").
		First(&term).Error
	if err != nil {
		return model.Term{}, err
	}
	return term, nil
}

func removeStudentFromClassCourseGroups(tx *gorm.DB, termID, studentID, classID uint64) error {
	var relations []model.CourseGroupStudent
	if err := tx.
		Where("term_id = ? AND student_id = ? AND class_id = ? AND status = 1", termID, studentID, classID).
		Find(&relations).Error; err != nil {
		return err
	}

	for _, relation := range relations {
		hasHistory, err := courseGroupMemberHasHistory(tx, relation.CourseGroupID, func(db *gorm.DB) *gorm.DB {
			return db.Where("student_id = ?", studentID)
		})
		if err != nil {
			return err
		}
		if hasHistory {
			if err := tx.Model(&model.CourseGroupStudent{}).
				Where("id = ? AND status = 1", relation.ID).
				Updates(map[string]interface{}{
					"status":     2,
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				}).Error; err != nil {
				return err
			}
			continue
		}
		if err := tx.Delete(&model.CourseGroupStudent{}, relation.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func addStudentToClassCourseGroups(tx *gorm.DB, termID, studentID, classID uint64) error {
	var groupIDs []uint64
	if err := tx.Model(&model.CourseGroupStudent{}).
		Distinct("course_group_student.course_group_id").
		Joins("JOIN course_group ON course_group.id = course_group_student.course_group_id").
		Where("course_group_student.term_id = ? AND course_group_student.class_id = ? AND course_group_student.status = 1 AND course_group.status = 1", termID, classID).
		Pluck("course_group_student.course_group_id", &groupIDs).Error; err != nil {
		return err
	}

	for _, groupID := range groupIDs {
		classRef := uint64Ptr(classID)
		if err := upsertCourseGroupStudent(tx, model.CourseGroupStudent{
			TermID:        termID,
			CourseGroupID: groupID,
			StudentID:     studentID,
			ClassID:       classRef,
			Status:        1,
		}); err != nil {
			return err
		}
	}
	return nil
}

func sameUint64PointerValue(left, right *uint64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func uint64Ptr(value uint64) *uint64 {
	ptr := new(uint64)
	*ptr = value
	return ptr
}
