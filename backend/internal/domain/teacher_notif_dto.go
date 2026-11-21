package domain

import "github.com/google/uuid"

type TeacherNotifyStudentRequest struct {
	StudentID uuid.UUID `json:"student_id" validate:"required"`
	Title     string    `json:"title" validate:"required"`
	Message   string    `json:"message" validate:"required"`
}

type TeacherNotifyClassRequest struct {
	ClassID int64  `json:"class_id" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type TeacherNotifyAllStudentsRequest struct {
	Title   string `json:"title" validate:"required"`
	Message string `json:"message" validate:"required"`
}
