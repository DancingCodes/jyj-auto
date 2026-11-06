package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

const maxLoginAttempts = 5

func GetInfo(ctx context.Context) (json.RawMessage, error) {
	client, token, err := Login(ctx)
	if err != nil {
		return nil, err
	}
	result, err := client.GetInfo(ctx, token)
	if err != nil {
		return nil, errors.New("用户信息获取失败")
	}
	return result, nil
}

func BuildMachineExport(ctx context.Context, date string) ([]byte, error) {
	client, token, err := Login(ctx)
	if err != nil {
		return nil, err
	}
	result, err := client.GetMachineList(ctx, token, date, date)
	if err != nil {
		return nil, errors.New("机具列表获取失败")
	}
	var response struct {
		Data struct {
			Total int                `json:"total"`
			List  []map[string]any `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, errors.New("机具列表数据格式错误")
	}
	records := response.Data.List
	if records == nil {
		records = []map[string]any{}
	}
	fmt.Printf("获取机具列表：接口总数 %d，当前列表 %d 条\n", response.Data.Total, len(records))
	for index, record := range records {
		data, marshalErr := json.Marshal(record)
		if marshalErr != nil {
			fmt.Printf("第 %d 条机具数据打印失败：%v\n", index+1, marshalErr)
			continue
		}
		fmt.Printf("第 %d 条机具数据：%s\n", index+1, data)
	}
	file, err := BuildMachineExcel(records)
	if err != nil {
		return nil, errors.New("机具列表导出失败")
	}
	return file, nil
}

func Login(ctx context.Context) (*PCClient, string, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, "", errors.New("读取 PC 配置失败")
	}
	client := NewPCClient(config.PC.ChannelCode, config.PC.CFK)

	if _, err := client.CheckChannelFromKey(ctx); err != nil {
		return nil, "", errors.New("渠道识别失败")
	}
	if _, err := client.GetLoginConfig(ctx); err != nil {
		return nil, "", errors.New("登录配置获取失败")
	}

	var redirectURL string
	for range maxLoginAttempts {
		captchaResponse, err := client.CaptchaImage(ctx)
		if err != nil {
			return nil, "", errors.New("验证码获取失败")
		}
		var captcha struct {
			Image string `json:"img"`
			UUID  string `json:"uuid"`
		}
		if err := json.Unmarshal(captchaResponse, &captcha); err != nil || captcha.Image == "" || captcha.UUID == "" {
			return nil, "", errors.New("验证码响应数据无效")
		}
		code, err := RecognizeCaptcha(ctx, captcha.Image)
		if err != nil {
			return nil, "", errors.New("验证码识别失败")
		}
		loginResponse, err := client.Login(ctx, config.PC.Username, config.PC.Password, code, captcha.UUID)
		if err != nil {
			continue
		}
		var login struct {
			RedirectURL string `json:"redirectUrl"`
		}
		if err := json.Unmarshal(loginResponse, &login); err == nil && login.RedirectURL != "" {
			redirectURL = login.RedirectURL
			break
		}
	}
	if redirectURL == "" {
		return nil, "", errors.New("PC登录失败，验证码重试5次仍未成功")
	}
	redirect, err := url.Parse(redirectURL)
	if err != nil || redirect.Scheme != "https" || redirect.Host != "oagto.awillpay.com" || redirect.Path != "/login" {
		return nil, "", errors.New("登录跳转地址无效")
	}
	tk := redirect.Query().Get("tk")
	cfk := redirect.Query().Get("cfk")
	if tk == "" || cfk == "" {
		return nil, "", errors.New("登录跳转参数缺失")
	}
	token, err := client.ExchangeLoginToken(ctx, config.PC.Username, tk, cfk)
	if err != nil {
		return nil, "", errors.New("单点登录失败")
	}
	return client, token, nil
}
