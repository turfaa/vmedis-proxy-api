package cui

type Table struct {
	Header []string `json:"header"`
	Rows   []Row    `json:"rows"`
	// Footer is rendered as a summary row below the rows, e.g. to show totals.
	// It is omitted when empty.
	Footer []string `json:"footer,omitempty"`
}

type Row struct {
	ID      string   `json:"id"`
	Columns []string `json:"columns"`
}
