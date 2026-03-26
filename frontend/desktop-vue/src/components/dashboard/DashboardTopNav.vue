<template>
  <nav class="bg-white border-b-2 border-slate-300 h-12 flex items-center px-4 justify-between shrink-0">
    <div class="flex items-center gap-6">
      <div class="flex items-center gap-2">
        <div class="w-6 h-6 bg-brand-600 flex items-center justify-center text-white">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
        </div>
        <span class="font-black text-sm text-slate-900 tracking-tight uppercase">Sentinel AI</span>
      </div>

      <div class="flex items-center gap-1">
        <button @click="openTaskCenter" :class="['px-3 py-1 text-xs font-bold transition-colors', isTaskCenterTab(activeTab) ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">任务中心</button>
        <button @click="activeTab = 'simulation_quiz'" :class="['px-3 py-1 text-xs font-bold transition-colors', activeTab === 'simulation_quiz' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">反诈模拟</button>
        <button @click="activeTab = 'family'" :class="['px-3 py-1 text-xs font-bold transition-colors relative', activeTab === 'family' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">
          家庭中心
          <span v-if="familyUnreadCount > 0" class="absolute -top-1 -right-1 min-w-[14px] h-[14px] px-0.5 bg-rose-500 text-white text-[8px] font-black leading-[14px] text-center">{{ familyUnreadCount }}</span>
        </button>

        <div v-if="user.role === 'admin'" class="flex items-center gap-1 ml-2 pl-2 border-l border-slate-300">
          <button @click="activeTab = 'admin_stats'" :class="['px-3 py-1 text-xs font-bold transition-colors', activeTab === 'admin_stats' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">全景分析</button>
          <button @click="activeTab = 'geo_risk_map_full'" :class="['px-3 py-1 text-xs font-bold transition-colors', activeTab === 'geo_risk_map_full' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">地理态势</button>
          <button @click="activeTab = 'users'" :class="['px-3 py-1 text-xs font-bold transition-colors', activeTab === 'users' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">用户管理</button>
          <button @click="activeTab = 'case_review'" :class="['px-3 py-1 text-xs font-bold transition-colors', activeTab === 'case_review' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">案件审核</button>
          <button @click="activeTab = 'case_library'" :class="['px-3 py-1 text-xs font-bold transition-colors', activeTab === 'case_library' ? 'text-brand-700' : 'text-slate-700 hover:text-brand-700 hover:bg-slate-100']">案件库</button>
        </div>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <button @click="activeTab = 'profile'" class="flex items-center gap-2 px-2 py-1 hover:bg-slate-100 transition-colors">
        <div class="w-6 h-6 bg-brand-50 border border-brand-200 flex items-center justify-center text-[10px] font-bold text-brand-700">
          {{ getUserAvatarText(user) }}
        </div>
        <span class="text-xs font-bold text-slate-700">{{ getUserDisplayName(user) }}</span>
      </button>
      <button @click="logout" class="text-slate-500 hover:text-brand-600 p-1 transition-colors" title="退出登录">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path></svg>
      </button>
    </div>
  </nav>
</template>

<script>
export default {
  name: 'DashboardTopNav',
  props: {
    app: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    return props.app;
  }
};
</script>
