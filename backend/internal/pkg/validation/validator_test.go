package validation

import (
	"testing"
)

func TestSafeNameRule(t *testing.T) {
	v := NewValidator()

	type nameHolder struct {
		Name string `validate:"required,safe_name,min=2,max=100"`
	}

	valid := []string{
		"Иван",
		"Сайлыбаев",
		"Jean-Pierre",
		"O'Brien",
		"Алия Нұрланқызы",
		"김철수",  // Korean
		"محمد",   // Arabic
	}

	invalid := []string{
		"<img src=x onerror=alert(1)>",
		"<script>alert(1)</script>",
		"Test&Name",
		`"quoted"`,
		"name; DROP TABLE users;--",
		"javascript:alert(1)",
	}

	for _, name := range valid {
		if err := v.Struct(nameHolder{Name: name}); err != nil {
			t.Errorf("expected '%s' to be valid, got error: %v", name, err)
		}
	}

	for _, name := range invalid {
		if err := v.Struct(nameHolder{Name: name}); err == nil {
			t.Errorf("expected '%s' to be invalid, but validation passed", name)
		}
	}
}
