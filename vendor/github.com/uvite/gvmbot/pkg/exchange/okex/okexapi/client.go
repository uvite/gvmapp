package okexapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/types"
	"github.com/uvite/gvmbot/pkg/util"
)

const defaultHTTPTimeout = time.Second * 15

const RestBaseURL = "https://www.okx.com/"
const PublicWebSocketURL = "wss://ws.okx.com:8443/ws/v5/public"
const PrivateWebSocketURL = "wss://ws.okx.com:8443/ws/v5/private"
const PublicWebSocketURLTest = "wss://wspap.okx.com:8443/ws/v5/public?brokerId=9999"
const PrivateWebSocketURLTest = "wss://wspap.okx.com:8443/ws/v5/private?brokerId=9999"

type SideType string
type PosSideType string

const (
	SideTypeBuy     SideType    = "buy"
	SideTypeSell    SideType    = "sell"
	PosSideTypeBuy  PosSideType = "long"
	PosSideTypeSell PosSideType = "short"
)

type OrderType string

const (
	OrderTypeMarket           OrderType = "market"
	OrderTypeLimit            OrderType = "limit"
	OrderTypeStopLimit        OrderType = "STOP_LIMIT"
	OrderTypeStopMarket       OrderType = "STOP_MARKET"
	OrderTypeTakeProfitLimit  OrderType = "TAKE_PROFIT_LIMIT"
	OrderTypeTakeProfitMarket OrderType = "TAKE_PROFIT_MARKET"

	OrderTypeConditional OrderType = "conditional"

	OrderTypePostOnly OrderType = "post_only"
	OrderTypeFOK      OrderType = "fok"
	OrderTypeIOC      OrderType = "ioc"
)

type InstrumentType string

const (
	InstrumentTypeSpot    InstrumentType = "SPOT"
	InstrumentTypeSwap    InstrumentType = "SWAP"
	InstrumentTypeFutures InstrumentType = "FUTURES"
	InstrumentTypeOption  InstrumentType = "OPTION"
)

type OrderState string

const (
	OrderStateCanceled        OrderState = "canceled"
	OrderStateLive            OrderState = "live"
	OrderStatePartiallyFilled OrderState = "partially_filled"
	OrderStateFilled          OrderState = "filled"
)

func PaperTrade() bool {
	v, ok := util.GetEnvVarBool("PAPER_TRADE")
	return ok && v
}

type RestClient struct {
	BaseURL *url.URL

	client *http.Client

	Key, Secret, Passphrase string

	TradeService      *TradeService
	PublicDataService *PublicDataService
	MarketDataService *MarketDataService
}

func NewClient() *RestClient {
	u, err := url.Parse(RestBaseURL)
	if err != nil {
		panic(err)
	}

	client := &RestClient{
		BaseURL: u,
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}

	client.TradeService = &TradeService{client: client}
	client.PublicDataService = &PublicDataService{client: client}
	client.MarketDataService = &MarketDataService{client: client}
	return client
}

func (c *RestClient) Auth(key, secret, passphrase string) {
	c.Key = key
	// pragma: allowlist nextline secret
	c.Secret = secret
	c.Passphrase = passphrase
}

// NewRequest create new API request. Relative url can be provided in refURL.
func (c *RestClient) newRequest(method, refURL string, params url.Values, body []byte) (*http.Request, error) {
	rel, err := url.Parse(refURL)
	if err != nil {
		return nil, err
	}

	if params != nil {
		rel.RawQuery = params.Encode()
	}

	pathURL := c.BaseURL.ResolveReference(rel)
	return http.NewRequest(method, pathURL.String(), bytes.NewReader(body))
}

// sendRequest sends the request to the API server and handle the response
func (c *RestClient) sendRequest(req *http.Request) (*util.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// newResponse reads the response body and return a new Response object
	response, err := util.NewResponse(resp)
	if err != nil {
		log.Error(err)
		return response, err
	}

	// Check error, if there is an error, return the ErrorResponse struct type
	if response.IsError() {
		return response, errors.New(string(response.Body))
	}

	return response, nil
}

// newAuthenticatedRequest creates new http request for authenticated routes.
func (c *RestClient) newAuthenticatedRequest(method, refURL string, params url.Values, payload interface{}) (*http.Request, error) {
	if len(c.Key) == 0 {
		return nil, errors.New("empty api key")
	}

	if len(c.Secret) == 0 {
		return nil, errors.New("empty api secret")
	}

	rel, err := url.Parse(refURL)
	if err != nil {
		return nil, err
	}

	if params != nil {
		rel.RawQuery = params.Encode()
	}

	pathURL := c.BaseURL.ResolveReference(rel)
	path := pathURL.Path
	if rel.RawQuery != "" {
		path += "?" + rel.RawQuery
	}

	// set location to UTC so that it outputs "2020-12-08T09:08:57.715Z"
	t := time.Now().In(time.UTC)
	timestamp := t.Format("2006-01-02T15:04:05.999Z07:00")

	var body []byte

	if payload != nil {
		switch v := payload.(type) {
		case string:
			body = []byte(v)

		case []byte:
			body = v

		default:
			body, err = json.Marshal(v)
			if err != nil {
				return nil, err
			}
		}
	}

	signKey := timestamp + strings.ToUpper(method) + path + string(body)
	signature := Sign(signKey, c.Secret)

	req, err := http.NewRequest(method, pathURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("OK-ACCESS-KEY", c.Key)
	req.Header.Add("OK-ACCESS-SIGN", signature)
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("OK-ACCESS-PASSPHRASE", c.Passphrase)

	if PaperTrade() {
		///fmt.Println("pagetrade:", PaperTrade())
		req.Header.Add("x-simulated-trading", "1")
	}
	//fmt.Println(req)
	return req, nil
}

type BalanceDetail struct {
	Currency                string                     `json:"ccy"`
	Available               fixedpoint.Value           `json:"availEq"`
	CashBalance             fixedpoint.Value           `json:"cashBal"`
	OrderFrozen             fixedpoint.Value           `json:"ordFrozen"`
	Frozen                  fixedpoint.Value           `json:"frozenBal"`
	Equity                  fixedpoint.Value           `json:"eq"`
	EquityInUSD             fixedpoint.Value           `json:"eqUsd"`
	UpdateTime              types.MillisecondTimestamp `json:"uTime"`
	UnrealizedProfitAndLoss fixedpoint.Value           `json:"upl"`
}

type Account struct {
	TotalEquityInUSD fixedpoint.Value `json:"totalEq"`
	UpdateTime       string           `json:"uTime"`
	Details          []BalanceDetail  `json:"details"`
}

func (c *RestClient) AccountBalances() (*Account, error) {
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/account/balance", nil, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.sendRequest(req)
	//fmt.Println(response)
	if err != nil {
		return nil, err
	}

	var balanceResponse struct {
		Code    string    `json:"code"`
		Message string    `json:"msg"`
		Data    []Account `json:"data"`
	}

	if err := response.DecodeJSON(&balanceResponse); err != nil {
		return nil, err
	}

	if len(balanceResponse.Data) == 0 {
		return nil, errors.New("empty account data")
	}

	return &balanceResponse.Data[0], nil
}

type AssetBalance struct {
	Currency  string           `json:"ccy"`
	Balance   fixedpoint.Value `json:"bal"`
	Frozen    fixedpoint.Value `json:"frozenBal,omitempty"`
	Available fixedpoint.Value `json:"availBal,omitempty"`
}

type AssetBalanceList []AssetBalance

func (c *RestClient) AssetBalances() (AssetBalanceList, error) {
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/asset/balances", nil, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)
	var balanceResponse struct {
		Code    string           `json:"code"`
		Message string           `json:"msg"`
		Data    AssetBalanceList `json:"data"`
	}
	if err := response.DecodeJSON(&balanceResponse); err != nil {
		return nil, err
	}

	return balanceResponse.Data, nil
}

func (c *RestClient) AccountPositions() (string, error) {
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/account/positions", nil, nil)
	if err != nil {
		return "nil", err
	}

	response, err := c.sendRequest(req)

	//
	if err != nil {
		return "nil", err
	}

	//var postionsResponse struct {
	//	Code    string `json:"code"`
	//	Message string `json:"msg"`
	//	Data    string `json:"data"`
	//}

	//result := make([]map[string]interface{}, 0)
	//err = json.Unmarshal(response.Body, &result)
	//fmt.Println(result)

	value := gjson.Get(response.String(), "data")

	//println(value.String())

	//fmt.Println(response.DecodeJSON(&postionsResponse))
	//if err := response.DecodeJSON(&postionsResponse); err != nil {
	//	return "nil", err
	//}
	//fmt.Println(postionsResponse, err)
	return value.String(), err
}

func (c *RestClient) OrdersHistory() (string, error) {

	//data := map[string]interface{}{
	//
	//	"instType": "SWAP",
	//}
	//
	//payload, err := json.Marshal(data)

	var params = url.Values{}
	params.Add("instType", "SWAP")
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/trade/orders-history", params, nil)
	if err != nil {
		return "nil", err
	}

	response, err := c.sendRequest(req)
	fmt.Println(response)
	if err != nil {
		return "nil", err
	}

	//var postionsResponse struct {
	//	Code    string `json:"code"`
	//	Message string `json:"msg"`
	//	Data    string `json:"data"`
	//}

	//result := make([]map[string]interface{}, 0)
	//err = json.Unmarshal(response.Body, &result)
	//fmt.Println(result)

	value := gjson.Get(response.String(), "data")

	//println(value.String())

	//fmt.Println(response.DecodeJSON(&postionsResponse))
	//if err := response.DecodeJSON(&postionsResponse); err != nil {
	//	return "nil", err
	//}
	//fmt.Println(postionsResponse, err)
	return value.String(), err
}

func (c *RestClient) OrdersHistoryMax(symbol string, options *types.TradeQueryOptions) (string, error) {

	var params = url.Values{}
	params.Add("instType", "SWAP")
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/trade/orders-history-archive", params, nil)
	if err != nil {
		return "nil", err
	}

	response, err := c.sendRequest(req)
	fmt.Println(response)
	if err != nil {
		return "nil", err
	}

	//var postionsResponse struct {
	//	Code    string `json:"code"`
	//	Message string `json:"msg"`
	//	Data    string `json:"data"`
	//}

	//result := make([]map[string]interface{}, 0)
	//err = json.Unmarshal(response.Body, &result)
	//fmt.Println(result)

	value := gjson.Get(response.String(), "data")

	//println(value.String())

	//fmt.Println(response.DecodeJSON(&postionsResponse))
	//if err := response.DecodeJSON(&postionsResponse); err != nil {
	//	return "nil", err
	//}
	//fmt.Println(postionsResponse, err)
	return value.String(), err
}

// 市价平仓
func (c *RestClient) ClosePositions(direct string) (string, error) {

	data := map[string]interface{}{
		"instId":  "ETH-USDT-SWAP",
		"mgnMode": "cross",
		"posSide": direct,
	}

	payload, err := json.Marshal(data)
	req, err := c.newAuthenticatedRequest("POST", "/api/v5/trade/close-position", nil, payload)
	if err != nil {
		return "nil", err
	}

	response, err := c.sendRequest(req)
	fmt.Println("response", response)
	if err != nil {
		return "nil", err
	}

	//var postionsResponse struct {
	//	Code    string `json:"code"`
	//	Message string `json:"msg"`
	//	Data    string `json:"data"`
	//}

	//result := make([]map[string]interface{}, 0)
	//err = json.Unmarshal(response.Body, &result)
	//fmt.Println(result)

	value := gjson.Get(response.String(), "data")

	//println(value.String())

	//fmt.Println(response.DecodeJSON(&postionsResponse))
	//if err := response.DecodeJSON(&postionsResponse); err != nil {
	//	return "nil", err
	//}
	//fmt.Println(postionsResponse, err)
	return value.String(), err
}

// 市价平仓
func (c *RestClient) GetPendingAlgos() (string, error) { // 市价平仓

	query := url.Values{}

	query.Add("ordType", "conditional")
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/trade/orders-algo-pending", query, nil)

	if err != nil {
		return "nil", err
	}

	response, err := c.sendRequest(req)
	fmt.Println(response, "asdfasf")
	//if err != nil {
	//	return nil, err
	//}
	//

	value := gjson.Get(response.String(), "data.0.algoId")

	//var orderResponse struct {
	//	Code    string         `json:"code"`
	//	Message string         `json:"msg"`
	//	Data    []OrderDetails `json:"data"`
	//}
	//if err := response.DecodeJSON(&orderResponse); err != nil {
	//	return nil, err
	//}

	return value.String(), nil
}

func toLocalSymbol(symbol string) string {
	if s, ok := spotSymbolMap[symbol]; ok {
		return s
	}

	log.Errorf("failed to look up local symbol from %s", symbol)
	return symbol
}

// 批量撤单
func (c *RestClient) CancelPendingAlgos(symbol string) (string, error) {

	query := url.Values{}
	query.Add("ordType", "conditional")
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/trade/orders-algo-pending", query, nil)

	if err != nil {
		return "nil", err
	}
	symbol = toLocalSymbol(symbol)

	response, err := c.sendRequest(req)
	value := gjson.Get(response.String(), "data")
	var data []map[string]interface{}
	for _, order := range value.Array() {
		dt := map[string]interface{}{
			"instId": symbol,
			"algoId": order.Get("algoId").Int(),
		}
		//fmt.Println("%+v", dt)
		//fmt.Println("%+v", order)
		data = append(data, dt)
	}
	payload, err := json.Marshal(data)
	//fmt.Println(payload)
	req, err = c.newAuthenticatedRequest("POST", "/api/v5/trade/cancel-algos", nil, payload)
	if err != nil {
		return "nil", err
	}
	response, err = c.sendRequest(req)
	//fmt.Println("response", response)

	return response.String(), nil
}
func (c *RestClient) CancelAlgos(orderid string) (string, error) {

	var parameterList []map[string]interface{}

	data := map[string]interface{}{
		"instId": "ETH-USDT-SWAP",
		"algoId": orderid,
	}
	parameterList = append(parameterList, data)

	payload, err := json.Marshal(parameterList)
	fmt.Println(payload)
	req, err := c.newAuthenticatedRequest("POST", "/api/v5/trade/cancel-algos", nil, payload)
	if err != nil {
		return "nil", err
	}

	response, err := c.sendRequest(req)
	fmt.Println("response", response)
	if err != nil {
		return "nil", err
	}
	value := gjson.Get(response.String(), "data")

	return value.String(), err
}

type AssetCurrency struct {
	Currency               string           `json:"ccy"`
	Name                   string           `json:"name"`
	Chain                  string           `json:"chain"`
	CanDeposit             bool             `json:"canDep"`
	CanWithdraw            bool             `json:"canWd"`
	CanInternal            bool             `json:"canInternal"`
	MinWithdrawalFee       fixedpoint.Value `json:"minFee"`
	MaxWithdrawalFee       fixedpoint.Value `json:"maxFee"`
	MinWithdrawalThreshold fixedpoint.Value `json:"minWd"`
}

func (c *RestClient) AssetCurrencies() ([]AssetCurrency, error) {
	req, err := c.newAuthenticatedRequest("GET", "/api/v5/asset/currencies", nil, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var currencyResponse struct {
		Code    string          `json:"code"`
		Message string          `json:"msg"`
		Data    []AssetCurrency `json:"data"`
	}

	if err := response.DecodeJSON(&currencyResponse); err != nil {
		return nil, err
	}

	return currencyResponse.Data, nil
}

type MarketTicker struct {
	InstrumentType string `json:"instType"`
	InstrumentID   string `json:"instId"`

	// last traded price
	Last fixedpoint.Value `json:"last"`

	// last traded size
	LastSize fixedpoint.Value `json:"lastSz"`

	AskPrice fixedpoint.Value `json:"askPx"`
	AskSize  fixedpoint.Value `json:"askSz"`

	BidPrice fixedpoint.Value `json:"bidPx"`
	BidSize  fixedpoint.Value `json:"bidSz"`

	Open24H           fixedpoint.Value `json:"open24h"`
	High24H           fixedpoint.Value `json:"high24H"`
	Low24H            fixedpoint.Value `json:"low24H"`
	Volume24H         fixedpoint.Value `json:"vol24h"`
	VolumeCurrency24H fixedpoint.Value `json:"volCcy24h"`

	// Millisecond timestamp
	Timestamp types.MillisecondTimestamp `json:"ts"`
}

func (c *RestClient) MarketTicker(instId string) (*MarketTicker, error) {
	// SPOT, SWAP, FUTURES, OPTION
	var params = url.Values{}
	params.Add("instId", instId)

	req, err := c.newRequest("GET", "/api/v5/market/ticker", params, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var tickerResponse struct {
		Code    string         `json:"code"`
		Message string         `json:"msg"`
		Data    []MarketTicker `json:"data"`
	}
	if err := response.DecodeJSON(&tickerResponse); err != nil {
		return nil, err
	}

	if len(tickerResponse.Data) == 0 {
		return nil, fmt.Errorf("ticker of %s not found", instId)
	}

	return &tickerResponse.Data[0], nil
}

func (c *RestClient) MarketTickers(instType InstrumentType) ([]MarketTicker, error) {
	// SPOT, SWAP, FUTURES, OPTION
	var params = url.Values{}
	params.Add("instType", string(instType))

	req, err := c.newRequest("GET", "/api/v5/market/tickers", params, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var tickerResponse struct {
		Code    string         `json:"code"`
		Message string         `json:"msg"`
		Data    []MarketTicker `json:"data"`
	}
	if err := response.DecodeJSON(&tickerResponse); err != nil {
		return nil, err
	}

	return tickerResponse.Data, nil
}

func Sign(payload string, secret string) string {
	var sig = hmac.New(sha256.New, []byte(secret))
	_, err := sig.Write([]byte(payload))
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(sig.Sum(nil))
	// return hex.EncodeToString(sig.Sum(nil))
}
