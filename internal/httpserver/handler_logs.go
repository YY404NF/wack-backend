package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/query"
)

func (h *apiHandler) adminAttendanceRecordLogs(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	logs, err := h.logs.RecordLogs(id)
	if err != nil {
		fail(c, 500, "load attendance record logs failed")
		return
	}
	ok(c, logs)
}

func (h *apiHandler) listAttendanceRecordLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	logs, total, err := h.logs.AttendanceRecordLogs(query.AttendanceRecordLogListInput{
		Term:                c.Query("term"),
		CourseGroupLessonID: c.Query("course_group_lesson_id"),
		LessonDate:          c.Query("lesson_date"),
		Section:             c.Query("section"),
		CourseName:          c.Query("course_name"),
		TeacherName:         c.Query("teacher_name"),
		StudentID:           c.Query("student_id"),
		RealName:            c.Query("real_name"),
		ClassName:           c.Query("class_name"),
		OldStatus:           c.Query("old_status"),
		NewStatus:           c.Query("new_status"),
		OperatorName:        c.Query("operator_name"),
		OperatedDate:        c.Query("operated_date"),
		Page:                page,
		PageSize:            pageSize,
	})
	if err != nil {
		fail(c, 500, "list attendance record logs failed")
		return
	}
	ok(c, pageResult[query.AttendanceRecordLogListItem]{Items: logs, Page: page, PageSize: pageSize, Total: total})
}
