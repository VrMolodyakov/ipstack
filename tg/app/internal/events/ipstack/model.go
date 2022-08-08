package ipstack

import "app/internal/events/model"

type IPInfoRequest struct {
	RequestID string `json:"request_id"`
	IP        string `json:"ip"`
	Nickname  string `json:"nickname"`
}

type IPInfoResponse struct {
	Meta model.ResponseMeta `json:"meta"`
	Data ResponseData       `json:"data"`
}

type ResponseData struct {
	Info map[string]interface{} `json:"info"`
}
