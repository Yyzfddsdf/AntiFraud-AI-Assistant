# 独立桌面端前端

这是桌面端的独立 Vue 模块化前端工程。

当前桌面端已经收敛为纯管理员控制台，只保留管理员分析、地理态势、用户管理、案件审核、案件库和独立管理助手入口。
账号体系仍然统一保留注册/登录；非管理员登录桌面端后，可通过邀请码升级为管理员再进入控制台。
桌面端管理助手已切换到管理员专用接口：`/api/admin/chat`，与普通用户聊天上下文隔离。

## 启动

```bash
npm install
npm run dev
```

默认开发地址：`http://localhost:5173`

## 联调要求

- 后端默认运行在 `http://127.0.0.1:8081`
- Vite 已代理 `/api`

## 当前模块划分

- `src/app/useDesktopApp.js`：桌面端总装配入口
- `src/modules/case-library/`：案件库与审核模块
- `src/modules/charts/`：趋势图与图谱模块
- `src/modules/chat/`：聊天模块
- `src/modules/router/`：标签路由模块
- `src/modules/session/`：登录会话模块
- `src/modules/tabs/`：标签副作用模块
