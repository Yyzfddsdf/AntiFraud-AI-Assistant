# Session Memory

日期：2026-03-25
项目：D:\Work\AntiFraud-AI-Assistant
用户称呼要求：必须使用全称 yyz先生

## 已完成事项
- 确认模拟题生成失败默认不会删除题包，只会把生成任务标记为 ailed。
- 已修改后端：每次创建新题目前先删除当前用户历史 ailed 生成任务。
  - 文件：internal/modules/scam_simulation/service.go
- 已修改模拟题提示词：不允许 10 题全部使用同一个答案字母作为正确答案。
  - 文件：internal/platform/config/config.json
  - 文件：internal/platform/config/config.go
  - 文件：internal/modules/multi_agent/core/simulation_quiz_agent.go
- 已把前端模拟题生成轮询间隔改为 15 秒。
  - 文件：rontend/desktop-vue/src/app/useDesktopApp.js
  - 文件：rontend/mobile-vue/src/app/useMobileApp.js
- 已修复前端“点选项没反应”问题，根因是开始答题后没有正确写回 simulationPackId。
- 已修复 rontend/mobile-vue/src/App.template.html 的一处模板闭合错误，并用本地 Vue 编译器验证过 OK。
- 已做过桌面端模拟答题界面优化。
- 已做过桌面端家庭界面优化。
- 已全局优化桌面端网页界面（移除杂乱色块、升级高定极简风 tech-bg、优化品牌色调至反诈深蓝调、升级 Glassmorphism 参数），满足“优雅、高级、简洁”的主题要求。

## 当前高风险/未完全收敛区域
- rontend/mobile-vue/src/App.template.html 仍然非常大，维护困难，用户明确要求后续拆分成小块。
- 移动端 simulation_quiz 页面做过多轮布局调整，顶部栏、底部导航避让、考试页样式来回改过，后续需要在拆分后再继续稳定化。
- 桌面端模拟答题区做过重构，中途曾把模板改坏，后来恢复并重做，需要继续谨慎验证。

## 用户明确偏好
- 品牌名不要乱改。
- 移动端左上角主标题：反诈卫士
- 移动端左上角副标题：Sentinel AI
- 桌面端左上角标题：Sentinel AI
- 以后优先精确定位再改，不要靠猜测不断调参数。
- 对超长模板，优先拆分，不要继续在一个大文件里缝补。

## 用户反馈过的问题重点
- 很在意我是否改错平台（桌面端/移动端）。
- 很在意我是否只改了局部却说“改回原样”。
- 对移动端模拟页遮挡、空白、滚动问题非常敏感。

## 建议的后续顺序
1. 拆分 frontend/mobile-vue/src/App.template.html
2. 拆分后再逐页修复移动端模拟页问题
3. 对桌面端模板改动做编译级验证
4. 保持品牌文案稳定

## 2026-03-25 本次新增记录
- 已完成移动端大模板拆分：将 `frontend/mobile-vue/src/App.template.html` 从超大单文件拆为装配层 + 9 个子组件，当前主模板约 306 行。
  - 新增文件：`frontend/mobile-vue/src/components/auth/MobileAuthScreen.vue`
  - 新增文件：`frontend/mobile-vue/src/components/common/MobileLoadingOverlay.vue`
  - 新增文件：`frontend/mobile-vue/src/components/common/MobileToastStack.vue`
  - 新增文件：`frontend/mobile-vue/src/components/layout/MobileAppHeader.vue`
  - 新增文件：`frontend/mobile-vue/src/components/layout/MobileBottomNav.vue`
  - 新增文件：`frontend/mobile-vue/src/components/screens/MobileDashboardScreens.vue`
  - 新增文件：`frontend/mobile-vue/src/components/screens/MobileSocialScreens.vue`
  - 新增文件：`frontend/mobile-vue/src/components/screens/MobileProfileScreens.vue`
  - 新增文件：`frontend/mobile-vue/src/components/screens/MobileWorkflowScreens.vue`
  - 修改文件：`frontend/mobile-vue/src/App.vue`
  - 修改文件：`frontend/mobile-vue/src/App.template.html`
- 本次拆分策略：
  - `UTF-8`：所有读取与验证均明确使用 UTF-8。
  - `低耦合`：`useMobileApp.js` 保持原业务逻辑不动，`App.vue` 作为装配层，通过分组 state 将页面模板与业务状态解耦。
  - `防连环报错`：优先保留原有方法、字段名和数据流，只移动模板边界，避免同时重写交互逻辑。
- 已完成最小验证：
  - 直接 `vite build` 受当前沙箱 `spawn EPERM` 限制，未能完成正式构建。
  - 已使用 `@vue/compiler-sfc` 对 `src` 下 10 个 Vue 文件做语法级校验，结果通过。
- 下一个任务建议：
  - 继续把 `family_invite`、`selectedTask`、`activeAlertEvent`、`activeFamilyNotification` 这 4 个剩余 modal 从 `App.template.html` 再拆成独立组件。

## 2026-03-25 本次继续新增记录
- 已继续完成剩余 4 个移动端 modal 拆分，`frontend/mobile-vue/src/App.template.html` 现约 19 行，仅保留页面装配。
  - 新增文件：`frontend/mobile-vue/src/components/modals/MobileFamilyManageModal.vue`
  - 新增文件：`frontend/mobile-vue/src/components/modals/MobileTaskDetailModal.vue`
  - 新增文件：`frontend/mobile-vue/src/components/modals/MobileAlertDetailModal.vue`
  - 新增文件：`frontend/mobile-vue/src/components/modals/MobileFamilyAlertModal.vue`
  - 修改文件：`frontend/mobile-vue/src/App.vue`
  - 修改文件：`frontend/mobile-vue/src/App.template.html`
- 本次拆分策略继续保持：
  - `UTF-8`：所有读取、校验与记录均显式采用 UTF-8。
  - `低耦合`：modal 逻辑通过分组 state 注入，不把业务细节重新塞回装配模板。
  - `防连环报错`：不改原始业务方法，仅增加关闭函数和可见性包装，避免影响既有行为。
- 已完成最小验证：
  - 已再次使用 `@vue/compiler-sfc` 对 `src` 下 14 个 Vue 文件做语法级校验，结果通过。
- 下一个任务建议：
  - 开始逐页修复移动端模拟页在真实交互下的遮挡、空白和滚动稳定性问题。

## 2026-03-25 本次继续新增记录 2
- 已对移动端演练页做一轮滚动稳定性修正：
  - 修改文件：`frontend/mobile-vue/src/app/useMobileApp.js`
  - 修改文件：`frontend/mobile-vue/src/components/screens/MobileWorkflowScreens.vue`
- 本次修正内容：
  - 不再只滚动 `main`，额外覆盖 `submit`、`simulation overview`、`simulation exam` 三类滚动容器。
  - 在进入考试态、退出回总览时，主动把对应容器滚动到顶部，减少 fixed overlay 场景下的残留滚动位置问题。
  - 为相关滚动容器补充显式标识与 `overscroll-behavior: contain`，降低页面穿透滚动和空白回弹风险。
- 已完成最小验证：
  - 已再次使用 `@vue/compiler-sfc` 对 `src` 下 14 个 Vue 文件做语法级校验，结果通过。
- 下一个任务建议：
  - 在真实移动端交互路径下继续验证演练总览页和考试页的顶部遮挡、底部导航避让、返回后滚动位置三类问题。
