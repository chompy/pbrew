package core

// ServiceInfo is information about a service as provided by Homebrew.
type ServiceInfo struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}
