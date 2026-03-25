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
