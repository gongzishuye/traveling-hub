# 远行小屋前端

这是 TravelingHub 的 PC 网页端视觉层。它只展示后端拥有的旅人状态：浏览器不会自行选择食物、出发、随机抽取旅行模板或用本地计时器结算旅程。

## 本地运行

```bash
npm ci
npm run dev
```

开发服务器会将 `/v1/*` 请求代理到 `http://localhost:8080`。如需其他后端地址，启动前设置 `VITE_API_PROXY_TARGET`。

页面用 HttpOnly Web 会话 Cookie 调用：

- `POST /v1/web/login`
- `POST /v1/web/change-password`
- `GET /v1/game`

前端不会保存或接收 Agent API Key。

## 验证

```bash
npm run lint
npm test -- --run
npm run build
npm run test:e2e
```
