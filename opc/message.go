package opc

// Message represent an opc message
type Message struct {
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
	Data      string `json:"data"`
}
