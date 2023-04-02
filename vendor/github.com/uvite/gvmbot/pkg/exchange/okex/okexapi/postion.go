package okexapi

import (
	"context"
)

type GetPostionRequest struct {
	client *RestClient

	instType *InstrumentType

	instId *string

	posId *string
}
type Postions struct {
	PositionSide     string `json:"posSide"`
	PositionAmt      string `json:"pos"`
	UnRealizedProfit string `json:"upl"`
	EntryPrice       string `json:"avgPx"`
	Leverage         string `json:"lever"`
	LiquidationPrice string `json:"liqPx"`
	MarkPrice        string `json:"markPx"`
	Symbol           string `json:"instId"`
}

func (c *TradeService) NewGetAccountPostions() *GetPostionRequest {
	return &GetPostionRequest{
		client: c.client,
	}
}
func (r *GetPostionRequest) InstrumentType(instType InstrumentType) *GetPostionRequest {
	r.instType = &instType
	return r
}

func (r *GetPostionRequest) InstrumentID(instId string) *GetPostionRequest {
	r.instId = &instId
	return r
}

func (r *GetPostionRequest) PosId(posId string) *GetPostionRequest {
	r.posId = &posId
	return r
}

func (r *GetPostionRequest) Parameters() map[string]interface{} {
	var payload = map[string]interface{}{}

	if r.instType != nil {
		payload["instType"] = r.instType
	}

	if r.instId != nil {
		payload["instId"] = r.instId
	}

	if r.posId != nil {
		payload["posId"] = r.posId
	}

	return payload
}

func (r *GetPostionRequest) Do(ctx context.Context) ([]*Postions, error) {
	payload := r.Parameters()
	req, err := r.client.newAuthenticatedRequest("GET", "/api/v5/account/positions", nil, payload)
	if err != nil {
		return nil, err
	}

	response, err := r.client.sendRequest(req)
	if err != nil {
		return nil, err
	}


	var orderResponse struct {
		Code    string      `json:"code"`
		Message string      `json:"msg"`
		Data    []*Postions `json:"data"`
	}
	if err := response.DecodeJSON(&orderResponse); err != nil {
		return nil, err
	}

	return orderResponse.Data, nil
}
