# ADMIN_PAYMENT_INTEGRATION_API

> 单文件中英双语文档 / Single-file bilingual documentation (Chinese + English)

---

## 中文

### 目标
本文档用于对接外部支付系统与 Sub2API，覆盖两类集成方式：
- 旧模式：外部支付页 / iframe + Admin API 充值
- 新模式：原生模型广场 + 支付渠道跳转 + 公共回调完单

如果你正在基于 `sub2api` 做二次开发，推荐优先使用“原生商业化模式”，因为它：
- 不需要硬编码支付渠道
- 支持多个支付提供方并行配置
- 下单、订单、钱包、权益发放都走站内链路
- 对后续 upstream 更新更友好

---

### 1. 两种集成模式

#### 1.1 旧模式：外部充值页
适用于你已经有单独支付站点（例如 `sub2apipay`）的情况：
- 后台配置 `purchase_subscription_url`
- 前端用 iframe 或新窗口打开该页面
- 支付成功后，由外部系统调用 Admin API 进行充值 / 发卡 / 余额调整

#### 1.2 新模式：原生商业化
适用于你希望直接在 Sub2API 内完成：
- 模型广场展示
- 商品购买
- 支付渠道选择
- 订单创建
- 支付回调
- 余额充值 / 订阅发放 / 可用分组授权

---

### 2. 原生商业化模式配置

#### 2.1 后台设置项
在后台系统设置中开启以下项目：
- `purchase_subscription_enabled = true`
- `native_purchase_mode = "native"`
- `native_marketplace_enabled = true`
- `native_orders_enabled = true`
- `native_wallet_enabled = true`（可选但推荐）

同时配置：
- `commerce_callback_secret`
- `commerce_payment_providers`

其中：
- `commerce_callback_secret` 用于校验公开支付回调请求头
- `commerce_payment_providers` 用于定义支付渠道列表

#### 2.2 支付渠道配置结构
`commerce_payment_providers` 是一个 JSON 数组，每个元素结构如下：

```json
[
  {
    "code": "alipay",
    "name": "支付宝",
    "description": "适合中国大陆用户扫码支付",
    "icon": "https://pay.example.com/assets/alipay.png",
    "checkout_url": "https://pay.example.com/alipay/checkout?order_no={order_no}&amount={amount}&user_id={user_id}",
    "enabled": true,
    "sort_order": 10
  },
  {
    "code": "wechat_pay",
    "name": "微信支付",
    "description": "适合微信内与扫码支付场景",
    "icon": "https://pay.example.com/assets/wechat-pay.png",
    "checkout_url": "https://pay.example.com/wechat/checkout?order_no={order_no}&amount={amount}&currency={currency}",
    "enabled": true,
    "sort_order": 20
  },
  {
    "code": "yipay",
    "name": "易支付",
    "description": "适合聚合支付网关接入",
    "icon": "https://pay.example.com/assets/yipay.png",
    "checkout_url": "https://pay.example.com/yipay/submit?out_trade_no={order_no}&money={amount}&type={payment_provider}",
    "enabled": false,
    "sort_order": 30
  }
]
```

#### 2.3 配置校验规则
- `code` 必填，格式：`^[a-z0-9][a-z0-9_-]{0,31}$`
- `name` 必填
- `code` 不能重复
- `enabled = true` 时，`checkout_url` 必填
- `checkout_url` 必须能生成合法的绝对 `http(s)` URL
- 用户侧只会看到安全字段：
  - `code`
  - `name`
  - `description`
  - `icon`
  - `sort_order`
- `checkout_url` 模板不会暴露给用户前端 catalog

#### 2.4 URL 模板变量
`checkout_url` 支持以下占位符：

| 变量 | 含义 |
|---|---|
| `{order_id}` | 订单 ID |
| `{order_no}` | 订单号 |
| `{user_id}` | 用户 ID |
| `{product_id}` | 商品 ID |
| `{price_id}` | 价格 ID |
| `{amount}` | 金额，固定两位小数 |
| `{currency}` | 币种，大写 |
| `{payment_provider}` | 当前支付渠道 code |

示例：

```text
https://pay.example.com/checkout?order_no={order_no}&amount={amount}&currency={currency}
```

创建订单后会自动展开成：

```text
https://pay.example.com/checkout?order_no=CO20260319000123&amount=99.00&currency=CNY
```

---

### 3. 用户侧原生下单流程

#### 3.1 获取模型广场
`GET /api/v1/commerce/catalog`

认证：
- `Authorization: Bearer <user-jwt>`

返回中会包含：
- 商品列表
- 模型广场数据
- `features.payment_providers`

示例响应片段：

```json
{
  "features": {
    "purchase_enabled": true,
    "purchase_mode": "native",
    "marketplace_enabled": true,
    "wallet_enabled": true,
    "orders_enabled": true,
    "legacy_purchase_url": "https://pay.example.com/legacy",
    "payment_providers": [
      {
        "code": "alipay",
        "name": "支付宝",
        "description": "适合中国大陆用户扫码支付",
        "icon": "https://pay.example.com/assets/alipay.png",
        "sort_order": 10
      }
    ]
  }
}
```

#### 3.2 创建订单
`POST /api/v1/commerce/orders`

认证：
- `Authorization: Bearer <user-jwt>`

请求头：
- `Content-Type: application/json`
- `Idempotency-Key: <unique-key>`（必填，推荐）

请求体：

```json
{
  "product_id": 101,
  "price_id": 1001,
  "payment_provider": "alipay"
}
```

curl 示例：

```bash
curl -X POST "${BASE}/api/v1/commerce/orders" \
  -H "Authorization: Bearer ${USER_JWT}" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: commerce-order-u123-p101-pr1001-001" \
  -d '{
    "product_id": 101,
    "price_id": 1001,
    "payment_provider": "alipay"
  }'
```

成功响应示例：

```json
{
  "id": 9527,
  "order_no": "CO20260319000123",
  "user_id": 123,
  "product_id": 101,
  "price_id": 1001,
  "status": "pending",
  "payment_status": "pending",
  "payment_provider": "alipay",
  "amount": 99,
  "currency": "CNY",
  "payment_action": {
    "provider": "alipay",
    "provider_name": "支付宝",
    "checkout_url": "https://pay.example.com/alipay/checkout?order_no=CO20260319000123&amount=99.00&user_id=123"
  }
}
```

#### 3.3 订单创建行为说明
- 如果只启用了 **一个** 支付渠道，前端即使不传 `payment_provider`，后端也会自动选中它
- 如果启用了 **多个** 支付渠道，前端应显式传 `payment_provider`
- 如果多渠道场景下不传 `payment_provider`，订单会以 `manual` 方式创建，通常不会返回 `payment_action.checkout_url`
- 前端拿到 `payment_action.checkout_url` 后，应直接跳转或新窗口打开支付页

---

### 4. 支付成功回调

#### 4.1 公共回调地址
`POST /api/v1/commerce/providers/:provider/callback`

例如：
- `POST /api/v1/commerce/providers/alipay/callback`
- `POST /api/v1/commerce/providers/wechat_pay/callback`

这是一个 **公开回调接口**，不需要用户 JWT，也不需要管理员 JWT，但必须携带共享密钥请求头。

#### 4.2 请求头
- `Content-Type: application/json`
- `X-Commerce-Callback-Secret: <your-secret>`

#### 4.3 请求体

```json
{
  "order_no": "CO20260319000123",
  "provider_trade_no": "ALI20260319008888",
  "paid_amount": 99.0,
  "paid_currency": "CNY",
  "request_payload": {
    "trade_status": "TRADE_SUCCESS"
  },
  "callback_payload": {
    "notify_id": "notify_123",
    "buyer_id": "2088xxxx"
  }
}
```

说明：
- `order_no` 必填
- `provider_trade_no` 可选
- `paid_amount` / `paid_currency` 可选
- `request_payload` / `callback_payload` 可为对象或字符串，服务端会统一转成字符串存储

curl 示例：

```bash
curl -X POST "${BASE}/api/v1/commerce/providers/alipay/callback" \
  -H "Content-Type: application/json" \
  -H "X-Commerce-Callback-Secret: ${COMMERCE_CALLBACK_SECRET}" \
  -d '{
    "order_no":"CO20260319000123",
    "provider_trade_no":"ALI20260319008888",
    "paid_amount":99.00,
    "paid_currency":"CNY",
    "request_payload":{"trade_status":"TRADE_SUCCESS"},
    "callback_payload":{"notify_id":"notify_123","buyer_id":"2088xxxx"}
  }'
```

#### 4.4 回调成功后的动作
回调校验通过后，系统会自动：
- 把订单标记为已支付 / 已完成
- 写入支付流水
- 根据商品权益发放内容执行：
  - 余额充值
  - 订阅发放
  - `allowed_group` 授权

#### 4.5 回调幂等行为
- 已完成订单再次回调时，不会重复发放权益
- 服务端会直接返回该订单当前状态
- 因此支付平台重试通知是安全的

---

### 5. 管理员手动补单

当支付平台未能正常回调，或者你需要人工补发时，可以使用：

`POST /api/v1/admin/commerce/orders/:id/manual-complete`

认证：
- `x-api-key: admin-<64hex>`，或管理员 JWT

请求头：
- `Content-Type: application/json`
- `Idempotency-Key: <unique-key>`

请求体示例：

```json
{
  "payment_provider": "alipay",
  "provider_trade_no": "ALI20260319008888",
  "paid_amount": 99.0,
  "paid_currency": "CNY",
  "request_payload": "{\"source\":\"manual\"}",
  "callback_payload": "{\"operator\":\"admin\"}"
}
```

这条接口会复用同一套完单逻辑，确保人工补单与自动回调行为一致。

---

### 6. 旧模式兼容：外部支付页 + Admin API

如果你当前仍使用外部支付页，可以继续沿用以下接口。

#### 6.1 一步完成创建并兑换
`POST /api/v1/admin/redeem-codes/create-and-redeem`

用途：原子完成“创建兑换码 + 兑换到指定用户”。

请求头：
- `x-api-key`
- `Idempotency-Key`

请求体示例：

```json
{
  "code": "s2p_cm1234567890",
  "type": "balance",
  "value": 100.0,
  "user_id": 123,
  "notes": "sub2apipay order: cm1234567890"
}
```

#### 6.2 查询用户（可选前置校验）
`GET /api/v1/admin/users/:id`

```bash
curl -s "${BASE}/api/v1/admin/users/123" \
  -H "x-api-key: ${KEY}"
```

#### 6.3 余额调整（已有接口）
`POST /api/v1/admin/users/:id/balance`

用途：人工补偿 / 扣减，支持 `set` / `add` / `subtract`。

请求体示例：

```json
{
  "balance": 100.0,
  "operation": "subtract",
  "notes": "manual correction"
}
```

#### 6.4 外部购买页 URL Query 透传
当 Sub2API 打开 `purchase_subscription_url` 或用户自定义 iframe 页面时，会统一追加：
- `user_id`
- `token`
- `theme`
- `lang`
- `ui_mode=embedded`

示例：

```text
https://pay.example.com/pay?user_id=123&token=<jwt>&theme=light&lang=zh&ui_mode=embedded
```

---

### 7. 对接建议

#### 7.1 推荐做法
- 支付侧保留自己的签名与验签逻辑
- 回调到 Sub2API 前，先在支付侧落库
- 回调 Sub2API 时始终带：
  - `order_no`
  - `provider_trade_no`
  - `paid_amount`
  - `paid_currency`
- 始终传 `Idempotency-Key` 调用写接口

#### 7.2 不推荐做法
- 在前端写死“支付宝/微信/Stripe”分支
- 把支付跳转链接硬编码进页面代码
- 直接信任浏览器回跳作为支付成功依据
- 在支付系统里自行改余额，但不回写 Sub2API 订单状态

#### 7.3 推荐的落地顺序
1. 先在后台配置 `commerce_payment_providers`
2. 用原生模型广场完成下单和跳转
3. 打通公共回调
4. 验证余额 / 订阅 / 分组权益发放
5. 最后再决定是否保留旧 `purchase_subscription_url`

---

## English

### Purpose
This document covers two payment integration styles for Sub2API:
- Legacy external payment page + Admin API fulfillment
- Native commerce marketplace + provider checkout + public callback completion

For new secondary development based on `sub2api`, the native commerce mode is recommended because it:
- avoids hardcoded payment providers
- supports multiple providers via settings
- keeps orders, wallet, and entitlement delivery inside Sub2API
- is friendlier to future upstream updates

---

### 1. Two integration modes

#### 1.1 Legacy mode
Use your own external payment page and call Admin APIs after payment success.

#### 1.2 Native commerce mode
Use the built-in marketplace, native order creation, provider checkout redirection, callback completion, and entitlement delivery.

---

### 2. Native commerce configuration

Enable these settings in admin:
- `purchase_subscription_enabled = true`
- `native_purchase_mode = "native"`
- `native_marketplace_enabled = true`
- `native_orders_enabled = true`
- `native_wallet_enabled = true` (optional but recommended)

Also configure:
- `commerce_callback_secret`
- `commerce_payment_providers`

Provider config is a JSON array. Example:

```json
[
  {
    "code": "alipay",
    "name": "Alipay",
    "description": "QR payment for mainland China users",
    "icon": "https://pay.example.com/assets/alipay.png",
    "checkout_url": "https://pay.example.com/alipay/checkout?order_no={order_no}&amount={amount}&user_id={user_id}",
    "enabled": true,
    "sort_order": 10
  }
]
```

Validation rules:
- `code` is required and must match `^[a-z0-9][a-z0-9_-]{0,31}$`
- `name` is required
- enabled providers must have `checkout_url`
- only safe provider fields are exposed in the public catalog
- checkout templates stay server-side only

Supported checkout template variables:
- `{order_id}`
- `{order_no}`
- `{user_id}`
- `{product_id}`
- `{price_id}`
- `{amount}`
- `{currency}`
- `{payment_provider}`

---

### 3. User checkout flow

#### 3.1 Get catalog
`GET /api/v1/commerce/catalog`

Auth:
- `Authorization: Bearer <user-jwt>`

The response includes `features.payment_providers`.

#### 3.2 Create order
`POST /api/v1/commerce/orders`

Headers:
- `Authorization: Bearer <user-jwt>`
- `Content-Type: application/json`
- `Idempotency-Key: <unique-key>`

Request body:

```json
{
  "product_id": 101,
  "price_id": 1001,
  "payment_provider": "alipay"
}
```

Example response:

```json
{
  "order_no": "CO20260319000123",
  "payment_provider": "alipay",
  "payment_action": {
    "provider": "alipay",
    "provider_name": "Alipay",
    "checkout_url": "https://pay.example.com/alipay/checkout?order_no=CO20260319000123&amount=99.00&user_id=123"
  }
}
```

Behavior notes:
- if exactly one provider is enabled, the backend can auto-select it
- if multiple providers are enabled, the frontend should explicitly send `payment_provider`
- if omitted in a multi-provider scenario, the order is usually created as `manual` and no hosted checkout URL is returned

---

### 4. Payment success callback

Public callback endpoint:

`POST /api/v1/commerce/providers/:provider/callback`

Required header:
- `X-Commerce-Callback-Secret: <your-secret>`

Example request:

```json
{
  "order_no": "CO20260319000123",
  "provider_trade_no": "ALI20260319008888",
  "paid_amount": 99.0,
  "paid_currency": "CNY",
  "request_payload": {
    "trade_status": "TRADE_SUCCESS"
  },
  "callback_payload": {
    "notify_id": "notify_123"
  }
}
```

After successful verification, Sub2API will:
- mark the order as paid / completed
- create payment transaction records
- deliver balance, subscription, or allowed-group entitlements

Repeated callbacks are safe and do not double-deliver entitlements.

---

### 5. Manual admin fallback

If provider callback fails or you need manual recovery:

`POST /api/v1/admin/commerce/orders/:id/manual-complete`

Auth:
- `x-api-key: admin-<64hex>` or admin JWT

Headers:
- `Content-Type: application/json`
- `Idempotency-Key: <unique-key>`

This endpoint reuses the same completion pipeline as the public callback.

---

### 6. Legacy external-page compatibility

You can still keep the old flow:
- open `purchase_subscription_url`
- call admin APIs after payment success

Useful legacy APIs:
- `POST /api/v1/admin/redeem-codes/create-and-redeem`
- `GET /api/v1/admin/users/:id`
- `POST /api/v1/admin/users/:id/balance`

When Sub2API opens the external purchase page, it appends:
- `user_id`
- `token`
- `theme`
- `lang`
- `ui_mode=embedded`

---

### 7. Recommendations
- keep payment-side signature verification in your own payment service
- always send `order_no`, `provider_trade_no`, `paid_amount`, and `paid_currency`
- always use `Idempotency-Key` on write APIs
- do not hardcode provider branches in frontend code
- do not treat browser return URL as payment-success truth
