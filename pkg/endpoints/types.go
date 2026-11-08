package endpoints

// Endpoint represents a Portainer environment (endpoint).
type Endpoint struct {
	ID     int    `json:"Id"`
	Name   string `json:"Name"`
	Type   int    `json:"Type"`
	URL    string `json:"URL"`
	Status int    `json:"Status"`
}
