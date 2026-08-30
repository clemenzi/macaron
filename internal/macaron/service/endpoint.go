package service

// Endpoint describes a running service exposed by Macaron.
type Endpoint struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}
