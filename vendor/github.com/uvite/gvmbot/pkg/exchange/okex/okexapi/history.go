package okexapi

import (
	"context"
	"fmt"
	"net/url"
)

func (c *TradeService) NewGetHistoryTrade() *GetHistoryRequest {
	return &GetHistoryRequest{
		client: c.client,
	}
}

type GetHistoryRequest struct {
	client *RestClient

	instType *InstrumentType

	instId *string

	ordId *string
}

func (r *GetHistoryRequest) InstrumentType(instType InstrumentType) *GetHistoryRequest {
	r.instType = &instType
	return r
}

func (r *GetHistoryRequest) InstrumentID(instId string) *GetHistoryRequest {
	r.instId = &instId
	return r
}

func (r *GetHistoryRequest) Parameters() url.Values {

	payload := url.Values{}

	if r.instType != nil {

		payload.Add("instType", string(*r.instType))
	}

	if r.instId != nil {

		//payload.Add("instId", *r.instId)

	}

	payload.Add("limit", fmt.Sprintf("%d", 100))

	return payload
}

func (r *GetHistoryRequest) Do(ctx context.Context) ([]OrderDetails, error) {
	params := r.Parameters()
	req, err := r.client.newAuthenticatedRequest("GET", "/api/v5/trade/orders-history-archive", params, nil)
	if err != nil {
		return nil, err
	}

	response, err := r.client.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var orderResponse struct {
		Code    string         `json:"code"`
		Message string         `json:"msg"`
		Data    []OrderDetails `json:"data"`
	}
	if err := response.DecodeJSON(&orderResponse); err != nil {
		return nil, err
	}

	return orderResponse.Data, nil
}
