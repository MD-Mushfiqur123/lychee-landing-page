package api

type ModelEndpoint struct {
	Host  string `json:"host"`            // e.g. "http://192.168.1.10:11434"
	Model string `json:"model,omitempty"` // if different from route name
}

type ModelRoute struct {
	Name      string          `json:"name"`       // virtual model name, e.g. "fast"
	Endpoints []ModelEndpoint `json:"endpoints"`
	Strategy  string          `json:"strategy"`   // round_robin, random, least_loaded
}

type RouteRequest struct {
	Name      string          `json:"name"`
	Endpoints []ModelEndpoint `json:"endpoints"`
	Strategy  string          `json:"strategy,omitempty"` // default: round_robin
}
