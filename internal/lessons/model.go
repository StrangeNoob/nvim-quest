package lessons

type Cursor struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Success struct {
	CursorOnWord      string   `json:"cursor_on_word,omitempty"`
	BufferEquals      []string `json:"buffer_equals,omitempty"`
	LineDeleted       string   `json:"line_deleted,omitempty"`
	WordDeleted       string   `json:"word_deleted,omitempty"`
	ContainsText      string   `json:"contains_text,omitempty"`
	SearchMatchActive string   `json:"search_match_active,omitempty"`
}

type Challenge struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Objective       string   `json:"objective"`
	InitialBuffer   []string `json:"initial_buffer"`
	InitialCursor   Cursor   `json:"initial_cursor"`
	AllowedCommands []string `json:"allowed_commands"`
	Success         Success  `json:"success"`
	Hint            string   `json:"hint"`
	XP              int      `json:"xp"`
}

type Lesson struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Topic      string      `json:"topic"`
	Challenges []Challenge `json:"challenges"`
}
