package health

type Status struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}
