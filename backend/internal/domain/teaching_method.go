package domain

import (
	"time"

	"github.com/google/uuid"
)

// TeachingMethod represents a teaching method from the "50 methods" list
type TeachingMethod struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	NameKaz    string     `json:"name_kaz" db:"name_kaz"`
	NameRus    string     `json:"name_rus" db:"name_rus"`
	NameEng    string     `json:"name_eng" db:"name_eng"`
	TypeKaz    *string    `json:"type_kaz,omitempty" db:"type_kaz"`
	PurposeKaz *string    `json:"purpose_kaz,omitempty" db:"purpose_kaz"`
	PurposeRus *string    `json:"purpose_rus,omitempty" db:"purpose_rus"`
	PurposeEng *string    `json:"purpose_eng,omitempty" db:"purpose_eng"`
	Modules    *string    `json:"modules,omitempty" db:"modules"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// KMJPlanMethod represents the many-to-many relationship between plans and methods
type KMJPlanMethod struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PlanID    uuid.UUID `json:"plan_id" db:"plan_id"`
	MethodID  uuid.UUID `json:"method_id" db:"method_id"`
	SortOrder int       `json:"sort_order" db:"sort_order"` // 1-4
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetLocalizedName returns method name in specified language
func (tm *TeachingMethod) GetLocalizedName(lang string) string {
	switch lang {
	case "kaz":
		return tm.NameKaz
	case "rus":
		return tm.NameRus
	case "eng":
		return tm.NameEng
	default:
		return tm.NameKaz
	}
}

// GetLocalizedPurpose returns purpose in specified language
func (tm *TeachingMethod) GetLocalizedPurpose(lang string) string {
	switch lang {
	case "kaz":
		if tm.PurposeKaz != nil {
			return *tm.PurposeKaz
		}
	case "rus":
		if tm.PurposeRus != nil {
			return *tm.PurposeRus
		}
	case "eng":
		if tm.PurposeEng != nil {
			return *tm.PurposeEng
		}
	}
	if tm.PurposeKaz != nil {
		return *tm.PurposeKaz
	}
	return ""
}
