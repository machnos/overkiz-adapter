package domain

type actionRequest struct {
	Label   string    `json:"label"`
	Actions []*action `json:"actions"`
}

type action struct {
	Commands  []*command `json:"commands"`
	DeviceURL string     `json:"deviceURL" json:"device_url"`
}

type command struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters,omitempty"`
}
