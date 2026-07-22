---
name: MoveKit
description: Windows 用户文件夹迁移工具 — 安全、一键、Fluent Design
colors:
  accent: "#0078D4"
  accent-dark: "#006CBE"
  accent-light: "#429CE3"
  surface: "#F3F3F3"
  surface-alt: "#E5E5E5"
  card: "#FFFFFF"
  text-primary: "#1A1A1A"
  text-secondary: "#616161"
  text-disabled: "#ABABAB"
  border: "#E0E0E0"
  border-hover: "#C4C4C4"
  success: "#0F7B0F"
  warning: "#9D5D00"
  danger: "#C42B1C"
typography:
  body:
    fontFamily: "Segoe UI Variable, Segoe UI, system-ui, -apple-system, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
  title:
    fontFamily: "Segoe UI Variable, Segoe UI, system-ui, -apple-system, sans-serif"
    fontSize: "20px"
    fontWeight: 600
    lineHeight: 1.4
  label:
    fontFamily: "Segoe UI Variable, Segoe UI, system-ui, -apple-system, sans-serif"
    fontSize: "12px"
    fontWeight: 500
    lineHeight: 1.4
rounded:
  fluent: "8px"
  full: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "#FFFFFF"
    rounded: "{rounded.fluent}"
    padding: "10px 20px"
  button-primary-hover:
    backgroundColor: "{colors.accent-dark}"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.text-secondary}"
    rounded: "{rounded.fluent}"
    padding: "8px 16px"
  button-secondary-hover:
    backgroundColor: "{colors.surface-alt}"
  card:
    backgroundColor: "{colors.card}"
    rounded: "{rounded.fluent}"
    padding: "16px"
  badge:
    backgroundColor: "{colors.accent}10"
    textColor: "{colors.accent}"
    rounded: "{rounded.full}"
    padding: "2px 8px"
---

# Design System: MoveKit

## 1. Overview

**Creative North Star: "Windows 控制中心"**

MoveKit 的设计语言继承 Windows 11 Settings 的秩序感——每个元素都在它该在的位置，没有惊喜，只有确定性。用户打开应用时应该感到"这看起来就是 Windows 的一部分"，而不是一个第三方工具在模仿系统。

这是一个 product register 的设计系统。设计服务于任务，不是展示自己。用户带着焦虑来（C 盘快满了），带着安心走（迁移完成，随时可回滚）。界面不应该成为他们焦虑的来源——没有花哨的动画，没有需要学习的新交互，没有让人犹豫的模糊状态。

**Key Characteristics:**
- **系统感**：Segoe UI 字体、8px 圆角、蓝色主色调，与 Windows 11 视觉语言一致
- **秩序感**：信息层次清晰，主次分明，不填满每一寸空间
- **安全感**：每个操作都有明确的反馈，危险操作有警告，所有操作可回滚
- **克制感**：装饰性元素为零，每个像素都有功能目的

## 2. Colors: Windows 原生色板

系统蓝色（#0078D4）是 Windows 生态的通用信任色。用它建立与用户操作系统之间的视觉关联——看到这个蓝色，就知道"这是系统级的操作"。

### Primary
- **Windows 信任蓝** (#0078D4): 主按钮、选中状态、进度条、推荐标签。只用于"用户应该关注的东西"，不用于装饰。

### Neutral
- **页面底布** (#F3F3F3): 应用背景，与 Windows 11 Settings 背景一致。
- **卡片白** (#FFFFFF): 内容容器背景，通过微妙阴影与底布分层。
- **正文黑** (#1A1A1A): 标题和主要文字，不用纯黑（#000）以减少视觉压力。
- **辅助灰** (#616161): 次要信息、路径、描述文字。
- **禁用灰** (#ABABAB): 不可用状态的文字和图标。
- **分割线** (#E0E0E0): 边框和分割线，hover 时加深到 #C4C4C4。

### Semantic
- **成功绿** (#0F7B0F): 迁移完成、已迁移状态标记。
- **警告橙** (#9D5D00): 需要注意但不危险的提示（OneDrive 冲突、磁盘空间不足）。
- **危险红** (#C42B1C): 错误状态、失败标记、非 NTFS 磁盘警告。

### Named Rules

**The 10% Rule.** 蓝色主色调在任何屏幕上的使用面积不超过 10%。它是引导色，不是填充色。如果一个页面看起来"很蓝"，说明用多了。

**The No-Pure-Black Rule.** 永远不用 #000000 作为文字颜色。#1A1A1A 是最深的值。纯黑在白色背景上会产生振动效应，增加阅读疲劳。

**The Semantic Clarity Rule.** 红、橙、绿只用于语义（错误、警告、成功），不用于装饰。如果一个元素需要颜色来吸引注意力，它一定在传达某种状态。

## 3. Typography

**Display Font:** Segoe UI Variable（与 Windows 11 系统字体一致）
**Body Font:** Segoe UI Variable
**Fallback:** Segoe UI → system-ui → -apple-system → sans-serif

**Character:** 单一字体族，不使用字体配对。Segoe UI Variable 是 Windows 11 的系统字体，使用它让应用看起来像系统的一部分。字体个性是"安静的"——它不应该被注意到，只应该被阅读。

### Hierarchy
- **页面标题** (600, 20px, 1.4): 应用名称和页面级标题（"MoveKit"、"正在迁移"、"迁移完成"）。
- **章节标题** (600, 18px, 1.4): 区域级标题（"选择目标磁盘"、"系统文件夹"）。
- **正文** (400, 14px, 1.5): 主要内容文字。最大行宽 65–75ch。
- **辅助文字** (400, 14px, 1.5): 路径、描述、次要信息。颜色为辅助灰。
- **标签** (500, 12px, 1.4): 状态标签（"推荐"、"已迁移"、"待迁移"），常与圆角胶囊组合。

### Named Rules

**The One-Family Rule.** 只使用一个字体族。Product UI 不需要 display/body 配对。Segoe UI Variable 的多种字重（400/500/600）足以建立层次。

**The Fixed-Scale Rule.** 使用固定的 rem 级别（12px → 14px → 18px → 20px），不使用 clamp() 流体排版。Product UI 在固定 DPI 下运行，流体标题只会制造不稳定感。

## 4. Elevation

使用轻量级阴影建立层次，不使用背景色变化或边框叠加。阴影的角色是"分离"，不是"装饰"。

### Shadow Vocabulary
- **卡片阴影** (`0 2px 4px rgba(0, 0, 0, 0.08)`): 默认卡片状态，与底布微妙分离。
- **悬浮阴影** (`0 8px 16px rgba(0, 0, 0, 0.1)`): 确认对话框等需要强调的浮层。
- **选中光环** (`ring-2 ring-accent/20`): 选中状态的磁盘选项，用半透明蓝色环表示。

### Named Rules

**The Flat-By-Default Rule.** 元素在默认状态下是平的。阴影只在需要表示"这个元素在上面"时出现（hover、选中、浮层）。如果一个静态元素有阴影，它应该有理由。

**The One-Layer Rule.** 不嵌套阴影。卡片在底布上有阴影，但卡片内部的元素不再有阴影。嵌套阴影是"臃肿的国产软件"的典型特征。

## 5. Components

### Buttons

按钮是用户行动的触发器，它的视觉权重应该与行动的重要性成正比。

- **Shape:** 8px 圆角（--radius-fluent），与 Windows 11 按钮一致。
- **Primary:** 蓝色背景 (#0078D4)，白色文字，padding 10px 20px。用于"确认迁移"、"完成"等主操作。
- **Hover:** 加深到 #006CBE，transition 200ms。
- **Secondary:** 透明背景，辅助灰文字，1px 边框。用于"取消"、"回滚"、"取消迁移"等次要操作。
- **Hover:** 浅灰背景 (#E5E5E5)。

### Cards / Containers

卡片是内容的容器，不是装饰元素。

- **Corner Style:** 8px 圆角（--radius-fluent）。
- **Background:** 白色 (#FFFFFF)。
- **Shadow Strategy:** 默认无阴影，用 1px 边框 (#E0E0E0) 与底布区分。需要强调时（确认对话框）添加卡片阴影。
- **Internal Padding:** 16px（--spacing-md）。

### Inputs / Fields (DiskSelector)

磁盘选择器是自定义的 radio group，不是标准 input。

- **Style:** 卡片式选择项，8px 圆角，1px 边框。
- **Selected:** 蓝色边框 + 蓝色光环（ring-2 ring-accent/20）。
- **Disabled:** 灰色背景，60% 透明度，cursor-not-allowed。
- **Recommended badge:** 蓝色胶囊标签，半透明蓝底 (#0078D410)。

### Progress Bar

进度条是迁移过程中最重要的视觉元素，它告诉用户"事情在发生"。

- **Track:** 浅灰背景 (#E5E5E5)，圆角胶囊。
- **Fill:** 蓝色 (#0078D4)，圆角胶囊，300ms ease-out 过渡。
- **File count bar:** 使用较浅的蓝色 (#429CE3)，与字节进度条区分层次。

### Status Badges

状态标签是文件夹列表中的关键信息元素。

- **Shape:** 圆角胶囊（9999px），padding 2px 8px，12px 字号。
- **待迁移:** 灰底 (#E5E5E5) + 辅助灰文字。
- **已迁移:** 绿底 (#0F7B0F10) + 成功绿文字。
- **链接异常:** 红底 (#C42B1C10) + 危险红文字。

### Dialogs / Overlays

对话框用于需要用户确认的危险操作。

- **Overlay:** 半透明黑色遮罩。
- **Card:** 白色背景，8px 圆角，大阴影 (--shadow-fluent-lg)。
- **Warning border:** 左侧或顶部使用警告色边框。
- **Structure:** Header（图标 + 标题）→ Body（说明文字）→ Footer（操作按钮）。

## 6. Do's and Don'ts

### Do:
- **Do** 使用 Segoe UI Variable 作为唯一字体族。看到系统字体，用户就知道"这是 Windows 的一部分"。
- **Do** 保持蓝色主色调的克制使用——主按钮、选中状态、进度条、推荐标签。不多不少。
- **Do** 用 8px 圆角统一所有容器和按钮。一致性是信任的基础。
- **Do** 用图标 + 文字 + 颜色三重编码传达状态。成功用绿色 ✓ + "已迁移"，失败用红色 ✕ + 错误描述。
- **Do** 保持 max-w-2xl (672px) 的内容宽度。这是工具类应用，不是信息流，不需要撑满屏幕。
- **Do** 每个危险操作（迁移、回滚）都有确认步骤。PRODUCT.md 说"安全第一"，确认对话框是这句话的实体化。
- **Do** 用 150–250ms 的过渡时间。用户在任务流中，不要让他们等动画。

### Don't:
- **Don't** 使用渐变背景、毛玻璃效果、或任何装饰性视觉效果。PRODUCT.md 明确反对"过度设计的 SaaS"风格。
- **Don't** 嵌套卡片。卡片里再套卡片是"臃肿的国产软件"的典型特征。
- **Don't** 使用弹性和回弹动画（bounce/elastic easing）。这让人觉得应用不专业。
- **Don't** 用纯黑 (#000000) 作为文字颜色。#1A1A1A 是最深的值。
- **Don't** 在非语义场景使用红/橙/绿。颜色只传达状态，不用于装饰。
- **Don't** 使用超过两种阴影层级。底布 → 卡片 → 浮层，最多三层。
- **Don't** 让按钮看起来不像按钮。标准的填充/描边/圆角组合，不要发明新的交互范式。PRODUCT.md 说"不要简陋的命令行工具"，但也不要矫枉过正到"让人不敢点"。
- **Don't** 在加载状态使用居中大 spinner。用骨架屏或内联加载指示器。
