<template>
  <nav class="fixed bottom-0 left-0 right-0 w-full bg-white/90 backdrop-blur-xl border-t border-slate-100 flex justify-around items-center h-[72px] pb-safe" style="z-index: 999;">
    <button @click="state.activeTab = 'tasks'" class="flex flex-col items-center justify-center w-16 h-full gap-1 transition-all duration-200 active:scale-95 group">
      <div :class="['w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300', ['tasks','history','risk_trend','chat'].includes(state.activeTab) ? 'bg-emerald-50 text-emerald-600' : 'text-slate-400 group-hover:bg-slate-50']">
        <svg class="w-[22px] h-[22px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold transition-colors', ['tasks','history','risk_trend','chat'].includes(state.activeTab) ? 'text-emerald-600' : 'text-slate-400']">首页</span>
    </button>
    <button @click="state.activeTab = 'alerts'" class="flex flex-col items-center justify-center w-16 h-full gap-1 transition-all duration-200 active:scale-95 group relative">
      <div :class="['w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300', state.activeTab === 'alerts' ? 'bg-emerald-50 text-emerald-600' : 'text-slate-400 group-hover:bg-slate-50']">
        <svg class="w-[22px] h-[22px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path></svg>
        <span v-if="state.alertUnreadCount > 0" class="absolute top-1 right-2 w-2.5 h-2.5 bg-rose-500 rounded-full border-2 border-white"></span>
      </div>
      <span :class="['text-[10px] font-bold transition-colors', state.activeTab === 'alerts' ? 'text-emerald-600' : 'text-slate-400']">消息</span>
    </button>
    <button @click="state.activeTab = 'submit'" class="flex flex-col items-center justify-center w-16 h-full transition-all duration-200 group relative z-10">
      <div :class="['w-14 h-14 rounded-full flex items-center justify-center -mt-6 shadow-xl transition-transform duration-300', state.activeTab === 'submit' ? 'bg-emerald-500 text-white shadow-emerald-500/30 scale-105' : 'bg-slate-900 text-white shadow-slate-900/20 group-hover:scale-105 group-active:scale-95']">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold mt-1 transition-colors', state.activeTab === 'submit' ? 'text-emerald-600' : 'text-slate-400']">检测</span>
    </button>
    <button @click="state.activeTab = 'family'" class="flex flex-col items-center justify-center w-16 h-full gap-1 transition-all duration-200 active:scale-95 group">
      <div :class="['w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300', ['family','family_invite'].includes(state.activeTab) ? 'bg-emerald-50 text-emerald-600' : 'text-slate-400 group-hover:bg-slate-50']">
        <svg class="w-[22px] h-[22px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold transition-colors', ['family','family_invite'].includes(state.activeTab) ? 'text-emerald-600' : 'text-slate-400']">家庭</span>
    </button>
    <button @click="state.activeTab = 'profile'" class="flex flex-col items-center justify-center w-16 h-full gap-1 transition-all duration-200 active:scale-95 group">
      <div :class="['w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300', ['profile','profile_privacy'].includes(state.activeTab) ? 'bg-emerald-50 text-emerald-600' : 'text-slate-400 group-hover:bg-slate-50']">
        <svg class="w-[22px] h-[22px]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold transition-colors', ['profile','profile_privacy'].includes(state.activeTab) ? 'text-emerald-600' : 'text-slate-400']">我的</span>
    </button>
  </nav>
</template>

<script setup>
defineProps({
  state: {
    type: Object,
    required: true
  }
});
</script>
