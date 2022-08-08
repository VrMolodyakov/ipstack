package ipstack

import (
	"context"
	"encoding/json"
	"fmt"
	"ipstack/internal/domain/entity"
	"ipstack/pkg/logging"
	"net/http"
)

const url string = "http://api.ipstack.com"

type HttpService struct {
	logger *logging.Logger
	reques string
}

func NewHttpService(key string, logger *logging.Logger) *HttpService {
	var ipstack HttpService
	ipstack.reques = url + "/%s?access_key=" + key
	ipstack.logger = logger
	return &ipstack
}

func (h *HttpService) GetIPInfo(ctx context.Context, ip string) (entity.IPInfoDto, error) {
	url := fmt.Sprintf(h.reques, ip)
	cx, cancel := context.WithCancel(ctx)
	defer cancel()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		h.logger.Errorf("coudn't make request due to %v", err)
		return entity.IPInfoDto{}, err
	}
	request = request.WithContext(cx)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		h.logger.Errorf("coudn't parse response due to %v", err)
		return entity.IPInfoDto{}, err
	}
	defer resp.Body.Close()
	var info entity.IPInfoDto
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		h.logger.Errorf("coudn't parse response due to %v", err)
		return entity.IPInfoDto{}, err
	}
	return info, nil
}
