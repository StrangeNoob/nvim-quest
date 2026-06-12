package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"nvim-quest/internal/engine"
	"nvim-quest/internal/lessons"
	"nvim-quest/internal/progress"
	"nvim-quest/internal/scoring"
)

type challengeRef struct {
	lesson    lessons.Lesson
	challenge lessons.Challenge
}

type Model struct {
	items        []challengeRef
	index        int
	simulator    *engine.Simulator
	input        textinput.Model
	progress     progress.Model
	store        progress.Store
	feedback     string
	feedbackGood bool
	showHint     bool
	completed    bool
	finished     bool
}

func NewModel(selected []lessons.Lesson, saved progress.Model, store progress.Store, skipCompleted bool) Model {
	var items []challengeRef
	for _, lesson := range selected {
		for _, challenge := range lesson.Challenges {
			if skipCompleted && saved.HasCompleted(challenge.ID) {
				continue
			}
			items = append(items, challengeRef{lesson: lesson, challenge: challenge})
		}
	}
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "type a Vim command, then press Enter"
	input.Focus()

	model := Model{items: items, input: input, progress: saved, store: store, showHint: true}
	if len(items) == 0 {
		model.finished = true
	} else {
		model.loadCurrent()
	}
	return model
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.input.Value() == "" {
				return m, tea.Quit
			}
		case "?":
			if m.input.Value() == "" {
				m.showHint = !m.showHint
				return m, nil
			}
		case "enter":
			return m.submit()
		}
	}

	var command tea.Cmd
	m.input, command = m.input.Update(message)
	return m, command
}

func (m Model) View() string {
	if m.finished {
		return titleStyle.Render("nvim-quest") + "\n\n" +
			successStyle.Render("Quest complete. Your progress is saved.") + "\n\n" +
			fmt.Sprintf("XP: %d  Streak: %d\n\nPress q to quit.\n", m.progress.XP, m.progress.Streak)
	}

	item := m.items[m.index]
	var view strings.Builder
	fmt.Fprintf(&view, "%s  %s\n", titleStyle.Render("nvim-quest"), mutedStyle.Render(fmt.Sprintf("Lesson %s", item.lesson.ID)))
	fmt.Fprintf(&view, "%s\n\n", titleStyle.Render(item.challenge.Title))
	fmt.Fprintf(&view, "%s\n%s\n\n", sectionStyle.Render("Objective"), item.challenge.Objective)
	fmt.Fprintf(&view, "%s\n%s\n\n", sectionStyle.Render("Buffer"), renderBuffer(m.simulator.Buffer))
	fmt.Fprintf(&view, "%s\n%s\n", sectionStyle.Render("Command"), m.input.View())
	fmt.Fprintf(&view, "%s %s\n\n", mutedStyle.Render("Available:"), strings.Join(item.challenge.AllowedCommands, " "))
	if m.showHint {
		fmt.Fprintf(&view, "%s\n%s\n\n", sectionStyle.Render("Hint"), item.challenge.Hint)
	} else {
		fmt.Fprintf(&view, "%s\n\n", mutedStyle.Render("Press ? to show the hint."))
	}
	if m.feedback != "" {
		style := errorStyle
		if m.feedbackGood {
			style = successStyle
		}
		fmt.Fprintf(&view, "%s\n%s\n\n", sectionStyle.Render("Feedback"), style.Render(m.feedback))
	}
	fmt.Fprintf(&view, "%s\n%s\n", sectionStyle.Render("Progress"), progressStyle.Render(fmt.Sprintf("XP: %d  Streak: %d", m.progress.XP, m.progress.Streak)))
	fmt.Fprint(&view, mutedStyle.Render("q quit  ? toggle hint  Enter submit"))
	return view.String()
}

func (m Model) submit() (tea.Model, tea.Cmd) {
	if m.finished {
		return m, nil
	}
	if m.completed {
		m.index++
		m.feedback = ""
		m.completed = false
		if m.index >= len(m.items) {
			m.finished = true
		} else {
			m.loadCurrent()
		}
		return m, nil
	}

	value := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")
	if value == "" {
		return m, nil
	}
	command, err := engine.ParseCommand(value)
	if err != nil || !engine.IsAllowed(command, m.items[m.index].challenge.AllowedCommands) {
		m.feedback = helpfulInvalid(value)
		m.feedbackGood = false
		return m, nil
	}
	if err := m.simulator.Apply(command); err != nil {
		m.feedback = err.Error()
		m.feedbackGood = false
		return m, nil
	}

	item := m.items[m.index]
	if engine.Validate(m.simulator, item.challenge) {
		awarded := scoring.Complete(&m.progress, item.lesson, item.challenge)
		if err := m.store.Save(m.progress); err != nil {
			m.feedback = fmt.Sprintf("Correct, but progress could not be saved: %v", err)
			m.feedbackGood = false
			return m, nil
		}
		m.feedback = fmt.Sprintf("Correct. +%d XP. Press Enter for the next challenge.", awarded)
		m.feedbackGood = true
		m.completed = true
		return m, nil
	}
	m.feedback = "Command applied. Keep going."
	m.feedbackGood = true
	return m, nil
}

func (m *Model) loadCurrent() {
	challenge := m.items[m.index].challenge
	m.simulator = engine.NewSimulator(challenge.InitialBuffer, engine.Cursor{
		Row: challenge.InitialCursor.Row,
		Col: challenge.InitialCursor.Col,
	})
	m.input.SetValue("")
}

func helpfulInvalid(command string) string {
	switch command {
	case "x":
		return "x deletes one character. To delete a full word, try dw."
	case "dd":
		return "dd deletes a whole line. Check the available commands for this challenge."
	case "dw":
		return "dw deletes from the cursor through the current word. Check the objective and available commands."
	default:
		return "That command is not available in this challenge."
	}
}

func renderBuffer(buffer engine.Buffer) string {
	var result strings.Builder
	for row, line := range buffer.Lines {
		fmt.Fprintf(&result, "%2d  ", row+1)
		if row == buffer.Cursor.Row {
			col := min(buffer.Cursor.Col, len(line))
			before := line[:col]
			current := " "
			after := ""
			if col < len(line) {
				current = line[col : col+1]
				after = line[col+1:]
			}
			result.WriteString(before + cursorStyle.Render(current) + after)
		} else {
			result.WriteString(line)
		}
		result.WriteByte('\n')
	}
	return result.String()
}
