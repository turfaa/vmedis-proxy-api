package sale

type AggregatedSale struct {
	DrugName string  `json:"drugName"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}
