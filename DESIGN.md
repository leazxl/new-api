# DESIGN.md · PURESNAKE 设计系统 v0.3-draft · 独立版

> **版本**：v0.3-draft · **日期**：2026-04-22 · **owner**：阿琛（俞俊琛）
>
> **状态**：独立版，已移除外部文档入口、本地素材索引与跨文件引用。
>
> **本版亮点**：
> - 明确品牌色演进、灰阶命名、红色演进
> - 新增"品牌色演进"叙事
> - 色彩体系完整重写（`3Grey1-5` / Snake-red 3 代 / 新旧 BrandGreen 并存 / SnakeDark）
> - 补充字体家族（DINPro / OPPOSans40_Light）
> - 新增"画板尺寸规范"章节（iOS @2x / @3x / Web / 海报各自规格）
> - 按钮体系、圆角、禁用色、字号全部修正

---

## 开篇 · 设计哲学

> **上下文优先于想象力。一致性优先于花活。**
>
> 好的高保真设计**不从零开始** —— 根植于现有设计上下文。从零搭建是最后手段（LAST RESORT），会导致糟糕设计。
>
---

## 一、视觉主题与氛围

### 一句话品牌故事

**PURESNAKE · 潮流循环交易的专业鉴别者**

### 品牌调性 × 5

- **专业** —— 鉴别业务是信任经济
- **潮感** —— 年轻用户，不奢侈不土
- **干练** —— 高对比、去装饰
- **科技** —— 黑 + 亮绿 + DIN 数字组合
- **循环** —— 二手/回收品类专属

### 反调性 × 4

- 不奢侈品（不金色镀光）
- 不土（不聚划算浮夸）
- 不可爱（IP 除外）
- 不苹果无色（品牌绿+红必须出现）

### 品牌色演进轨迹（★ 新版）

阿琛六年在文件里的命名证据链显示了**明确的品牌色演进**：

```
2020 极简期      · 纯白黑字 / 无系统化色板
2021 体系建立期  · 组件库 1.0 建立 BrandGreen #2EBD7C + BrandBlue #26273A
                   灰阶用 Grey1/2/3 简化 3 档
                   Snake-red = #FE0832 (Red1)
                   ↓
2024 新版升级期  · 新版品牌绿 #06D290（命名"新版品牌绿"，取代 #2EBD7C 作为主 CTA）
                   灰阶扩展到 3Grey1-5 完整 5 档
                   SnakeDark #111111 引入（暗色主题 / IP / 海报深底）
                   SnakeRed 品牌红迭代到 #FF3367（与 #FE0832 并存）
                   ↓
2026 IP 时代     · 新 Logo 黑底亮绿渐变
                   荣德兔 3D IP 主视觉
                   IP 字体矩阵扩展（DINPro / OPPOSans40_Light 加入）
```

**锚点选择**（v0.3 原则）：
- 以 **2024-2026 版本**为主干规范
- 保留既有组件命名习惯作为迁移期参考
- 旧版 Swatches（#2EBD7C 等）在迁移期保留共存

### 整体叙事

PURESNAKE 视觉语言 = 「**潮流循环的专业鉴别者**」。

以 `#06D290` 新版品牌绿承载主 CTA / 确认感、`#26273A` 品牌深墨蓝承载专业感 / 深色底、`#111111 SnakeDark` 承载暗色主题 / IP 底、`#FE0832 / #FF3367` 双红承载警示与价格、`#7A3BFF Pupple1` 承载潮玩活动。中文 PingFangSC 全字重、数字 DIN / DINPro、Logo 与装饰 OPPOSans，IP 场景并用 UnboundedSans。六年沉淀：**高对比、去装饰、数字用 DIN、中文用苹方** —— 走「**精准干练 + 科技潮流 + IP 辅助**」。

---

## 二、色彩体系 ★ 完整重写

### 2.1 品牌核心色（Brand Core）

#### 主品牌色（新旧并存）

| HEX | 命名 | 语义 | 场景 |
|---|---|---|---|
| **`#06D290`** | **新版品牌绿**（BrandGreen v2）| 主 CTA / 成功 / 鉴别通过 / 加载态 | **新项目默认用**，全局 CTA 按钮、寄售流程节点完成态 |
| `#2EBD7C` | 旧版品牌绿（BrandGreen v1 / `0BrandGreen`）| Ghost 按钮边 / 迁移期兼容 | 组件库 1.0 老项目、"我要寄出"等次要鼓励行动 |
| **`#26273A`** | **BrandBlue** / `0BrandBlue` / `1HaFine-Branddarkblue` | 深品牌色 / 主文字 / 深色底 / 大号主按钮 | Logo 底、APP 深 header、Button/big702/Primary |
| **`#111111`** | **SnakeDark** | 暗色主题 / IP 底 / 海报深底 | 新 Logo 方框、IP 场景、活动 KV 黑底 |

**核心规则**：
- 新版项目的主 CTA = `#06D290`
- 产品次要 Ghost 按钮边 = `#2EBD7C`（旧版兼容）
- 专业感 / 深色大号按钮 = `#26273A` BrandBlue
- IP / Logo / 深底场景 = `#111111` SnakeDark

### 2.2 Snake 红系列（三代并存，场景化使用）

| HEX | 命名 | 语义 |
|---|---|---|
| `#FE0832` | `Red1` / `Snake-red` | **警示红**（价格 / 优惠券 / 错误 / 鉴别不通过）|
| **`#FF3367`** | **`SnakeRed`** | **品牌红**（KV / 节日 / 品牌情绪红）|
| `#FD5B5B` | `新增品牌色_red` | 弱警示红（二次确认 / 轻提示）|
| `#FF132B` | `FF132B` | 加强警示（极端情况）|
| `#FF525D` | `5Red1` | 粉橙辅助红 |

**使用规则**：
- 价格/警示 → `#FE0832`
- 品牌 KV 大红色 → `#FF3367`
- 其他红变体仅在迁移期使用

### 2.3 辅助品牌色

| HEX | 命名 | 语义 |
|---|---|---|
| `#7A3BFF` | **Pupple1**（保留原拼写）| 潮玩 / 活动 / 专家鉴别 |
| `#2D57E7` | Blue1 | 次要链接 / Info |
| `#FFE3E3` / `#FFEAEE` | Red2 / Red3 | 浅警示背景 |
| `#EAF1FF` | Blue2 | 浅蓝背景 |
| `#FF6A0C` | — | 警告橙 / 待处理状态 |
| `#F4E8DE` + `#83684F` | Gold（会员价胶囊）| 浅暖米底 + 暖棕字 |
| `#00FF67` / `#20FF7E` / `#FDE834` | — | 活动 KV 高饱和绿黄（海报/招商专用）|

### 2.4 中性色阶 · 完整 5 档（**正式采用 3Grey1-5 系统**）

| HEX | 命名 | 语义 |
|---|---|---|
| `#0F1113` / `#111111 SnakeDark` | Text-Dark | 主标题 / 最深文字 |
| `#201614` | Text-Darker | 深色正文 |
| `#404155` | Text-2 | 副标题 / 次级文字 |
| `#707184` | **3Grey1** | 文字三级 / 辅助文字 |
| `#B1B2C1` | **3Grey2** | 中灰 |
| `#DBDCE2` | **3Grey3** | 深分割线 |
| `#E8E8E8` | **3Grey4** | 主分割线 / 边框 |
| `#F6F6F6` | **3Grey5** | 浅灰背景 / 输入框底 |
| `#888891` / `#A8AAB0` / `#ACACB7` | 占位灰 / 禁用灰 | 占位文字 / 按钮禁用态 |
| `#D0D3D5` | — | 按钮普通禁用 fill |
| `#FFFFFF` | — | 纯白 |

**阴影色**：`#07070734`（柔和 alpha 0.2）· `#F6F6F699`（半透明遮罩 alpha 0.6）· `#0000001A`（Switch knob 阴影）

### 2.5 语义色

| 角色 | HEX | 说明 |
|---|---|---|
| Success | `#06D290` | 新版品牌绿 |
| Warning | `#FF6A0C` | 警告橙 |
| Error / Danger | `#FE0832` | Snake-red 警示 |
| Info | `#2D57E7` | Blue1 |

### 2.6 PURESNAKE 鉴别业务专属语义

| 状态 | HEX | 对应按钮/徽章 |
|---|---|---|
| 鉴别通过 | `#06D290` | 新版品牌绿 |
| 鉴别不通过 | `#FE0832` | Snake-red |
| 待鉴别 | `#FF6A0C` | 警告橙 |
| 专家鉴别 | `#7A3BFF` | Pupple1 |
| 机器鉴别 | `#2D57E7` | Blue1 |

### 2.7 色彩速查卡（给 Agent 用）

```
/* PURESNAKE Palette · v0.3 */

/* Brand Core */
--brand-green-new:   #06D290;   /* 新版品牌绿 · 主 CTA / 成功 / 鉴别通过 */
--brand-green-old:   #2EBD7C;   /* 旧版品牌绿 · Ghost 按钮边 / 兼容 */
--brand-blue:        #26273A;   /* BrandBlue · 深文字 / 深按钮 / Logo 底 */
--snake-dark:        #111111;   /* SnakeDark · IP / 海报深底 */

/* Snake Red */
--snake-red:         #FE0832;   /* Red1 · 价格 / 警示 / 鉴别不通过 */
--snake-red-brand:   #FF3367;   /* SnakeRed · 品牌情绪红 (KV / 节日) */

/* Aux */
--pupple-1:          #7A3BFF;   /* 潮玩活动 */
--blue-1:            #2D57E7;   /* 次要链接 */
--warning:           #FF6A0C;   /* 警告橙 */
--gold-bg:           #F4E8DE;   /* 会员价胶囊底 */
--gold-text:         #83684F;   /* 会员价胶囊字 */

/* Neutral */
--text-1:            #0F1113;
--text-2:            #404155;
--grey-1:            #707184;   /* 3Grey1 */
--grey-2:            #B1B2C1;   /* 3Grey2 */
--grey-3:            #DBDCE2;   /* 3Grey3 */
--grey-4:            #E8E8E8;   /* 3Grey4 · 主分割线 */
--grey-5:            #F6F6F6;   /* 3Grey5 · 浅灰底 */
--disabled:          #D0D3D5;   /* 按钮禁用 */
--disabled-terminal: #ACACB7;   /* 终态禁用 */
```

### 2.8 色彩 Don't

- ❌ pure black `#000000` 作文字（用 `#0F1113` / `#111111`）
- ❌ 同屏 4+ 种高饱和品牌色
- ❌ 新版绿 `#06D290` + 旧版绿 `#2EBD7C` 并用（选一个时代）
- ❌ 任何品牌色 + 深金渐变（走向奢侈品）
- ❌ 自创 HEX（用上面色板派生 OKLCH）
- ❌ B 端临时色板（`1.Blue` / `2.Green` 等）误用到产品 UI

---

## 三、排版规则 ★ 重写

### 3.1 字体家族（完整版）

| 字体 | 字重 | 用途 |
|---|---|---|
| **PingFangSC** | Regular / Medium / Light / Semibold / Bold | **中文主字体**（全场景）|
| **DIN** | Regular / Medium / Bold | **数字/价格/时间**标准字体 |
| **DINPro** | Bold | 数字大号装饰（IP / 排行榜）|
| **OPPOSans** | Regular / Medium / Bold | Logo + 装饰英文 + 排行榜 / ROUND2 徽章 |
| **OPPOSans40_Light** | Light 46 号 | **仅 IP 页面大号装饰字** |
| **UnboundedSans** | Regular | 装饰英文 / 状态标识（如"未通过" "待验货"）|
| SFProText | Semibold | iOS 系统字体（status bar）|

**CSS 回退链**：

```css
/* 中文 */
font-family: "PingFang SC", "Microsoft YaHei", "Hiragino Sans GB", sans-serif;

/* 数字/英文（价格/时间） */
font-family: "DIN", "DIN Alternate", "Helvetica Neue", sans-serif;

/* IP / Logo 专用 */
font-family: "OPPOSans", "PingFang SC", sans-serif;
```

### 3.2 画板尺寸规范（★ 新增章节）

阿琛实战使用**多种画板标准**，字号计量方式不同：

| 场景 | 画板宽 | 字号 → CSS 规则 |
|---|---|---|
| **iOS APP @2x** | 750 (750×1334 / 750×1696 / 750×2061) | **Sketch ÷ 2 = CSS px** |
| **iOS APP @3x** | 1125 / 1242 | **Sketch ÷ 3 = CSS px** |
| **Web 1x** | 375 / 1280 | **Sketch = CSS px**（不除）|
| **海报 / 朋友圈** | 500×400 / 750×1624 / 1503×4264 | **原尺寸输出**（不转换）|
| **IP 宣传** | 1080×1956 / 1104×1858 | 社交媒体标准尺寸，原尺寸输出 |
| **PPT** | 1280×720 | 原尺寸（16:9 幻灯片）|
| **iOS Icon** | 48×48 / 54×54 / 34×34 / 28×28 (@2x) | ÷ 2 得 CSS 24/27/17/14 |

**Agent 使用时的判断**：
- 默认假设 iOS @2x 750 宽设计稿 → 除 2 得 CSS
- 如果用户明确说"Web/Landing" → 不除
- 如果用户明确说"海报" → 原尺寸

### 3.3 字号层级（按 iOS @2x 750 宽场景）

| 角色 | Sketch @1x | CSS | 字重 | 行高 | 场景 |
|---|---|---|---|---|---|
| H1 主标题 | 36 | 18 | PingFangSC Medium | 1.3 | 页面顶部标题 |
| H2 区块 | 32 | 16 | PingFangSC Medium | 1.3 | 卡片组标题 |
| H3 副标题 | 28 | 14 | PingFangSC Medium | 1.4 | 列表项主文字 |
| **Body** ★ 最高频 | **24-28** | **12-14** | PingFangSC Regular | 1.5 | 主要正文 |
| Caption | 22 | 11 | PingFangSC Regular | 1.5 | 说明 / 备注 |
| Meta | 20 | 10 | PingFangSC Regular | 1.5 | 最小辅助 |

**修正**：Body 最高频**不是单一字号**，产品 UI 是 24，组件精规范是 28。两档都是 Body，根据密度场景选用。

### 3.4 特殊场景字号

| 场景 | 字号 | 字体 |
|---|---|---|
| 价格特大 | 52+ | DIN-Bold |
| 海报装饰字 | 82+ (原尺寸) | PingFangSC Regular / OPPOSans |
| Logo | 166-179 | DIN-BoldItalicAlt + OPPOSans |
| IP 页面超大装饰 | 46 | OPPOSans40_Light |

### 3.5 IP 场景字体矩阵（★ 新增）

IP 页面字体**比产品 UI 复杂**，阿琛刻意用多字体做"炫感":

```
主英文标题     OPPOSans-B / 20-36
装饰英文次标    UnboundedSans-Regular / 22-26
数字装饰大号    DINPro-Bold / 25-31
大号装饰字      OPPOSans40_Light / 46
中文副标题      PingFangSC Medium / 36
```

**规则**：IP 场景**可以 7-8 种字体并用**，其他场景（产品 UI / B 端）限制 3-4 种。

### 3.6 CJK 排版规则

- 中文标点全角：`。`、`，`
- 英文括号 `()` 中英混排首选
- 破折号 `——`、省略号 `……`
- **无需全角空格**（代码 space + CSS 字距）
- 数字+中文单位无空格（`100件`）
- 金额+货币符号无空格（`¥100`）

### 3.7 排版 Don't

- ❌ 中英混排全角空格
- ❌ 同屏 3+ 字重（IP 场景例外）
- ❌ 价格用 PingFangSC（必须 DIN）
- ❌ 中文用衬线字体
- ❌ 非 IP 场景用 OPPOSans40_Light / UnboundedSans（装饰字体只用在 IP / Logo / 排行榜）

---

## 四、组件样式 ★ 重写

以下为 PURESNAKE 组件样式的精确规范。

### 4.1 按钮体系 3 档 + 多子类

#### 4.1.A 全局 CTA（最高优先级）

```css
.btn-global-cta {
  background: #06D290;          /* 新版品牌绿 */
  color: #FFFFFF;
  font: 500 16px/1 "PingFang SC";  /* Sketch 32/@2x = 16 */
  height: 44px;                  /* Sketch 88/@2x = 44 */
  padding: 0 24px;
  border-radius: 2px;            /* Sketch r=4/@2x = 2 · 微圆角不是胶囊 */
  border: none;
}
.btn-global-cta:disabled {
  background: #D0D3D5;
  color: #FFFFFF;
}
```

**阿琛亲定**："全局按钮最多支持十字"（文字 ≤ 10 汉字）。

#### 4.1.B 大号主按钮（Button/big702/Primary）

- 尺寸 702×84 (@2x) = **351×42 CSS**
- fill `#26273A` **深墨蓝**（不是绿！）
- PingFangSC-Medium 14px 白字 · r=2

#### 4.1.C 大号次按钮（Button/big702/Secondary）

- 同尺寸 · white + border `#26273A`/1 + 深墨蓝字

#### 4.1.D 小号 Ghost（Button/Small190）

- 190×64 (@2x) = 95×32 CSS · white + border `#26273A`/1 + 深墨蓝字

#### 4.1.E Ghost 绿按钮（"我要寄出"等次要鼓励）

- 168×60 (@2x) = 84×30 CSS · white + border `#2EBD7C`/2 + 旧版品牌绿字 · r=1

#### 4.1.F 终态标识（"已售出"等）

- 600×100 (@2x) = 300×50 CSS · fill `#ACACB7` + 白字

#### 按钮决策树（给 Agent）

```
是行动召唤吗？
├── 是 → 优先级？
│   ├── 最高（下单/支付/出价）→ 全局 CTA #06D290 亮绿
│   ├── 高（确认/提交）→ 大号主按钮 #26273A 深墨蓝
│   ├── 中（次要行动）→ 大号次按钮 白底+深墨蓝边
│   │                  OR Ghost 绿按钮 白底+旧绿边（鼓励性）
│   └── 低 → 小号 Ghost 白底+深墨蓝边
└── 否（已终态标识）→ 禁用按钮 #ACACB7 灰底（"已售出"等）
```

### 4.2 Checkbox

- 40×40 · 选中 fill **`#26273A`**（**不是品牌绿**）+ 对号 #FFFFFF
- 未选 fill `#CCCCCC` / border `#E8E8E8`

### 4.3 Switch

- 102×62 (@2x) = 51×31 CSS
- 开启 track `#26273A` + knob 白圆 + shadow `#0000001A` blur=2
- 关闭 track `#EEEEEE` + border `#E5E5EA`

### 4.4 Toast

- 256×84 (@2x) = 128×42 CSS · r=2
- fill `#000000B2` (70% 透明黑) + 白字 PingFang Regular 16

### 4.5 标签体系（7 种）

| 类型 | 样式 |
|---|---|
| 省钱标签 | 白底+墨蓝边+分栏 "¥8000 \| 省" |
| VIP 会员 | fill `#28293C` + 白字 |
| **会员价胶囊** | fill `#F4E8DE` + 文字 `#83684F` + **r=17 半圆** |
| 倒计时徽章 | Blue1/Pupple1 底 + 白 DIN 数字 |
| 优惠券 | Red2 底 + Red1 字 |
| 鉴别结果 | 5 档语义色（见 2.6）|
| VIP 奖章 | 带 shadow，形状特殊 |

### 4.6 Tab Bar Icon

- 54×54 (@2x) = 27×27 CSS 统一规格
- 选中色 `#0F1113` + 文字色变 `#06D290` 新版品牌绿
- 未选中 `#A8AAB0`
- 第二回合 APP 5 Tab：首页 / 出售 / 鉴别 / 我的 / 服务

### 4.7 订单明细信息行（6 变体）

标准模式 702×64 (@2x) = 351×32 CSS · 横向排列：
- 左：标题 PingFang Regular 13
- 右：值 DIN-Medium 13（数字）or PingFang Regular 13（中文）

变体：标题+文字 / 标题+价格 / 价格变大 / 标题加粗+价格变大 / 带复制按钮 / 标题+负价格+标签

### 4.8 输入框

```css
.input {
  background: #F6F6F6;            /* 3Grey5 */
  border: none;
  border-radius: 4px;
  padding: 12px 16px;
  color: #0F1113;
  font-size: 14px;
  height: 44px;                   /* iOS 最小触控 */
}
.input:focus {
  background: #FFFFFF;
  border: 1px solid #06D290;      /* 新版品牌绿聚焦边 */
}
.input.error {
  background: #FFE3E3;
  border: 1px solid #FE0832;
  color: #FE0832;
}
```

### 4.9 寄售订单流程节点（Steps）

- 34×34 oval · fill `#06D290` + 对号 white
- 未完成节点样式待阿琛补

### 4.10 头像 / 选择器

- 默认头像 54×54 · line-only 设计（不填色）
- 选择器未选 34×34 oval · fill `#DDDDDD` / border `#E8E8E8`

### 4.11 待补组件（缩短清单）

基于深度挖后重新判断：
- [ ] Modal / Drawer（v3.0 没单独 page）
- [ ] Loading 全屏动画
- [ ] 空状态（+ 荣德兔 IP 插画）
- [ ] 下拉刷新
- [ ] 页内 Tab 切换
- [ ] Radio
- [ ] 评分（鉴别打分）

---

## 五、布局原则 ★ 重写

### 5.1 间距单元（4 的倍数）

基于 v3.0 + 组件库实测：

| Sketch @2x | CSS | Token | 场景 |
|---|---|---|---|
| 4 | 2 | space-1 | 极小 |
| 8 | 4 | space-2 | 小 |
| 16 | 8 | space-3 | 基础 gap ★ |
| 24 | 12 | space-4 | 卡片 padding |
| 32 | 16 | space-5 | 页面两侧边距 ★ |
| 48 | 24 | space-6 | 区块间距 |
| 64 | 32 | space-7 | 大区块 |
| 96 | 48 | space-8 | 超大 |

### 5.2 圆角系统（★ 实测修正）

| Sketch @2x | CSS | 使用频率 | 场景 |
|---|---|---|---|
| **4** | **2** | **最高频**（15+ 次）| 按钮 / 卡片 / Toast ★★★★★ |
| **2** | **1** | 次高频（14+ 次）| 小组件 / icon 内矩形 |
| 3 | 1.5 | 高频 | 过渡值 |
| 0 | 0 | 中频 | 真矩形 / 分割线 |
| 17 | 8 | 偶用 | 会员价半圆胶囊 |
| 100 / 200 | 50 / 100 | 2 次 | 全圆头像 |

**结论**：PURESNAKE 按钮/卡片**不是胶囊**，是 **r=2 CSS 微圆角**。

### 5.3 画板尺寸规范

见章 3.2。Agent 在生成时要识别场景。

### 5.4 栅格

- Mobile：2×N / 3×N 商品网格 · 4×N icon / IP 网格 · 单列信息流
- B 端 Web（1280 宽）：12 栅格

### 5.5 布局 Don't

- ❌ 非 4 倍数间距
- ❌ 页面边距 < 24 或 > 40
- ❌ 圆角同屏混用（4 / 8 / 17 并用）

---

## 六、层次与阴影

### 6.1 阴影层级

| Token | CSS | 场景 |
|---|---|---|
| shadow-xs | `0 1px 2px rgba(7,7,7,0.05)` | 微凸起 |
| shadow-sm | `0 2px 4px rgba(7,7,7,0.08)` | 卡片默认 |
| shadow-md | `0 4px 12px rgba(7,7,7,0.12)` | 悬浮 / Hover |
| shadow-lg | `0 8px 24px rgba(7,7,7,0.20)` | 弹窗 / Drawer |
| shadow-xl | `0 16px 48px rgba(7,7,7,0.30)` | 全屏模态 |

Switch knob 特殊：`0 1px 2px rgba(0,0,0,0.1)` + `border: 1px rgba(0,0,0,0.04)`

### 6.2 Z-index

base 0 · dropdown 100 · sticky 200 · overlay 900 · modal 1000 · toast 2000 · tooltip 3000

### 6.3 遮罩

- 深色弹窗背后 `rgba(0,0,0,0.5)`
- 浅色加载 `#F6F6F6CC`
- Toast 背景 `#000000B2`（70% 黑）

---

## 七、Do / Don't（护栏）★ 最重要

### 7.1 通用反模式

**色彩 / 背景**：激进渐变 · emoji 滥用 · 自创 HEX
**组件**：圆角+左边框强调色 · SVG 假图 · pure black 文字
**字体**：滥用 Inter/Roboto/Arial · 同屏 3+ 字重（IP 除外）· Body 用 Bold/Light
**填充**：placeholder 文案（Lorem ipsum）· 数据 slop
**交互**：`scrollIntoView` · 移动端过度 hover · 自动播放无静音

### 7.2 PURESNAKE 专属反模式

**品牌**：金色镀光 · 聚划算浮夸 · 苹果性冷淡 · 品牌绿做春节海报底 · 圣诞绿+红并排
**色彩**：pure black · 同屏 4+ 品牌色 · **新旧品牌绿同屏使用** · Gold × Pupple
**字体**：价格用 PingFangSC · 中文衬线 · **非 IP 场景用 OPPOSans40_Light / UnboundedSans**
**按钮**：主按钮不用 BrandGreen · 文字 > 10 字 · 禁用态文字用灰色（应白字 + 灰底）
**组件**：卡片无圆角 · 输入框默认白底 · Tab 无选中动画
**IP**：荣德兔改色 · 去掉墨镜 · 和镀金组合 · 严肃场景 · 2D 降级
**海报**：KV 商品用 3D 渲染 · 节日用平面 flat · 教程彩底 · B 端招商无流程图
**鉴别**：徽章自创色 · 证书视觉与普通卡片混同 · "100% 正品"夸大文案

### 7.3 画板规范 Don't（★ 新增）

- ❌ 混用多种画板尺寸系统（同一设计必须统一 @2x 或 @3x）
- ❌ 用 Web 1x 375 画板做 APP 设计（差 2x 精度）
- ❌ 海报/朋友圈用 iOS 750 画板（海报要独立尺寸）

### 7.4 待阿琛补（L3）

- [ ] 过去翻车的 3-5 个典型样本
- [ ] 用户/客户吐槽清单
- [ ] 老板明确反对方向
- [ ] 法律合规视觉禁忌

---

## 八、响应式行为

> 大部分仍是 L3 占位，待阿琛补

| 断点 | 宽度 |
|---|---|
| Mobile 小 | < 375 |
| Mobile 主流 | 375 - 430 |
| Tablet | 430 - 1024 |
| Desktop | 1024 - 1440 |
| Wide | > 1440 |

- 导航：Mobile 底 Tab / Tablet 侧 drawer / Desktop 顶栏
- 网格：Mobile 2×N / Tablet 3×N / Desktop 4-5×N
- 触控：最小 44×44px / 按钮间距 ≥ 8px

**待阿琛拍**：PC 端规范 / APP 最小宽度 / 官网断点 / 海外多语言

---

## 九、Agent 提示词指南

### 9.1 语境开头

```
上下文：PURESNAKE 是潮流循环交易的专业鉴别平台。
视觉语言：精准干练 + 科技潮流 + IP 辅助。
主色：BrandGreen #06D290（新版 CTA）+ BrandBlue #26273A（深）+ Snake-red #FE0832（价格/警示）+ SnakeRed #FF3367（品牌红 KV）+ SnakeDark #111111（IP/海报底）+ Pupple1 #7A3BFF（潮玩）。
字体：中文 PingFang SC，数字 DIN / DINPro，IP 可加 OPPOSans + UnboundedSans + OPPOSans40_Light。
画板：iOS APP @2x 750 宽 → CSS 除 2；Web 1x 375 不除；海报原尺寸。
圆角：产品 UI 默认 r=2 CSS 微圆角（不是胶囊）。
遵循本文 token，不自创颜色 / 字体 / 圆角 token。
```

### 9.2 色彩速查（最高频 12 色）

```
#06D290  BrandGreen-new   · 主 CTA / 成功 / 鉴别通过
#2EBD7C  BrandGreen-old   · Ghost 边 / 兼容
#26273A  BrandBlue        · 深文字 / 深按钮
#111111  SnakeDark        · IP / 海报深底
#FE0832  Snake-red        · 价格 / 警示
#FF3367  SnakeRed         · 品牌情绪红（KV）
#7A3BFF  Pupple1          · 潮玩活动
#FF6A0C  Warning          · 警告橙
#F6F6F6  3Grey5           · 浅底
#E8E8E8  3Grey4           · 主分割
#707184  3Grey1           · 辅助文字
#0F1113  Text-1           · 主文字
```

### 9.3 字体速查

```css
/* 中文 */ font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
/* 数字 */ font-family: "DIN", "DIN Alternate", sans-serif;
/* IP   */ font-family: "OPPOSans", "PingFang SC", sans-serif;
```

### 9.4 场景提示词模板

#### 主 CTA 按钮
```
主按钮：fill #06D290（新版品牌绿）· 白字 PingFang Medium 16 · height 44 · r=2（微圆角不胶囊）·
三态（正常/加载同色+spinner/禁用 #D0D3D5+白字）· 文字 ≤10 汉字。
```

#### 鉴别结果徽章
```
顶部 BrandBlue #26273A 深 banner · 状态徽章按语义色（通过 #06D290 / 不通过 #FE0832 / 待鉴 #FF6A0C / 专家 #7A3BFF / 机器 #2D57E7）·
下方白底详情卡 · 鉴别师 + 时间戳 · 证书感。
```

#### 活动 KV 海报
```
PURESNAKE 活动 KV：画布 750×1624（竖图 banner）或 1080×1956（社交标准）·
亮品牌绿 #06D290 纯色底 · 主标题 2-8 字粗黑 PingFangSC-Bold 白字 · 实物商品大图白底去背 · 第2回合 Logo 小字角落 · 爆炸装饰四角。
不要 3D 渲染商品。
```

#### 节日海报（模板 A）
```
节日海报：画布 750×1334 or 社交 1080×1080 ·
红金色系（禁用品牌绿）· 立体气球字/3D 雕塑 IP · 中英双语（主英文花体 + 中文副标）· 对称排版 · 四角装饰。
使用 OPPOSans / DIN-BoldItalicAlt 装饰字。
```

#### B 端招商（模板 D）
```
B 端招商 landing：画布 1280×2774 or PPT 1280×720 ·
亮品牌绿底 · 价值主张大字 · 3D 潮流人物（不是荣德兔）· 5 步流程图 · 二维码 + SNAKE 星空潮奢咨询师。
```

### 9.5 反模式清单（防跑偏）

```
不用：渐变 / emoji / 圆角+左边框 / SVG 假图 / Inter-Roboto-Arial / 聚划算浮夸 / 奢侈镀金 / 性冷淡无色
色彩：pure black 改 #0F1113 · 新旧绿别同屏 · 同屏 ≤3 品牌色
字体：价格必须 DIN · 非 IP 场景不用 OPPOSans40_Light / UnboundedSans · 不用中文衬线
按钮：主 CTA = #06D290 新版绿 · 文字 ≤10 字 · r=2 CSS 微圆角（不胶囊）
禁用：白字 + #D0D3D5 灰底，不是灰字
```

### 9.6 追问触发器

- [ ] 没明确设计规范来源 → 先问
- [ ] 画板尺寸不明（iOS @2x/@3x / Web / 海报）→ 先问
- [ ] 场景不清（产品/KV/B 端）→ 先问是哪个模板
- [ ] IP 用荣德兔 / 生肖 / 不用？
- [ ] 要变体还是一版？

---

### 10.4 IP 场景字体矩阵（★ 新增）

IP 页面可并用多字体营造"炫感"：

```
OPPOSans-R / 20-22           · 主英文标题
OPPOSans-B / 20-32           · 装饰英文加粗
OPPOSans40_Light / 46        · 超大号装饰字（仅 IP 用）
UnboundedSans-Regular / 22-26 · 装饰英文状态
DIN-Bold / 22-31              · 大号数字
DINPro-Bold / 25              · 装饰数字
PingFangSC-Medium / 36        · 中文副标
```

IP 场景唯一允许**突破 "同屏 ≤3 字体"** 规则的场景。

### 10.5 IP 扩展色板

| HEX | 用途 |
|---|---|
| `#00F900` | IP 高饱和荧光绿 |
| `#FCCD79` | IP 暖黄 |
| `#231916` | 近黑棕（IP 底）|
| `#012F10` | 深绿（IP 底）|
| `#EEFFE0` | 浅绿（IP 辅助）|


---

## 十一、运营活动风格模板

### 11.1 5 大模板

| 模板 | 场景 | 主色 | 关键元素 |
|---|---|---|---|
| A · 节日主题 | 春节/元宵/元旦/生肖 | 红金 | 立体字 + 雕塑 IP + 双语 + 对称 |
| B · 功能教程 | 规则说明/流程 | 白/浅灰 | isometric 插画 + 书籍层级 |
| C · 活动 KV | 拍卖/促销/品牌活动 | 亮品牌绿 | 实物大图 + 爆炸装饰 + 简短大标题 |
| D · 商家招商 | B 端拉新 | 亮品牌绿 | 3D 人物 + 流程图 + 信息图表 |
| E · 合作特化 | 第三方生态（支付宝等）| 对方品牌色 | 双 Logo + 嫁接元素 |

### 11.2-11.6 模板细节

模板细节待补。

### 11.7 文件管理规范

- **命名**：`YYMMDD-主题`
- **副本机制**：大改版前 `_副本` 保留
- 避免重名累积（当前有 3 对重复要清理）

### 11.8 海报标准尺寸表（★ 新增）

实测阿琛历史海报尺寸规范：

| 场景 | 尺寸 |
|---|---|
| 朋友圈 / 小红书 竖图 | **750×1624** |
| 朋友圈 1:1 方图 | **1080×1080** |
| 朋友圈长图 | **1503×4264** |
| APP Banner | **500×400** |
| 社交 IP 海报 | **1080×1956 / 1104×1858** |
| 全场景展开 | **14681×6793**（超大）|
| iOS APP 全屏 | **750×1334 / 750×2061** |
| Web Landing | **1280×2774** |
| PPT 16:9 | **1280×720** |
| A4 印刷物料 | **2480×3508**（A4 300dpi）|

### 11.9 运营海报独立色集

海报特有色值（不用在产品 UI）：

- `#FDE834` 荧光黄（活动）
- `#E4F0FF` 淡蓝（教程）
- `#00FF67` / `#20FF7E` 高饱和亮绿（招商 KV）
- `#FFE615` 荧光黄（招商）


---

## 附录 A · 工程纪律

Agent 生成 PURESNAKE UI 代码必须遵守：

- 单文件 ≤ 1000 行
- 大改 v2 复制保留
- 描述性文件名
- React styles 对象带组件名前缀（`productCardStyles` 不是 `styles`）
- 只拷贝需要资产，不一锅端
- 禁止 `scrollIntoView`
- 使用本文 token，不自创

