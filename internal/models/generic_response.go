package models

// GenericStatus is a type that represents a generic status
type GenericStatus string

// GenericStatus constants
const (
	Success GenericStatus = "success" // Success   GenericStatus = "success"
	Error   GenericStatus = "error"   // Error     GenericStatus = "error"
)

// GenericResponse is a struct that represents a generic response
type GenericResponse struct {
	Status     GenericStatus `json:"status"`
	Message    string        `json:"message"`
	StatusCode int           `json:"status_code"`
	Data       interface{}   `json:"data,omitempty"`
}
