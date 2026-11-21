package domain

import (
	"context"

	"github.com/google/uuid"
)

type ClassRepository interface {
	Create(ctx context.Context, class *Class) error
	GetByID(ctx context.Context, id int64) (*Class, error)
	GetByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]Class, error)
	Update(ctx context.Context, class *Class) error
	Delete(ctx context.Context, id int64) error
	AddStudent(ctx context.Context, classID int64, studentID uuid.UUID) error
	RemoveStudent(ctx context.Context, classID int64, studentID uuid.UUID) error
	GetStudents(ctx context.Context, classID int64) ([]User, error)
	GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]Class, error)
	IsStudentInClass(ctx context.Context, classID int64, studentID uuid.UUID) (bool, error)

	// Invites
	CreateInvite(ctx context.Context, invite *ClassInvite) error
	GetInviteByCode(ctx context.Context, code string) (*ClassInvite, error)
	GetActiveInvite(ctx context.Context, classID int64) (*ClassInvite, error)
	ExpireAllInvites(ctx context.Context, classID int64) error
	IncrementInviteUsage(ctx context.Context, id int64) error
	GetClassStudentCount(ctx context.Context, classID int64) (int, error)
}
