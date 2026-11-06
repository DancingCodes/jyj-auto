package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const parwebAPI = "https://parweb-api.awillpay.com"
const oagtoAPI = "https://oagto.awillpay.com"
const accountsPath = "config/accounts.json"
const machineListURL = oagtoAPI + "/serv-api/2.0/machine/getList"

type Config struct {
	PC PCAccount `json:"pc"`
}

type PCAccount struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ChannelCode string `json:"channelCode"`
	CFK         string `json:"cfk"`
}

func LoadConfig() (Config, error) {
	raw, err := os.ReadFile(accountsPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return Config{}, err
	}
	if config.PC.Username == "" || config.PC.Password == "" || config.PC.ChannelCode == "" || config.PC.CFK == "" {
		return Config{}, io.ErrUnexpectedEOF
	}
	return config, nil
}

type PCClient struct {
	channelCode string
	cfk         string
	httpClient  *http.Client
}

func NewPCClient(channelCode, cfk string) *PCClient {
	jar, _ := cookiejar.New(nil)
	return &PCClient{
		channelCode: channelCode,
		cfk:         cfk,
		httpClient:  &http.Client{Timeout: 25 * time.Second, Jar: jar},
	}
}

func (c *PCClient) CheckChannelFromKey(ctx context.Context) (json.RawMessage, error) {
	body, err := json.Marshal(struct {
		CFK string `json:"cfk"`
	}{CFK: c.cfk})
	if err != nil {
		return nil, err
	}
	return c.request(ctx, parwebAPI+"/checkChannelFromKey", http.MethodPost, bytes.NewReader(body), "")
}

func (c *PCClient) CaptchaImage(ctx context.Context) (json.RawMessage, error) {
	return c.request(ctx, parwebAPI+"/captchaImage", http.MethodGet, nil, "")
}

func (c *PCClient) GetLoginConfig(ctx context.Context) (json.RawMessage, error) {
	return c.request(ctx, parwebAPI+"/getLoginConfig", http.MethodGet, nil, "")
}

func (c *PCClient) Login(ctx context.Context, username, password, code, uuid string) (json.RawMessage, error) {
	body, err := json.Marshal(struct {
		Code     string `json:"code"`
		Password string `json:"password"`
		Username string `json:"username"`
		UUID     string `json:"uuid"`
	}{
		Code:     code,
		Password: password,
		Username: username,
		UUID:     uuid,
	})
	if err != nil {
		return nil, err
	}
	return c.request(ctx, parwebAPI+"/login", http.MethodPost, bytes.NewReader(body), "")
}

func (c *PCClient) ExchangeLoginToken(ctx context.Context, username, tk, cfk string) (string, error) {
	c.setOAGTCookies(username, cfk)

	checkBody, err := json.Marshal(struct {
		CFK string `json:"cfk"`
	}{CFK: cfk})
	if err != nil {
		return "", err
	}
	if _, err := c.request(ctx, oagtoAPI+"/serv-api/checkChannelFromKey", http.MethodPost, bytes.NewReader(checkBody), ""); err != nil {
		return "", err
	}

	isLoginBody, err := json.Marshal(struct {
		ChannelCode string `json:"channelCode"`
		TK          string `json:"tk"`
	}{ChannelCode: c.channelCode, TK: tk})
	if err != nil {
		return "", err
	}
	isLoginResponse, err := c.request(ctx, oagtoAPI+"/serv-api/isLogin", http.MethodPost, bytes.NewReader(isLoginBody), "")
	if err != nil {
		return "", err
	}
	var loginStatus struct {
		Data []struct {
			UserID int64 `json:"userId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(isLoginResponse, &loginStatus); err != nil || len(loginStatus.Data) == 0 || loginStatus.Data[0].UserID == 0 {
		return "", io.ErrUnexpectedEOF
	}

	loginAccountBody, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
		UserID   int64  `json:"userId"`
		TK       string `json:"tk"`
	}{UserID: loginStatus.Data[0].UserID, TK: tk})
	if err != nil {
		return "", err
	}
	loginAccountResponse, err := c.request(ctx, oagtoAPI+"/serv-api/loginAccount", http.MethodPost, bytes.NewReader(loginAccountBody), "")
	if err != nil {
		return "", err
	}
	var tokenResponse struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(loginAccountResponse, &tokenResponse); err != nil || tokenResponse.Data == "" {
		return "", io.ErrUnexpectedEOF
	}
	return tokenResponse.Data, nil
}

func (c *PCClient) GetInfo(ctx context.Context, token string) (json.RawMessage, error) {
	return c.request(ctx, oagtoAPI+"/serv-api/getInfo", http.MethodGet, nil, token)
}

func (c *PCClient) GetMachineList(ctx context.Context, token, startTime, endTime string) (json.RawMessage, error) {
	form := url.Values{
		"goldPartner":      {""},
		"isAgents":         {""},
		"startSn":          {""},
		"endSn":            {""},
		"status":           {""},
		"activeStatus":     {""},
		"kfFlag":           {""},
		"doneKfFlag":       {""},
		"freezeFlag":       {""},
		"startTime":        {startTime},
		"endTime":          {endTime},
		"pMobile":          {""},
		"bSn":              {""},
		"businessMen":      {""},
		"pageSize":         {"999"},
		"pageNo":           {"1"},
		"projectType":      {""},
		"selfMachine":      {""},
		"createStartTime":  {""},
		"createEndTime":    {""},
		"selfMacStartTime": {""},
		"selfMacEndTime":   {""},
		"doneDbFlag":       {""},
		"machineType":      {""},
		"bindStatus":       {""},
		"model":            {""},
		"modelList":        {""},
	}
	return c.request(ctx, machineListURL, http.MethodPost, strings.NewReader(form.Encode()), token)
}

func (c *PCClient) setOAGTCookies(username, cfk string) {
	endpoint, _ := url.Parse(oagtoAPI)
	c.httpClient.Jar.SetCookies(endpoint, []*http.Cookie{
		{Name: "cfk", Value: cfk, Path: "/"},
		{Name: "channelCode", Value: c.channelCode, Path: "/"},
		{Name: "username", Value: username, Path: "/"},
		{Name: "sidebarStatus", Value: "0", Path: "/"},
	})
}

func (c *PCClient) request(ctx context.Context, endpoint, method string, body io.Reader, token string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("channelcode", c.channelCode)
	if body != nil {
		contentType := "application/json;charset=UTF-8"
		if endpoint == machineListURL {
			contentType = "application/x-www-form-urlencoded"
		}
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("上游接口返回 HTTP %d", resp.StatusCode)
	}
	if !json.Valid(responseBody) {
		return nil, io.ErrUnexpectedEOF
	}
	return json.RawMessage(responseBody), nil
}
