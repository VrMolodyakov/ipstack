package request

type IpSearchRequest struct {
	RequestID string `json:"request_id"`
	IP        string `json:"ip"`
	Nickname  string `json:"nickname"`
}
