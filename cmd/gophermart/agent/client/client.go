package client

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AccrualResponse struct {
	Order   string
	Status  string
	Accrual int
}

type Client struct {
	address    string
	restClient *resty.Client
}

func New(address string) *Client {
	client := resty.New()
	client.AddRetryCondition(
		func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		},
	)
	return &Client{address: address, restClient: client}
}

func (c *Client) GetAccrual(number string) (*AccrualResponse, error) {
	response := &AccrualResponse{}
	_, err := c.restClient.R().
		SetResult(response).
		ForceContentType("application/json").
		SetHeader("Content-Type", "application/json").
		SetPathParams(map[string]string{
			"address": c.address,
			"number":  number,
		}).
		Get("http://{address}/api/orders/{number}")

	return response, err
}
