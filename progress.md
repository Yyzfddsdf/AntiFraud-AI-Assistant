# Progress

## 2026-03-24
- 启动移动端 Vue 模块化重构
- 读取了移动端旧入口 `web/login-mobile/index.html` 与 `assets/js/app.js`
- 对照了桌面端 `frontend/desktop-vue` 的当前结构
- 确认本轮将新建 `frontend/mobile-vue`，不直接改旧移动端入口
- 已将移动端逻辑拆为 `tasks / alerts / family / charts / chat / router / session / tabs`
- 已将移动端源码并入 `frontend/desktop-vue/src/mobile/`
- 已将移动端页面入口并入 `frontend/desktop-vue/mobile/index.html`
- 已将移动端静态资源并入 `frontend/desktop-vue/public/mobile-assets/`
- 已将统一 Vite 服务保留在 `5173`，构建产物同时生成 `dist/index.html` 和 `dist/mobile/index.html`
- 已重建 `frontend/` 为真正的根工程，新增 `desktop/`、`mobile/`、`src/desktop/`、`src/mobile/`
- 已从 `frontend` 根目录执行 `npm install` 和 `npm run build`，构建产物包含 `dist/index.html`、`dist/desktop/index.html`、`dist/mobile/index.html`
- 已删除临时 `frontend/mobile-vue/` 目录
- 旧 `frontend/desktop-vue/` 因外部进程占用暂未删除，但已不再作为启动入口
- 已恢复 `frontend/desktop-vue/` 独立桌面端工程，并完成 `npm install`、`npm run build`
- 已恢复 `frontend/mobile-vue/` 独立移动端工程，并完成 `npm install`、`npm run build`
- 已清理统一根工程残留，`frontend/` 下当前只保留 `desktop-vue/`、`mobile-vue/` 和说明文档
