package domain

import (
	"time"

	"github.com/google/uuid"
)

// NationalValue represents a national value according to "Adal Azamat" program
type NationalValue struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	MonthNumber int        `json:"month_number" db:"month_number"` // 1-12
	NameKaz     string     `json:"name_kaz" db:"name_kaz"`
	NameRus     string     `json:"name_rus" db:"name_rus"`
	NameEng     string     `json:"name_eng" db:"name_eng"`
	DescKaz     *string    `json:"description_kaz,omitempty" db:"description_kaz"`
	DescRus     *string    `json:"description_rus,omitempty" db:"description_rus"`
	DescEng     *string    `json:"description_eng,omitempty" db:"description_eng"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// GetLocalizedName returns name in specified language
func (nv *NationalValue) GetLocalizedName(lang string) string {
	switch lang {
	case "kaz":
		return nv.NameKaz
	case "rus":
		return nv.NameRus
	case "eng":
		return nv.NameEng
	default:
		return nv.NameKaz // Default to Kazakh
	}
}

// GetLocalizedDescription returns description in specified language
func (nv *NationalValue) GetLocalizedDescription(lang string) string {
	switch lang {
	case "kaz":
		if nv.DescKaz != nil {
			return *nv.DescKaz
		}
	case "rus":
		if nv.DescRus != nil {
			return *nv.DescRus
		}
	case "eng":
		if nv.DescEng != nil {
			return *nv.DescEng
		}
	}
	if nv.DescKaz != nil {
		return *nv.DescKaz
	}
	return ""
}

// Month names in Kazakh
var MonthNamesKaz = map[int]string{
	1:  "Қаңтар",
	2:  "Ақпан",
	3:  "Наурыз",
	4:  "Сәуір",
	5:  "Мамыр",
	6:  "Маусым",
	7:  "Шілде",
	8:  "Тамыз",
	9:  "Қыркүйек",
	10: "Қазан",
	11: "Қараша",
	12: "Желтоқсан",
}
