package scoring

import (
	"slices"

	"nvim-quest/internal/lessons"
	"nvim-quest/internal/progress"
)

func Complete(model *progress.Model, lesson lessons.Lesson, challenge lessons.Challenge) int {
	if model.HasCompleted(challenge.ID) {
		model.LastLessonID = lesson.ID
		return 0
	}
	model.XP += challenge.XP
	model.Streak++
	model.CompletedChallenges = append(model.CompletedChallenges, challenge.ID)
	model.LastLessonID = lesson.ID
	addBadge(model, "First Steps")

	switch lesson.ID {
	case "002-motions":
		if lessonComplete(*model, lesson) {
			addBadge(model, "Motion Rookie")
		}
	case "005-delete":
		if lessonComplete(*model, lesson) {
			addBadge(model, "Delete Initiate")
		}
	case "007-search":
		if lessonComplete(*model, lesson) {
			addBadge(model, "Search Scout")
		}
	}
	return challenge.XP
}

func lessonComplete(model progress.Model, lesson lessons.Lesson) bool {
	for _, challenge := range lesson.Challenges {
		if !model.HasCompleted(challenge.ID) {
			return false
		}
	}
	return true
}

func addBadge(model *progress.Model, badge string) {
	if !slices.Contains(model.Badges, badge) {
		model.Badges = append(model.Badges, badge)
	}
}
