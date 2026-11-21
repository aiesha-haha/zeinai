package domain

type StudentDashboardResponse struct {
	User    User           `json:"user"`
	Classes []StudentClass `json:"classes"`
}

type StudentClass struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	TeacherName      string `json:"teacher_name"`
	CreatedAt        string `json:"created_at"`
	HasActiveSession bool   `json:"has_active_session"`
}
