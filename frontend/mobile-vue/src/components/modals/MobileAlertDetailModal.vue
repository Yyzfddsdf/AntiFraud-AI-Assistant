<template>
  <div v-if="state.activeAlertEvent" class="fixed inset-0 z-[1001] flex items-end justify-center sm:items-center p-0 sm:p-4 bg-black/50 backdrop-blur-sm" @click.self="state.acknowledgeActiveAlert">
    <div class="bg-white w-full max-h-[92vh] min-h-0 rounded-t-2xl sm:rounded-2xl overflow-hidden flex flex-col animate-slide-up">
      <div class="shrink-0 p-4 border-b border-gray-100 flex justify-between items-center">
        <h3 class="font-bold text-lg">风险预警</h3>
        <button @click="state.acknowledgeActiveAlert" class="text-gray-400 p-2"><svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button>
      </div>
      <div class="flex-1 min-h-0 overflow-y-auto p-5 space-y-4" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <div class="flex items-center gap-3 mb-2">
          <span :class="['px-2 py-1 text-xs font-bold rounded uppercase', state.getAlertSeverityTheme(state.activeAlertEvent.risk_level).modalBadgeClass]">{{ state.activeAlertEvent.risk_level }} Risk</span>
          <span class="text-xs text-gray-500">{{ state.formatTime(state.activeAlertEvent.created_at) }}</span>
        </div>
        <h2 class="text-xl font-bold leading-tight">{{ state.activeAlertEvent.title }}</h2>
        <div :class="['p-4 rounded-xl text-sm leading-relaxed text-gray-700', state.getAlertSeverityTheme(state.activeAlertEvent.risk_level).panelClass]">{{ state.activeAlertEvent.case_summary }}</div>
        <div class="grid grid-cols-2 gap-4 text-xs text-gray-500">
          <div><div class="font-bold mb-1">诈骗类型</div><div>{{ state.activeAlertEvent.scam_type }}</div></div>
          <div><div class="font-bold mb-1">案件ID</div><div class="truncate">{{ state.activeAlertEvent.record_id }}</div></div>
        </div>
      </div>
      <div class="shrink-0 p-4 border-t border-gray-100 bg-gray-50 flex gap-2" style="padding-bottom: max(1rem, env(safe-area-inset-bottom));">
        <button @click="state.acknowledgeActiveAlert" class="flex-1 text-sm font-bold text-slate-700 bg-white border border-slate-200 rounded-xl py-3">稍后处理</button>
        <button @click="state.openAlertHistory" :class="['flex-1 text-white text-sm font-bold rounded-xl py-3', state.getAlertSeverityTheme(state.activeAlertEvent.risk_level).actionClass]">查看案件</button>
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
