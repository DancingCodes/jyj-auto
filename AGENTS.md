# jyj-auto 开发说明

## 项目用途

这是一个 Go HTTP 服务，用于访问聚赢家 PC 端接口。目前已接入登录前准备、登录和获取用户信息所需的上游接口；商户及交易查询接口将在拿到对应接口信息后接入。

## 目录结构

```text
cmd/api/main.go       服务启动入口
internal/pc.go        聚赢家 PC 上游接口客户端
internal/ocr.go       Go 到内部 ddddocr 服务的调用
internal/server.go    HTTP 路由
internal/service.go   登录及用户信息流程编排
../ddddocr/app.py     ddddocr 验证码识别服务
config/accounts.json  本机运行配置（真实账号不提交）
```

## 已接入的上游接口

- `POST https://parweb-api.awillpay.com/checkChannelFromKey`
- `GET https://parweb-api.awillpay.com/captchaImage`
- `GET https://parweb-api.awillpay.com/getLoginConfig`
- `POST https://parweb-api.awillpay.com/login`
- `POST https://oagto.awillpay.com/serv-api/checkChannelFromKey`
- `POST https://oagto.awillpay.com/serv-api/isLogin`
- `POST https://oagto.awillpay.com/serv-api/loginAccount`
- `GET https://oagto.awillpay.com/serv-api/getInfo`

以上上游接口仅由内部客户端和服务流程调用，不单独暴露为本项目的 HTTP 路由。

## 启动与测试

```bash
go run ./cmd/api
go test ./...
```

服务监听 `9102`。

本地开发时需另启 OCR 服务：

```bash
cd ..\ddddocr
pip install -r requirements.txt
python app.py
```

Go 服务固定连接 Docker 网络中的 OCR 服务 `http://ddddocr:9103`。

- `GET /health`：健康检查
- `GET /api/info`：读取配置中的 PC 账号，自动识别验证码、完成登录换票并获取用户信息
- `GET /api/machine-export?date=YYYY-MM-DD`：自动登录后查询机具并导出 Excel；请求固定使用 `pageSize=999`、`pageNo=1`；`date` 可省略，默认查询昨天

机具导出只在内存中生成 Excel 并直接返回 HTTP 响应，不在服务端落地文件。

## 配置

复制 `config/accounts.example.json` 为 `config/accounts.json`，填写 PC 用户名、密码、渠道码和 `cfk`。程序每次调用服务流程时重新读取配置。真实配置必须保持在 `.gitignore` 中，不得提交、写入日志或放入 Docker 镜像。对外部署时应通过防火墙或反向代理限制接口来源；登录接口使用配置中的 PC 账号。

## 开发约定

- `main.go` 只启动 HTTP 服务；路由放在 `internal/server.go`；业务编排放在 `internal/service.go`。
- 外部请求放在 `internal/pc.go`，不直接写进 HTTP handler。
- 不在代码、日志、提交记录或接口响应中暴露账号、密码、验证码、会话 token 或 `cfk`。
- 修改 Go 代码后运行 `gofmt -w .` 和 `go test ./...`。

## Docker 部署

```bash
docker build -t jyj-auto:latest .
```

发布时 ddddocr 单独部署并加入与 jyj-auto 相同的 Docker 网络；jyj-auto 通过 `http://ddddocr:9103` 访问 OCR 服务，不在本 Compose 中重复启动 OCR：

```yaml
docker compose up -d --build
```
