package converter

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/schedule/model"
)

// ToDomain преобразует модель БД в доменный объект Schedule.
// Nullable-поля БД (NullUUID, NullInt64) маппятся в нулевые значения Go при отсутствии данных.
func ToDomain(m *model.ScheduleModel) *domain.Schedule {
	return &domain.Schedule{
		ID:          m.ID,
		CourseID:    nullUUIDToUUID(m.CourseID),
		TeacherID:   nullUUIDToUUID(m.TeacherID),
		StartsAt:    m.StartsAt,
		EndsAt:      m.EndsAt,
		IsCancelled: m.IsCancelled,
		Notes:       m.Notes,
		SheetsRowID: nullInt64ToInt(m.SheetsRowID),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		TeacherName: m.TeacherName,
		CourseTitle: m.CourseTitle,
		CourseType:  m.CourseType,
	}
}

// ToModel преобразует доменный объект Schedule в модель БД.
// Нулевые UUID и нулевые int сохраняются в БД как NULL.
func ToModel(s *domain.Schedule) *model.ScheduleModel {
	return &model.ScheduleModel{
		ID:          s.ID,
		CourseID:    uuidToNullUUID(s.CourseID),
		TeacherID:   uuidToNullUUID(s.TeacherID),
		StartsAt:    s.StartsAt,
		EndsAt:      s.EndsAt,
		IsCancelled: s.IsCancelled,
		Notes:       s.Notes,
		SheetsRowID: intToNullInt64(s.SheetsRowID),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// ScheduleStudentToDomain преобразует модель БД в доменный объект ScheduleStudent.
func ScheduleStudentToDomain(m *model.ScheduleStudentModel) *domain.ScheduleStudent {
	return &domain.ScheduleStudent{
		ScheduleID: m.ScheduleID,
		StudentID:  m.StudentID,
		CreatedAt:  m.CreatedAt,
	}
}

// nullUUIDToUUID возвращает uuid.UUID из uuid.NullUUID; uuid.Nil если не валидный.
func nullUUIDToUUID(n uuid.NullUUID) uuid.UUID {
	if !n.Valid {
		return uuid.Nil
	}
	return n.UUID
}

// uuidToNullUUID конвертирует uuid.UUID в uuid.NullUUID; невалидный если uuid.Nil.
func uuidToNullUUID(u uuid.UUID) uuid.NullUUID {
	if u == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: u, Valid: true}
}

// nullInt64ToInt возвращает int из sql.NullInt64; 0 если не валидный.
func nullInt64ToInt(n sql.NullInt64) int {
	if !n.Valid {
		return 0
	}
	return int(n.Int64)
}

// intToNullInt64 конвертирует int в sql.NullInt64; невалидный если значение равно 0.
func intToNullInt64(v int) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(v), Valid: true}
}
