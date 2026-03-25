<template>
  <div v-if="state.familyAlertModalVisible && state.activeFamilyNotification" class="fixed inset-0 z-[1002] flex items-end bg-black/50 backdrop-blur-sm" @click.self="state.acknowledgeFamilyAlert">
    <div class="bg-white w-full max-h-[92vh] min-h-0 rounded-t-2xl overflow-hidden flex flex-col animate-slide-up">
      <div class="shrink-0 p-4 border-b border-gray-100 flex justify-between items-center">
        <h3 class="font-bold text-lg">家庭联防通知</h3>
        <button @click="state.acknowledgeFamilyAlert" class="text-gray-400 p-2"><svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button>
      </div>
      <div class="flex-1 min-h-0 overflow-y-auto p-5 space-y-4" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <div class="flex items-center gap-3 mb-2">
          <span class="px-2 py-1 bg-rose-100 text-rose-600 text-xs font-bold rounded uppercase">{{ state.activeFamilyNotification.risk_level || '高' }} Risk</span>
          <span class="text-xs text-gray-500">{{ state.formatTime(state.activeFamilyNotification.event_at) }}</span>
        </div>
        <h2 class="text-xl font-bold leading-tight">{{ state.activeFamilyNotification.title || '高风险案件预警' }}</h2>
        <div class="p-4 bg-rose-50 rounded-xl text-sm leading-relaxed text-gray-700">{{ state.activeFamilyNotification.case_summary || state.activeFamilyNotification.summary }}</div>
        <div class="grid grid-cols-2 gap-4 text-xs text-gray-500">
          <div><div class="font-bold mb-1">家庭成员</div><div>{{ state.activeFamilyNotification.target_name }}</div></div>
          <div><div class="font-bold mb-1">诈骗类型</div><div>{{ state.activeFamilyNotification.scam_type || '待分析' }}</div></div>
        </div>
      </div>
      <div class="shrink-0 p-4 border-t border-gray-100 bg-gray-50 flex gap-2" style="padding-bottom: max(1rem, env(safe-area-inset-bottom));">
        <button @click="state.acknowledgeFamilyAlert" class="flex-1 text-sm font-bold text-slate-700 bg-white border border-slate-200 rounded-xl py-3">稍后处理</button>
        <button @click="state.openFamilyNotificationCenter" class="flex-1 m-btn-primary bg-red-600 text-white text-sm">进入家庭中心</button>
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
