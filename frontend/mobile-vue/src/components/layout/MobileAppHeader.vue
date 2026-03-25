<template>
  <header class="h-[68px] px-5 pt-safe flex items-center justify-between bg-white/80 backdrop-blur-xl shrink-0 sticky top-0 z-30 border-b border-slate-100/50">
    <div class="flex items-center gap-3">
      <div class="w-11 h-11 rounded-[14px] bg-gradient-to-br from-emerald-400 to-teal-500 flex items-center justify-center text-white shadow-lg shadow-emerald-500/20 ring-1 ring-white/50">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
      </div>
      <div class="flex flex-col">
        <span class="font-black text-xl tracking-tight text-slate-900 leading-none">反诈卫士</span>
        <span class="text-[10px] font-bold text-emerald-600 uppercase tracking-widest mt-0.5">Sentinel AI</span>
      </div>
    </div>
    <button v-if="state.activeTab !== 'alerts'" @click="state.activeTab = 'alerts'" class="relative w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center text-slate-600 active:scale-90 transition-transform">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path></svg>
      <span v-if="state.alertUnreadCount > 0" class="absolute top-2.5 right-2.5 w-2.5 h-2.5 bg-rose-500 rounded-full border-2 border-white"></span>
    </button>
  </header>
</template>

<script setup>
defineProps({
  state: {
    type: Object,
    required: true
  }
});
</script>
