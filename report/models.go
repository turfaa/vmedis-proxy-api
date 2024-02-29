package report

type DrugQuantity struct {
	DrugName string  `json:"drugName"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}
