# Mobile Vue Refactor Plan

## Goal
保持桌面端和移动端为两个独立可启动的 Vue 工程：`frontend/desktop-vue` 与 `frontend/mobile-vue`，分别运行、分别构建，同时保持现有 UI 和交互一致。

## Phases
- [completed] 梳理移动端现状与桌面端可复用结构
- [completed] 恢复 `frontend/desktop-vue` 独立桌面端工程
- [completed] 恢复 `frontend/mobile-vue` 独立移动端工程
- [completed] 清理统一根工程残留，恢复双服务目录结构
- [completed] 更新文档与忽略规则，完成双工程构建验证

## Constraints
- 保持移动端当前 UI、文案和主要交互一致
- 新前端必须独立启动，不依赖旧静态页面入口
- 新功能保持模块解耦，避免大面积联动改动
- 尽量不修改现有 `web/login-mobile` 目录

## Errors Encountered
- PowerShell 复制目录时一度产生嵌套目录，已修正结构后继续
- 中途用户要求“桌面端和移动端统一同一服务”，已短暂切到单服务方案，随后又按最新要求恢复为双服务方案
