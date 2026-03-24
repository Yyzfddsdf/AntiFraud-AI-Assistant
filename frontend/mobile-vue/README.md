# 独立移动端前端

这是移动端的独立 Vue 模块化前端工程。

## 启动

```bash
npm install
npm run dev
```

默认开发地址：`http://localhost:5174`

## 联调要求

- 后端默认运行在 `http://127.0.0.1:8081`
- Vite 已代理 `/api`

## 当前模块划分

- `src/app/useMobileApp.js`：移动端总装配入口
- `src/modules/tasks/`：分析任务与历史模块
- `src/modules/alerts/`：风险预警模块
- `src/modules/family/`：家庭守护模块
- `src/modules/charts/`：移动端趋势图模块
- `src/modules/chat/`：聊天模块
- `src/modules/router/`：标签路由模块
- `src/modules/session/`：登录会话模块
- `src/modules/tabs/`：标签副作用模块
