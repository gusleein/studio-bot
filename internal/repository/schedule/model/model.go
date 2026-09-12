package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ScheduleModel — модель занятия в расписании для хранения в БД.
// CourseID и TeacherID хранятся как uuid.NullUUID, поскольку в БД это nullable FK.
// SheetsRowID хранится как sql.NullInt64 (nullable).
type ScheduleModel struct {
	ID          uuid.UUID     `db:"id"`
	CourseID    uuid.NullUUID `db:"course_id"`
	TeacherID   uuid.NullUUID `db:"teacher_id"`
	StartsAt    time.Time     `db:"starts_at"`
	EndsAt      time.Time     `db:"ends_at"`
	Room        string        `db:"room"`
	MaxStudents int           `db:"max_students"`
	IsCancelled bool          `db:"is_cancelled"`
	Notes       string        `db:"notes"`
	SheetsRowID sql.NullInt64 `db:"sheets_row_id"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`

	// Joined поля — заполняются при запросе с JOIN на teachers и courses.
	TeacherName string `db:"teacher_name"`
	CourseTitle string `db:"course_title"`
	CourseType  string `db:"course_type"`
}

// ScheduleStudentModel — модель записи студента на занятие.
type ScheduleStudentModel struct {
	ScheduleID uuid.UUID `db:"schedule_id"`
	StudentID  uuid.UUID `db:"student_id"`
	CreatedAt  time.Time `db:"created_at"`
}
