<template>
  <div class="h-full flex flex-col transition-colors duration-300">
    <AppOverlays :app="app" />
    <DesktopAuthView v-if="authReady && !isAuthenticated" :app="app" />
    <section v-else-if="authReady && !isAdminUser" class="h-full flex items-center justify-center bg-slate-950 px-6">
      <div class="w-full max-w-2xl rounded-sm border border-white/10 bg-white/5 p-8 text-white shadow-2xl">
        <div class="text-[11px] font-black uppercase tracking-[0.24em] text-cyan-200/80">Admin Activation</div>
        <h1 class="mt-3 text-3xl font-extrabold tracking-tight">管理员开通</h1>
        <p class="mt-3 text-sm leading-7 text-slate-300">
          当前账号 <span class="font-bold text-white">{{ getUserDisplayName(user) }}</span> 已完成登录，但角色还是
          <span class="font-bold text-white">{{ user.role || 'user' }}</span>。桌面端业务区只开放给管理员，请先输入邀请码完成升级。
        </p>

        <div class="mt-6 grid gap-4 rounded-sm border border-white/10 bg-slate-900/40 p-5">
          <div>
            <label class="block text-xs font-bold uppercase tracking-[0.22em] text-cyan-200/80 mb-2">管理员邀请码</label>
            <input
              v-model="inviteCode"
              type="text"
              class="w-full px-4 py-3 rounded-sm border border-cyan-300/20 bg-white/95 text-slate-900 outline-none focus:ring-2 focus:ring-cyan-300"
              placeholder="输入邀请码后即可开通管理员桌面端"
            >
          </div>
          <div class="flex flex-wrap gap-3">
            <button
              @click="upgradeAccount"
              :disabled="!inviteCode || loading"
              class="px-5 py-2.5 rounded-sm bg-cyan-400 text-slate-950 text-sm font-bold hover:bg-cyan-300 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ loading ? '开通中...' : '立即开通管理员权限' }}
            </button>
            <button @click="logout" class="px-5 py-2.5 rounded-sm border border-white/15 bg-white/5 text-white text-sm font-bold hover:bg-white/10 transition-colors">
              退出当前账号
            </button>
          </div>
        </div>

        <div class="mt-5 text-xs leading-6 text-slate-400">
          说明：账号体系仍然统一，普通用户注册后可通过邀请码升级为管理员；升级成功后会直接进入管理员控制台。
        </div>
      </div>
    </section>
    <GeoRiskMapPage v-else-if="activeTab === 'geo_risk_map_full'" :app="app" />
    <DashboardWorkspace v-else :app="app" />
  </div>
</template>

<script>
import { useDesktopApp } from './app/useDesktopApp';
import AppOverlays from './components/shell/AppOverlays.vue';
import DesktopAuthView from './components/auth/DesktopAuthView.vue';
import GeoRiskMapPage from './components/geo/GeoRiskMapPage.vue';
import DashboardWorkspace from './components/dashboard/DashboardWorkspace.vue';

export default {
  name: 'DesktopApp',
  components: {
    AppOverlays,
    DesktopAuthView,
    GeoRiskMapPage,
    DashboardWorkspace
  },
  setup() {
    const app = useDesktopApp();
    return {
      app,
      ...app
    };
  }
};
</script>
