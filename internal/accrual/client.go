package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ilya-burinskiy/gophermart/internal/models"
)

type ApiClient interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (OrderInfo, error)
}

type apiClient struct {
	baseURL    string
	httpClient http.Client
}

type OrderInfo struct {
	Number  string             `json:"number"`
	Status  models.OrderStatus `json:"status"`
	Accrual int                `json:"accrual"`
}

func (info *OrderInfo) UnmarshalJSON(data []byte) error {
	type OrderInfoAlias OrderInfo

	var string2OrderStatus = map[string]models.OrderStatus{
		"REGISTERED": models.RegisteredOrder,
		"PROCESSING": models.ProcessingOrder,
		"PROCESSED":  models.ProcessedOrder,
		"INVALID":    models.InvalidOrder,
	}
	aliasValue := &struct {
		*OrderInfoAlias
		Status string `json:"status"`
	}{
		OrderInfoAlias: (*OrderInfoAlias)(info),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	info.Status = string2OrderStatus[aliasValue.Status]
	return nil
}

func NewClient(baseURL string) ApiClient {
	return apiClient{
		baseURL:    baseURL,
		httpClient: http.Client{},
	}
}

func (client apiClient) GetOrderInfo(ctx context.Context, orderNumber string) (OrderInfo, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/orders/%s", client.baseURL, orderNumber), nil)
	if err != nil {
		return OrderInfo{}, fmt.Errorf("failed to build request: %w", err)
	}

	request = request.WithContext(ctx)
	orderInfo, err := client.getOrderInfo(request)
	if err != nil {
		return OrderInfo{}, fmt.Errorf("failed to send request to accrual service: %w", err)
	}

	return orderInfo, nil
}

func (client apiClient) getOrderInfo(req *http.Request) (OrderInfo, error) {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := client.httpClient.Do(req)
	if err != nil {
		return OrderInfo{}, err
	}

	defer res.Body.Close()
	var orderInfo OrderInfo
	switch res.StatusCode {
	case http.StatusOK:
		err = json.NewDecoder(res.Body).Decode(&orderInfo)
		if err != nil {
			return OrderInfo{}, fmt.Errorf("failed to parse request body: %w", err)
		}

		return orderInfo, nil
	case http.StatusNoContent:
		return OrderInfo{}, fmt.Errorf("order with number not found")
	case http.StatusTooManyRequests:
	case http.StatusInternalServerError:
		return OrderInfo{}, fmt.Errorf("service is unavailable")
	}

	return orderInfo, nil
}
