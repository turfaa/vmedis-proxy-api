package cui

// Form is a display-ready input form that the frontend can render to create
// or update a resource. Field IDs match the JSON field names expected by the
// corresponding write endpoint, so a filled form can be submitted as-is.
type Form struct {
	Title  string  `json:"title"`
	Fields []Field `json:"fields"`
}

type Field struct {
	// ID is the JSON field name to use when submitting the field's value.
	ID    string    `json:"id"`
	Label string    `json:"label"`
	Type  FieldType `json:"type"`

	// Value is the raw current value of the field, used to prefill the input.
	Value string `json:"value"`

	// Options are the selectable values of a FieldTypeSelect field.
	Options []Option `json:"options,omitempty"`

	Required bool `json:"required"`
}

type FieldType string

const (
	FieldTypeText     FieldType = "TEXT"
	FieldTypeTextArea FieldType = "TEXTAREA"
	FieldTypeSelect   FieldType = "SELECT"
)
