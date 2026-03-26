<template>
  <div v-if="state.familyAlertModalVisible && state.activeFamilyNotification" class="fixed inset-0 z-[1002] flex items-end bg-slate-900/60 backdrop-blur-sm transition-all duration-300" @click.self="state.acknowledgeFamilyAlert">
    <div class="bg-white w-full max-h-[92vh] min-h-0 rounded-t-[32px] overflow-hidden flex flex-col animate-slide-up shadow-[0_-10px_40px_rgba(0,0,0,0.1)] relative">
      <!-- Handle -->
      <div class="absolute top-0 left-0 right-0 flex justify-center pt-3 pb-1 z-20">
        <div class="w-12 h-1.5 bg-gray-200/80 rounded-full"></div>
      </div>

      <!-- Header -->
      <div class="pt-6 pb-4 px-5 border-b border-gray-100 flex justify-between items-center sticky top-0 bg-white/95 backdrop-blur-md z-10">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full bg-rose-100 flex items-center justify-center">
            <svg class="w-4 h-4 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
          </div>
          <h3 class="font-[800] text-xl text-slate-800 tracking-tight">家庭联防通知</h3>
        </div>
        <button @click="state.acknowledgeFamilyAlert" class="text-slate-400 hover:text-slate-600 bg-slate-50 hover:bg-slate-100 p-2 rounded-full transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"></path></svg>
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 min-h-0 overflow-y-auto p-6 space-y-6" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <div>
          <div class="flex items-center gap-3 mb-3">
            <span class="px-2.5 py-1 bg-rose-100 text-rose-700 text-[11px] font-bold rounded-lg uppercase tracking-wider border border-rose-200/50 shadow-sm">{{ state.activeFamilyNotification.risk_level || '高' }} 风险</span>
            <span class="text-xs text-slate-400 font-medium">{{ state.formatTime(state.activeFamilyNotification.event_at) }}</span>
          </div>
          <h2 class="text-2xl font-[800] text-slate-800 leading-tight mb-4">{{ state.activeFamilyNotification.title || '高风险案件预警' }}</h2>
          
          <div class="p-4 bg-gradient-to-br from-rose-50 to-red-50/30 rounded-2xl text-[15px] leading-relaxed text-slate-700 border border-rose-100/50 shadow-sm relative overflow-hidden">
            <div class="absolute -right-4 -top-4 text-rose-500/5 opacity-50">
              <svg class="w-24 h-24" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"></path></svg>
            </div>
            <div class="relative z-10">{{ state.activeFamilyNotification.case_summary || state.activeFamilyNotification.summary }}</div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100/60">
            <div class="text-xs font-bold text-slate-400 mb-1.5 uppercase tracking-wider">家庭成员</div>
            <div class="font-bold text-slate-800 text-[15px]">{{ state.activeFamilyNotification.target_name }}</div>
          </div>
          <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100/60">
            <div class="text-xs font-bold text-slate-400 mb-1.5 uppercase tracking-wider">诈骗类型</div>
            <div class="font-bold text-slate-800 text-[15px]">{{ state.activeFamilyNotification.scam_type || '待分析' }}</div>
          </div>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="shrink-0 p-5 border-t border-slate-100 bg-white flex gap-3 shadow-[0_-4px_20px_rgba(0,0,0,0.02)] relative z-10" style="padding-bottom: max(1.25rem, env(safe-area-inset-bottom));">
        <button @click="state.acknowledgeFamilyAlert" class="flex-[0.4] text-[15px] font-bold text-slate-600 bg-slate-100 hover:bg-slate-200 active:bg-slate-300 rounded-2xl py-3.5 transition-colors">
          稍后处理
        </button>
        <button @click="state.openFamilyNotificationCenter" class="flex-1 bg-rose-600 hover:bg-rose-700 active:bg-rose-800 text-white font-bold text-[15px] rounded-2xl py-3.5 shadow-md shadow-rose-600/20 transition-all flex items-center justify-center gap-2">
          <span>进入家庭中心</span>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7"></path></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  state: {
    type: Object,
    required: true
  }
});
</script>
