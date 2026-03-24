# Findings

## Mobile Legacy Frontend
- 现有移动端在 `web/login-mobile/`
- 入口是 `index.html`
- 逻辑主要集中在 `assets/js/app.js`，体量约 104 KB
- 已存在独立的全局脚本拆分：`router.js`、`tab-config.js`、`tab-effects.js`、`session.js`
- 需要的第三方资源包括 `tailwindcss.cdn.js`、`echarts.min.js`、`marked.min.js`、字体和样式文件

## Desktop Vue Reference
- 独立桌面端工程已在 `frontend/desktop-vue/`
- 入口组织方式为 `App.vue + useDesktopApp.js`
- 已拆分 `router/session/tabs/alerts/family/case-library/charts/chat` 模块
- 组件层至少拆分为根组件、认证视图、工作台、工作台子视图

## Mobile Refactor Direction
- 新工程目录定为 `frontend/mobile-vue/`
- 公共静态资源直接复制旧移动端 assets，旧全局 JS 不再作为运行入口
- 移动端业务逻辑拆为 `tasks/alerts/family/charts/chat/router/session/tabs`
- 先按桌面端思路拆出根壳、认证页、工作台，再视情况继续细拆 tab 视图

## Current Deployment Decision
- 当前恢复为两个独立前端服务
- 桌面端工程：`frontend/desktop-vue/`
- 移动端工程：`frontend/mobile-vue/`
- 桌面端默认端口：`5173`
- 移动端默认端口：`5174`
- 两个工程都单独代理 `/api` 到 `http://127.0.0.1:8081`
