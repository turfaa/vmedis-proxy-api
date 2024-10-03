package cui

type Table struct {
	Header []string `json:"header"`
	Rows   []Row    `json:"rows"`
}

type Row struct {
	ID      string   `json:"id"`
	Columns []string `json:"columns"`
}
