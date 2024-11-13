package templates

import (
	_ "embed"
	"html/template"
)

//go:embed shift.html
var shiftTemplate string

var Shift = template.Must(template.New("shift.html").Parse(shiftTemplate))
