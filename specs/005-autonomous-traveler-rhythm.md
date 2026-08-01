# PRD：自主旅人日节律与远行小屋接入

**Feature Branch**：`005-autonomous-traveler-rhythm`  
**Created**：2026-08-01  
**Status**：Draft  
**Input**：旅人完全由后端控制；每天早晨 08:00 自动出门旅行，夜晚必定回家但归家时间不确定。Agent 与用户只读取旅人状态。保留远行小屋现有的视觉与固定旅行内容，并将其复制到 TravelingHub 仓库共同开发。

## Summary

- **问题**：远行小屋目前在浏览器内以“手动出发 + 45 秒本地计时”推进状态，刷新、离线、跨设备和多 Agent 场景均不可靠；TravelingHub 尚未提供旅行玩法状态。
- **目标**：由 TravelingHub 成为旅人日节律、旅行结果、事件和相册的唯一状态源；远行小屋只读取并呈现状态。
- **成功标准**：一个旅人在没有前端、用户或 Agent 在线的情况下，仍会在每天 08:00 出发并在当晚归家；浏览器再次打开时展示正确的庭院、时间线与旅行卡。

## Scope

### In Scope

- 每个已绑定旅人按 `Asia/Shanghai` 时区的本地日节律自动旅行。
- 每日 08:00 自动创建并启动一段旅行；重复调度不得创建第二段。
- 每次旅行在创建时锁定现有固定旅行模板、食物、出发时间和当日随机归家时间。
- 默认归家窗口为当日 18:00:00 至 22:00:00；具体归家时刻由服务端生成并持久化。
- 后端在出发、旅行中节点和归家时写入旅人可见的不可变事件；归家时创建旅行卡相册条目。
- 新增只读游戏快照接口，供远行小屋显示庭院、时间线、旅行卡和相册。
- Agent 首次接入时以邮箱注册；系统在一个事务中创建 Web 用户、Agent 和唯一旅人。
- 系统仅在首次成功注册响应中返回 Web 用户名、一次性初始密码和 Agent API Key。
- Web 用户以邮箱和密码登录远行小屋；浏览器会话解析到该用户绑定的 Agent 与旅人。
- 将 `/Users/faria/Documents/traveling-frogs/app` 复制为本仓库的 `frontend/`，排除 `node_modules`、`dist` 与任何 Git 元数据。
- 保留现有远行小屋的场景、美术、固定地点模板、旅行卡和组件结构。

### Out of Scope

- 用户或 Agent 手动选择食物、手动出发、手动重开旅行。
- 多旅人偶遇、书信、礼物、好友、公共天气和 LLM 叙事。
- 修改现有插画、美术资产或页面视觉布局。
- OAuth、第三方登录、支付、排行榜、通知推送和多用户管理同一 Agent。

### Assumptions

- 旅人的时区第一版统一为 `Asia/Shanghai`；未来可设为世界或旅人级配置。
- 每天只允许一段旅行；旅人若在服务启动、调度器停机或浏览器离线期间错过节点，恢复后必须补结算。
- 现有前端 45 秒计时仅是原型实现，不保留为真实玩法时间。页面视觉和旅行模板保留，但不再由 `useJourneyTimer` 或本地 reducer 决定状态。
- 当前后端内部实体仍名为 `frog`；对外 UI 和文案继续使用“旅人”。物种统一是后续产品决定。
- 第一版采用“一位 Web 用户对应一位 Agent、一位旅人”；未来多 Agent 管理需引入绑定表。
- 初始密码是高熵临时凭据，只在注册成功响应中出现一次；用户首次 Web 登录后必须修改密码。
- 公开部署前，邮箱必须完成验证后才可激活 Web 登录；开发环境可用显式开关跳过验证以便联调。

## Product Rules

| 时间/条件 | 后端状态变化 | 前端呈现 |
| --- | --- | --- |
| 每日 08:00 | 创建当天旅行，`home → travelling` | 旅行中庭院、已出发记录 |
| 旅行中 | 根据锁定模板产生固定旅行节点 | 现有时间线与旅行记录 |
| 当日锁定归家时刻 | `travelling → returned`，写入相册 | 归来庭院、旅行卡、相册新增 |
| 下一日 08:00 前 | `returned` 保持可阅读 | 归来状态与最新旅行卡 |

## User Flow

1. Agent 首次连接 Hub，提交邮箱；系统创建 Web 用户、Agent 与唯一旅人，返回一次性初始密码和 Agent API Key。
2. 用户使用邮箱与初始密码登录远行小屋，并在首次登录后修改密码。
3. Web 会话解析到绑定 Agent；到达每日 08:00 后，后端调度器为其旅人创建当天旅行并锁定模板与归家时刻。
4. 用户或 Agent 随时调用游戏快照接口；后端先补结算到当前时间，再返回状态。
5. 远行小屋按快照显示“在家 / 旅行中 / 归来”、旅行记录及相册。
6. 归家后，旅行卡作为已完成状态持续可见，直至下一日出发。

## Backend Design

### State authority

- 后端 MUST 是旅行状态、时间、模板选择、事件和相册的唯一裁决者。
- 浏览器 MUST NOT 提交完整游戏状态、完成时间、旅行卡或相册内容。
- 前端只能读取快照；第一版没有改变旅人行为的写接口。

### Identity and login

- `POST /v1/agent-registrations` MUST 接受有效邮箱，且在首次注册时原子创建 `users`、`agents` 和 `frogs`。
- Agent API Key 与 Web 密码 MUST 是不同凭据；二者均只在创建时返回明文，服务端分别保存安全哈希。
- Web 密码 MUST 使用专用密码哈希算法存储；初始密码 MUST 标记为需修改。
- `POST /v1/web/login` MUST 使用邮箱和密码创建 HttpOnly、Secure、SameSite 会话 Cookie；浏览器 MUST NOT 接收 Agent API Key。
- 后续 Web 游戏接口 MUST 通过 Web 会话取得 `user_id → agent_id → frog_id`；Agent API 接口继续通过 Agent API Key 取得同一 `agent_id`。
- 对已存在的邮箱，注册接口 MUST NOT 返回任何现有凭据、重置密码或重复创建 Agent；只返回非枚举性的“已提交/请登录或找回密码”结果。

### Scheduling

- 一个统一 Worker MUST 批量查询到期旅程，而不是为每个旅人创建内存定时器。
- `daily_journeys` MUST 保存 `local_date`、`departed_at`、`return_at`、`template_id`、`status` 和下一个待处理节点。
- `(frog_id, local_date)` MUST 唯一，保证 08:00 重试、并发 Worker 或重启时幂等。
- Worker MUST 以短周期批处理到期项；快照读取也 MUST 补结算遗漏的到期项。
- 归家时必须在一个事务中更新旅程、写事件、写相册，保证最多一次结算。

### API contract

首次 Agent 注册：

```json
POST /v1/agent-registrations
{ "email": "owner@example.com" }

201 Created
{
  "agent_id": "uuid",
  "frog_id": "uuid",
  "username": "owner@example.com",
  "initial_password": "returned-once-only",
  "agent_api_key": "returned-once-only",
  "must_change_password": true
}
```

Web 登录：`POST /v1/web/login` 接收邮箱与密码，成功后写入会话 Cookie。首次登录会话仅允许调用 `POST /v1/web/change-password`，改密成功后才能调用游戏接口。浏览器以该 Cookie 调用 `GET /v1/game`；Agent 以 API Key 调用既有 `/v1/me/*` 和未来的 Agent 快照接口。

`GET /v1/game` 是远行小屋的主接口，Web 会话认证后返回：

```json
{
  "frog_id": "uuid",
  "server_time": "2026-08-01T20:13:00+08:00",
  "local_date": "2026-08-01",
  "phase": "home | travelling | returned",
  "journey": {
    "template_id": "light-meal-willow-pond-02",
    "departed_at": "2026-08-01T08:00:00+08:00",
    "return_at": "2026-08-01T20:08:00+08:00"
  },
  "events": [{ "second": 0, "text": "旅人出发了。" }],
  "album_postcard_ids": ["willow-pond"]
}
```

- 接口 MUST 在读取前完成该旅人的到期状态结算。
- `return_at` 在旅行中可不返回给前端，以保留“归家时间不确定”的体验；返回后可作为历史数据提供。
- 既有 `GET /v1/me/events` 保留给 Agent 的增量事件消费，不取代 Web 游戏快照。

### Frontend adapter

- `frontend/` MUST 保持 Vite + React + TypeScript 独立可运行。
- `App.tsx` MUST 在加载时请求 `GET /v1/game`；旅行中以受控轮询刷新快照。
- 前端 MUST 用快照取代 `createInitialGame`、`startJourney`、`advanceJourney` 和 `startNextJourney` 的生产调用。
- `journeyCatalog` 和视觉目录继续按后端返回的 `template_id`、`postcard_id` 查找本地文案和资产；前端不得自行抽取模板。
- 旅行手账改为只读“今日旅程”信息，不显示可改变结果的出发、食物选择和重开控制。
- 开发环境 MUST 使用 Vite 代理或同源部署，避免浏览器直接跨域失败；浏览器只使用 Web 会话 Cookie，不得保存 Agent API Key 或开发 Bearer Token。

## Functional Requirements

- **FR-001**：系统 MUST 在每个本地日 08:00 为每个符合条件的旅人创建至多一段当天旅行。
- **FR-002**：系统 MUST 在创建旅行时锁定模板、出发时间和归家时间；后续读取不得改变它们。
- **FR-003**：系统 MUST 在归家时间到达时将状态更新为 `returned`，创建恰好一条相册记录与完成事件。
- **FR-004**：系统 MUST 在服务重启、Worker 延迟或旅人离线后补结算错过的状态转换。
- **FR-005**：`GET /v1/game` MUST 仅返回认证 Web 用户所绑定 Agent 的旅人状态。
- **FR-006**：前端 MUST 仅投影后端快照，MUST NOT 以本地时钟或随机数推进真实状态。
- **FR-007**：系统 MUST 保留当前固定地点模板及其对应旅行卡，不改变既有视觉内容。
- **FR-008**：系统 MUST 将前端源复制到 `frontend/`，且 MUST NOT 引入构建产物或依赖目录。
- **FR-009**：系统 MUST 在首次 Agent 邮箱注册时创建且仅创建一组 Web 用户、Agent 与旅人绑定，并返回一次性初始凭据。
- **FR-010**：系统 MUST 将 Web 密码与 Agent API Key 分开保存和验证，MUST NOT 向浏览器暴露 Agent API Key。
- **FR-011**：Web 用户 MUST 在首次登录后修改初始密码；重复邮箱注册 MUST NOT 泄露或重发任何凭据。
- **FR-012**：`GET /v1/game` MUST 从 Web 会话解析对应 Agent 与旅人，且不得接受浏览器指定的 `agent_id` 或 `frog_id`。
- **FR-013**：注册接口 MUST 拒绝无效邮箱，且失败、并发重复或部分失败 MUST NOT 留下孤立用户、Agent 或旅人记录。
- **FR-014**：初始密码会话 MUST 只能修改密码，MUST NOT 读取游戏快照、事件或相册。
- **FR-015**：Worker MUST 在 Agent 与 Web 用户均未发起请求时独立推进到期旅程。
- **FR-016**：旅行中的 Web 游戏快照 MUST NOT 暴露 `return_at`；旅人归来后才可在历史记录中返回该字段。
- **FR-017**：会话 Cookie MUST 使用 HttpOnly、Secure、SameSite 属性；HTTP 响应、前端存储和日志 MUST NOT 包含 Agent API Key 或初始密码。
- **FR-018**：前端在快照读取失败、会话失效或收到未知模板时 MUST 显示可理解的错误状态，MUST NOT 以本地随机数或本地时钟伪造旅行结果。

## Acceptance Criteria

- **AC-001**：给定一位旅人在 07:59 为 `home`，当系统结算至 08:00 后，存在唯一一段当天旅行且状态为 `travelling`。
- **AC-002**：给定一段已创建旅行，重复读取、重复 Worker 执行或服务重启后，其模板、出发时间和归家时间保持一致。
- **AC-003**：给定当前时间晚于锁定归家时间，读取快照后状态为 `returned`，相册恰好新增一张对应旅行卡。
- **AC-004**：给定另一 Agent 的凭证，读取接口不得获得该旅人的快照、事件或相册。
- **AC-005**：给定远行小屋加载任一后端快照，庭院、时间线和相册均能映射到现有组件与资产；前端不执行本地随机抽取或本地真实结算。
- **AC-006**：`frontend/` 不包含 `node_modules`、`dist` 或嵌套 Git 仓库，且可独立安装、测试和构建。
- **AC-007**：给定新邮箱注册，系统仅返回一次用户名、初始密码和 Agent API Key；用户以该密码登录后能看到同一旅人。
- **AC-008**：给定已注册邮箱，重复注册不创建第二个 Agent，不返回旧密码或 API Key。
- **AC-009**：给定 Web 会话，`GET /v1/game` 返回该会话绑定的旅人；篡改请求参数不能切换到另一旅人。
- **AC-010**：给定无效邮箱、并发相同邮箱或创建事务故障，系统不产生可用或孤立的身份记录。
- **AC-011**：给定初始密码登录，改密前访问游戏快照与事件被拒绝；改密后可正常读取。
- **AC-012**：给定 Agent 和用户均离线，Worker 仍在 08:00 出发并在锁定归家时间完成旅行。
- **AC-013**：给定 Agent API Key，Agent 只能读取自己旅人的增量事件；确认游标后不会重复收到已确认事件。
- **AC-014**：给定旅人旅行中，用户刷新、关闭并重新打开页面后，看到同一模板、同一阶段和正确累计节点，且不看到归家时间。
- **AC-015**：给定浏览器会话，响应 Cookie 满足安全属性，页面及浏览器存储中不存在 Agent API Key 或初始密码。
- **AC-016**：给定快照读取失败、会话失效或未知模板，前端显示恢复路径且不创建本地旅行、事件或相册内容。
- **AC-017**：给定用户整天离线后在次日 07:00 登录，看到昨日已归来记录；在次日 08:01 登录，看到新的当天旅行且昨日旅行卡仍在相册。
- **AC-018**：给定注册、登录和认证失败等请求，应用日志与非首次响应不包含初始密码或 Agent API Key 明文。

## Verification Plan

### Test setup and roles

- 所有后端时间测试 MUST 注入 `Asia/Shanghai` 时钟；不得依赖机器实际时间。
- API 集成测试 MUST 使用 PostgreSQL 与 Redis 容器；每个用例独立清理数据。
- **Agent 视角**只使用 Agent API Key，不使用 Web Cookie。
- **User 视角**只使用浏览器 Web 会话 Cookie，不使用 Agent API Key。
- E2E MUST 使用已迁入的 `frontend/`，以测试 API 服务或受控 mock 服务驱动页面；任何模拟快照必须与后端契约一致。

### Agent-view verification

| Case | AC | Scenario and steps | Expected result | Level | Evidence required |
| --- | --- | --- | --- | --- | --- |
| AG-01 | AC-007 | Agent 以新邮箱调用注册接口一次。 | 返回一次性用户名、初始密码、API Key；数据库中恰有一组 User→Agent→旅人绑定。 | API/DB | 脱敏响应、绑定断言 |
| AG-02 | AC-010 | Agent 提交格式错误邮箱；再模拟创建 User 后创建 Agent 失败。 | 返回验证/服务错误；事务回滚，不留下孤立记录。 | API/DB | 状态码、行数断言 |
| AG-03 | AC-008, AC-010 | 两个并发请求以同一邮箱注册；随后再次顺序注册。 | 最多一组绑定；后续响应不含密码、API Key、既有 ID 或可枚举信息。 | API/DB | 并发测试输出、脱敏响应 |
| AG-04 | AC-013 | Agent 用首次 API Key 查询身份、拉取事件、确认 cursor，再次拉取。 | 恒定对应同一旅人；确认前看到增量，确认后不重复投递。 | API | 请求/响应序列 |
| AG-05 | AC-004, AC-013 | Agent A 用自己的 Key 尝试读取/确认 Agent B 的 cursor 或资源。 | 仅能获得 A 的数据；跨 Agent cursor 被拒绝且 B 不受影响。 | API/DB | 双身份响应、游标断言 |
| AG-06 | AC-001, AC-012 | Agent 注册完成后停止一切请求；时钟跨越 08:00，Worker 独立运行。 | Worker 创建唯一当天旅行并写出发事件，不依赖 Agent 轮询。 | Worker/API | Worker 日志、旅程与事件断言 |
| AG-07 | AC-002, AC-003 | 对已出发旅行重复执行 Worker、重启服务并结算至归家后。 | 模板、出发/归家时刻不变；仅一张旅行卡和一次完成事件。 | Worker/DB | 重试日志、唯一性断言 |

### User-view verification

| Case | AC | Scenario and steps | Expected result | Level | Evidence required |
| --- | --- | --- | --- | --- | --- |
| US-01 | AC-007, AC-011 | 用户用初始密码登录；在改密前请求游戏页；提交符合规则的新密码后再次请求。 | 首次会话只能改密；改密后建立正常会话并看到与注册 Agent 相同的旅人。 | API/E2E | Cookie 流程、页面截图 |
| US-02 | AC-015 | 检查登录响应、浏览器 Cookie、localStorage/sessionStorage 与浏览器网络记录。 | Cookie 为 HttpOnly/Secure/SameSite；页面与浏览器存储没有 API Key 或初始密码。 | E2E/Manual | 头部断言、浏览器检查记录 |
| US-03 | AC-009 | 用户 A、B 各自登录；A 在 URL、请求体和开发者工具中伪造 B 的 Agent/Frog ID。 | `GET /v1/game` 始终返回 A 的旅人；不接受客户端指定身份。 | API/E2E | 双会话响应与网络记录 |
| US-04 | AC-001, AC-005 | 在 07:59 登录远行小屋，浏览器到 08:00 刷新快照。 | 页面从在家切换为旅行中；没有“选择食物”“出发”“重开”等可变更控制。 | E2E | home/travelling 截图与 DOM 断言 |
| US-05 | AC-014 | 旅行中刷新页面、关闭浏览器、在另一浏览器登录并重新打开。 | 所有页面显示同一模板、同一累计旅行节点和同一旅人；无本地重新抽取。 | E2E/API | 两浏览器快照、模板 ID 断言 |
| US-06 | AC-014, AC-016 | 旅行中检查 API 响应、时间线与页面文案。 | 可看到已发生节点，不能看到 `return_at`、精确归家倒计时或未解锁旅行卡。 | API/E2E | 响应 JSON、页面截图 |
| US-07 | AC-003, AC-005 | 将时钟推进到锁定归家时间后打开/刷新页面，再重复打开。 | 显示归来场景和一张旅行卡；旅行卡进入相册且重复刷新不重复新增。 | E2E/API | returned 截图、相册数断言 |
| US-08 | AC-017 | 用户整天不登录，在次日 07:00 登录。 | 首次快照已补结算昨日旅行，显示归来记录和旅行卡，且无丢失/重复记录。 | API/E2E | 跨日测试输出、页面截图 |
| US-09 | AC-017 | 用户整天不登录，在次日 08:01 登录。 | 昨日旅行卡仍在相册；后端已创建唯一一段当天旅行，页面显示旅行中。 | API/E2E | 跨日响应、页面截图 |
| US-10 | AC-016 | API 暂时失败、会话失效或返回未知模板 ID。 | 页面不伪造本地旅行结果；显示可理解的重试/重新登录/内容不可用状态，保留最后合法快照。 | E2E | 错误状态截图、控制台无未处理异常 |

### System boundary verification

| Case | AC | Scenario and steps | Expected result | Level | Evidence required |
| --- | --- | --- | --- | --- | --- |
| SY-01 | AC-006 | 检查 `frontend/`，执行依赖安装、单元测试、E2E 与生产构建。 | 没有 `node_modules`、`dist` 或嵌套 Git；源项目可独立通过。 | Build/E2E | 命令输出 |
| SY-02 | AC-002, AC-012 | Worker 在 07:59、08:00、归家边界、23:59 和跨日 08:00 运行；每个时点重复执行。 | 所有状态转换幂等，按 `Asia/Shanghai` 日历日生成，未产生重复旅程。 | Unit/Integration | 固定时钟测试输出 |
| SY-03 | AC-004, AC-009 | 使用无 Cookie、过期 Cookie、无效 API Key 和错误方法访问 Web/Agent 接口。 | 返回统一认证错误；不泄露用户、Agent、旅人或凭据存在性。 | API/Security | 负向测试响应 |
| SY-04 | AC-018 | 捕获注册、登录、重复注册和认证失败的应用日志与响应。 | 仅首次成功注册响应包含一次性凭据；日志及其他响应均不含凭据明文。 | API/Security | 脱敏日志审查、响应断言 |

## Acceptance Evidence Report

- **Pass condition**：所有 AG、US、SY 用例通过；任何时间边界、凭据、隔离或重复结算异常必须有明确修复记录。
- **Report fields**：Case、角色、AC、结果、证据路径、测试人、日期、备注。
- **Evidence archive**：`docs/qa/005-autonomous-traveler-rhythm/{agent,user,system}/`。

## Open Items

- 默认归家窗口暂定 18:00–22:00，是否需要按季节、地点或旅人偏好变化。
- 每日固定旅行模板是否允许重复，以及首次旅行的默认模板策略。
- 邮箱验证邮件、密码找回邮件的供应商与发送域名；公开环境启用前必须确定。
- `returned` 状态是否在次日 08:00 前自动回到 `home`，还是只由下一段旅行覆盖；本 PRD 采用后者。
