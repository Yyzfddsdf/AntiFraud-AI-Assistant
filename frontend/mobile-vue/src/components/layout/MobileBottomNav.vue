<template>
  <nav class="fixed bottom-0 left-0 right-0 w-full bg-white/80 backdrop-blur-2xl border-t border-slate-100/50 flex justify-around items-center h-[88px] pb-safe-bottom px-2 shadow-[0_-4px_24px_-8px_rgba(0,0,0,0.05)]" style="z-index: 999;">
    <!-- Home Tab -->
    <button @click="state.activeTab = 'tasks'" class="relative flex flex-col items-center justify-center w-[20%] h-full gap-1 transition-all duration-300 active:scale-90 group pt-2">
      <div :class="['w-12 h-10 rounded-2xl flex items-center justify-center transition-all duration-300 relative overflow-hidden', ['tasks','history','risk_trend','chat'].includes(state.activeTab) ? 'text-emerald-600' : 'text-slate-400 group-hover:text-slate-600']">
        <div v-if="['tasks','history','risk_trend','chat'].includes(state.activeTab)" class="absolute inset-0 bg-emerald-100/50 mix-blend-multiply transition-opacity"></div>
        <svg class="w-6 h-6 relative z-10 transition-transform duration-300" :class="['tasks','history','risk_trend','chat'].includes(state.activeTab) ? 'scale-110' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold tracking-wide transition-colors', ['tasks','history','risk_trend','chat'].includes(state.activeTab) ? 'text-emerald-700' : 'text-slate-500']">首页</span>
    </button>
    
    <!-- Alerts Tab -->
    <button @click="state.activeTab = 'alerts'" class="relative flex flex-col items-center justify-center w-[20%] h-full gap-1 transition-all duration-300 active:scale-90 group pt-2">
      <div :class="['w-12 h-10 rounded-2xl flex items-center justify-center transition-all duration-300 relative overflow-hidden', state.activeTab === 'alerts' ? 'text-emerald-600' : 'text-slate-400 group-hover:text-slate-600']">
        <div v-if="state.activeTab === 'alerts'" class="absolute inset-0 bg-emerald-100/50 mix-blend-multiply transition-opacity"></div>
        <svg class="w-6 h-6 relative z-10 transition-transform duration-300" :class="state.activeTab === 'alerts' ? 'scale-110' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path></svg>
        <span v-if="state.alertUnreadCount > 0" class="absolute top-1.5 right-2.5 w-2.5 h-2.5 bg-rose-500 rounded-full border-2 border-white shadow-sm z-20"></span>
      </div>
      <span :class="['text-[10px] font-bold tracking-wide transition-colors', state.activeTab === 'alerts' ? 'text-emerald-700' : 'text-slate-500']">消息</span>
    </button>
    
    <!-- Submit (Detect) Tab - Floating Center Button -->
    <button @click="state.activeTab = 'submit'" class="relative flex flex-col items-center justify-center w-[20%] h-full transition-all duration-300 group z-10 -mt-8">
      <div class="relative">
        <!-- Glow Effect -->
        <div :class="['absolute inset-0 rounded-full blur-xl transition-opacity duration-500', state.activeTab === 'submit' ? 'bg-emerald-500/40 opacity-100' : 'bg-slate-900/20 opacity-0 group-hover:opacity-100']"></div>
        
        <!-- Button Core -->
        <div :class="['relative w-[60px] h-[60px] rounded-[24px] rotate-3 flex items-center justify-center shadow-2xl transition-all duration-300', state.activeTab === 'submit' ? 'bg-gradient-to-tr from-emerald-500 to-emerald-400 text-white shadow-emerald-500/40 scale-105' : 'bg-gradient-to-tr from-slate-800 to-slate-900 text-white shadow-slate-900/30 group-hover:scale-105 group-active:scale-95']">
          <svg class="w-8 h-8 -rotate-3 transition-transform duration-300" :class="state.activeTab === 'submit' ? 'scale-110' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4"></path></svg>
        </div>
      </div>
      <span :class="['text-[10px] font-bold tracking-wide mt-2 transition-colors', state.activeTab === 'submit' ? 'text-emerald-700' : 'text-slate-500']">检测</span>
    </button>
    
    <!-- Family Tab -->
    <button @click="state.activeTab = 'family'" class="relative flex flex-col items-center justify-center w-[20%] h-full gap-1 transition-all duration-300 active:scale-90 group pt-2">
      <div :class="['w-12 h-10 rounded-2xl flex items-center justify-center transition-all duration-300 relative overflow-hidden', ['family','family_invite'].includes(state.activeTab) ? 'text-emerald-600' : 'text-slate-400 group-hover:text-slate-600']">
        <div v-if="['family','family_invite'].includes(state.activeTab)" class="absolute inset-0 bg-emerald-100/50 mix-blend-multiply transition-opacity"></div>
        <svg class="w-6 h-6 relative z-10 transition-transform duration-300" :class="['family','family_invite'].includes(state.activeTab) ? 'scale-110' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold tracking-wide transition-colors', ['family','family_invite'].includes(state.activeTab) ? 'text-emerald-700' : 'text-slate-500']">家庭</span>
    </button>
    
    <!-- Profile Tab -->
    <button @click="state.activeTab = 'profile'" class="relative flex flex-col items-center justify-center w-[20%] h-full gap-1 transition-all duration-300 active:scale-90 group pt-2">
      <div :class="['w-12 h-10 rounded-2xl flex items-center justify-center transition-all duration-300 relative overflow-hidden', ['profile','profile_privacy'].includes(state.activeTab) ? 'text-emerald-600' : 'text-slate-400 group-hover:text-slate-600']">
        <div v-if="['profile','profile_privacy'].includes(state.activeTab)" class="absolute inset-0 bg-emerald-100/50 mix-blend-multiply transition-opacity"></div>
        <svg class="w-6 h-6 relative z-10 transition-transform duration-300" :class="['profile','profile_privacy'].includes(state.activeTab) ? 'scale-110' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
      </div>
      <span :class="['text-[10px] font-bold tracking-wide transition-colors', ['profile','profile_privacy'].includes(state.activeTab) ? 'text-emerald-700' : 'text-slate-500']">我的</span>
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
