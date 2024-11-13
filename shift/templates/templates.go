package templates

import (
	_ "embed"
	"html/template"
)

//go:embed shift.html
var shiftTemplate string

var Shift = template.Must(template.New("shift.html").Parse(shiftTemplate))

type ShiftData struct {
	Code                string
	Cashier             string
	StartedAt           string
	EndedAt             string
	InitialCash         string
	ExpectedFinalCash   string
	ActualFinalCash     string
	FinalCashDifference string
	Supervisor          string
	Notes               string
}
