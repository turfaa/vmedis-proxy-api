package drug

type DrugsResponseV2 struct {
	Drugs []DrugsResponseV2_Drug `json:"drugs"`
}

type DrugsResponseV2_Drug struct {
	Name     string    `json:"name"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Title string   `json:"title"`
	Rows  []string `json:"rows"`
}
