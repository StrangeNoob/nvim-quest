package progress

type Model struct {
	XP                  int      `json:"xp"`
	Streak              int      `json:"streak"`
	CompletedChallenges []string `json:"completed_challenges"`
	Badges              []string `json:"badges"`
	LastLessonID        string   `json:"last_lesson_id"`
}

func (m Model) HasCompleted(id string) bool {
	for _, completed := range m.CompletedChallenges {
		if completed == id {
			return true
		}
	}
	return false
}
