package cui

// Options is a standalone list of selectable options,
// e.g. for filter dropdowns.
type Options struct {
	Options []Option `json:"options"`
}

// Option is a raw value with its human-friendly label.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
