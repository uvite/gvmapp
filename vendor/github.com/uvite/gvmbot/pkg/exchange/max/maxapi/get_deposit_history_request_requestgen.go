// Code generated by "requestgen -method GET -url v2/deposits -type GetDepositHistoryRequest -responseType []Deposit"; DO NOT EDIT.

package max

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
)

func (g *GetDepositHistoryRequest) Currency(currency string) *GetDepositHistoryRequest {
	g.currency = &currency
	return g
}

func (g *GetDepositHistoryRequest) From(from int64) *GetDepositHistoryRequest {
	g.from = &from
	return g
}

func (g *GetDepositHistoryRequest) To(to int64) *GetDepositHistoryRequest {
	g.to = &to
	return g
}

func (g *GetDepositHistoryRequest) State(state string) *GetDepositHistoryRequest {
	g.state = &state
	return g
}

func (g *GetDepositHistoryRequest) Limit(limit int) *GetDepositHistoryRequest {
	g.limit = &limit
	return g
}

// GetQueryParameters builds and checks the query parameters and returns url.Values
func (g *GetDepositHistoryRequest) GetQueryParameters() (url.Values, error) {
	var params = map[string]interface{}{}

	query := url.Values{}
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParameters builds and checks the parameters and return the result in a map object
func (g *GetDepositHistoryRequest) GetParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}
	// check currency field -> json key currency
	if g.currency != nil {
		currency := *g.currency

		// assign parameter of currency
		params["currency"] = currency
	} else {
	}
	// check from field -> json key from
	if g.from != nil {
		from := *g.from

		// assign parameter of from
		params["from"] = from
	} else {
	}
	// check to field -> json key to
	if g.to != nil {
		to := *g.to

		// assign parameter of to
		params["to"] = to
	} else {
	}
	// check state field -> json key state
	if g.state != nil {
		state := *g.state

		// assign parameter of state
		params["state"] = state
	} else {
	}
	// check limit field -> json key limit
	if g.limit != nil {
		limit := *g.limit

		// assign parameter of limit
		params["limit"] = limit
	} else {
	}

	return params, nil
}

// GetParametersQuery converts the parameters from GetParameters into the url.Values format
func (g *GetDepositHistoryRequest) GetParametersQuery() (url.Values, error) {
	query := url.Values{}

	params, err := g.GetParameters()
	if err != nil {
		return query, err
	}

	for k, v := range params {
		if g.isVarSlice(v) {
			g.iterateSlice(v, func(it interface{}) {
				query.Add(k+"[]", fmt.Sprintf("%v", it))
			})
		} else {
			query.Add(k, fmt.Sprintf("%v", v))
		}
	}

	return query, nil
}

// GetParametersJSON converts the parameters from GetParameters into the JSON format
func (g *GetDepositHistoryRequest) GetParametersJSON() ([]byte, error) {
	params, err := g.GetParameters()
	if err != nil {
		return nil, err
	}

	return json.Marshal(params)
}

// GetSlugParameters builds and checks the slug parameters and return the result in a map object
func (g *GetDepositHistoryRequest) GetSlugParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

func (g *GetDepositHistoryRequest) applySlugsToUrl(url string, slugs map[string]string) string {
	for k, v := range slugs {
		needleRE := regexp.MustCompile(":" + k + "\\b")
		url = needleRE.ReplaceAllString(url, v)
	}

	return url
}

func (g *GetDepositHistoryRequest) iterateSlice(slice interface{}, f func(it interface{})) {
	sliceValue := reflect.ValueOf(slice)
	for i := 0; i < sliceValue.Len(); i++ {
		it := sliceValue.Index(i).Interface()
		f(it)
	}
}

func (g *GetDepositHistoryRequest) isVarSlice(v interface{}) bool {
	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.Slice:
		return true
	}
	return false
}

func (g *GetDepositHistoryRequest) GetSlugsMap() (map[string]string, error) {
	slugs := map[string]string{}
	params, err := g.GetSlugParameters()
	if err != nil {
		return slugs, nil
	}

	for k, v := range params {
		slugs[k] = fmt.Sprintf("%v", v)
	}

	return slugs, nil
}

func (g *GetDepositHistoryRequest) Do(ctx context.Context) ([]Deposit, error) {

	// empty params for GET operation
	var params interface{}
	query, err := g.GetParametersQuery()
	if err != nil {
		return nil, err
	}

	apiURL := "v2/deposits"

	req, err := g.client.NewAuthenticatedRequest(ctx, "GET", apiURL, query, params)
	if err != nil {
		return nil, err
	}

	response, err := g.client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var apiResponse []Deposit
	if err := response.DecodeJSON(&apiResponse); err != nil {
		return nil, err
	}
	return apiResponse, nil
}
