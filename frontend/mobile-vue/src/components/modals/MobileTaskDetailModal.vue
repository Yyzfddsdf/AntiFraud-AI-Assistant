<template>
  <div v-if="state.selectedTask" class="fixed inset-0 z-[1100] flex items-end bg-black/50 backdrop-blur-sm" @click.self="state.close">
    <div class="bg-white w-full h-[92vh] rounded-t-2xl overflow-hidden flex flex-col animate-slide-up">
      <div class="p-4 border-b border-gray-100 flex justify-between items-center sticky top-0 bg-white z-10">
        <div>
          <h3 class="font-bold text-lg">任务详情</h3>
          <p class="text-xs text-slate-400 font-mono">ID: {{ state.selectedTask.task_id }}</p>
        </div>
        <button @click="state.close" class="text-gray-400 p-2"><svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button>
      </div>
      <div class="flex-1 overflow-y-auto p-5 space-y-5 pb-24 pb-safe" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <div class="flex items-center justify-between flex-wrap gap-2">
          <span :class="state.getStatusClass(state.selectedTask.status)">{{ state.getStatusLabel(state.selectedTask.status) }}</span>
          <span class="text-xs text-slate-500">{{ state.formatTime(state.selectedTask.created_at) }}</span>
        </div>

        <div class="flex justify-end">
          <div class="m-dropdown" data-custom-dropdown>
            <button type="button" @click="state.toggleDropdown('task-export-menu')" :class="['w-11 h-11 rounded-full flex items-center justify-center transition-all', state.openDropdownKey === 'task-export-menu' ? 'text-emerald-600' : 'text-slate-900']">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
            </button>
            <transition name="fade">
              <div v-if="state.openDropdownKey === 'task-export-menu'" class="m-dropdown-menu !right-0 !left-auto !w-44">
                <button type="button" @click="state.exportData('md'); state.closeDropdown()" class="m-dropdown-option">
                  <div>
                    <div class="text-sm font-semibold text-slate-900">Markdown</div>
                    <div class="text-[11px] text-slate-400 mt-1">导出结构化文本报告</div>
                  </div>
                </button>
                <button type="button" @click="state.exportData('json'); state.closeDropdown()" class="m-dropdown-option">
                  <div>
                    <div class="text-sm font-semibold text-slate-900">JSON</div>
                    <div class="text-[11px] text-slate-400 mt-1">导出完整结构化数据</div>
                  </div>
                </button>
              </div>
            </transition>
          </div>
        </div>

        <div v-if="state.selectedTask.summary" class="p-4 bg-slate-50 rounded-xl border border-slate-100">
          <h4 class="font-bold text-sm mb-2 text-slate-700">案件摘要</h4>
          <div class="text-sm leading-relaxed whitespace-pre-wrap text-slate-700">{{ state.selectedTask.summary }}</div>
        </div>

        <div v-if="state.selectedTask.risk_score || state.selectedTask.risk_summary" class="p-4 bg-slate-50 rounded-xl border border-slate-100 space-y-3">
          <div class="flex items-center justify-between">
            <h4 class="font-bold text-sm text-slate-700">风险评分</h4>
            <span class="px-2.5 py-1 rounded-full bg-emerald-50 border border-emerald-100 text-emerald-700 text-xs font-extrabold">{{ state.selectedTask.risk_score || 0 }}</span>
          </div>
          <div v-if="state.parseRiskSummary(state.selectedTask.risk_summary)" class="space-y-3">
            <div class="rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 space-y-2">
              <div class="flex items-center justify-between">
                <span>社工话术</span>
                <span class="font-bold">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.social_engineering || 0 }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span>诱导动作</span>
                <span class="font-bold">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.requested_actions || 0 }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span>证据强度</span>
                <span class="font-bold">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.evidence_strength || 0 }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span>受害暴露</span>
                <span class="font-bold">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.loss_exposure || 0 }}</span>
              </div>
            </div>
            <div v-if="state.parseRiskSummary(state.selectedTask.risk_summary).hit_rules && state.parseRiskSummary(state.selectedTask.risk_summary).hit_rules.length" class="flex flex-wrap gap-2">
              <span v-for="item in state.parseRiskSummary(state.selectedTask.risk_summary).hit_rules" :key="`risk-rule-${item}`" class="px-2.5 py-1 rounded-full bg-white border border-slate-200 text-xs font-medium text-slate-600">{{ item }}</span>
            </div>
          </div>
          <div v-else class="text-xs text-slate-600 whitespace-pre-wrap">{{ state.selectedTask.risk_summary }}</div>
        </div>

        <div v-if="state.selectedTask.report" class="space-y-3">
          <div class="p-4 bg-slate-50 rounded-xl border border-slate-100">
            <h4 class="font-bold text-sm mb-3 text-slate-700">综合分析报告</h4>
            <div v-for="section in state.parseReport(state.selectedTask.report)" :key="section.id" class="mb-3 last:mb-0">
              <div class="text-sm font-bold text-slate-900">{{ section.title }}</div>
              <div class="mt-1 text-sm text-slate-600 leading-relaxed whitespace-pre-wrap">{{ section.content }}</div>
            </div>
            <div v-if="state.parseReport(state.selectedTask.report).length === 0" class="text-sm leading-relaxed whitespace-pre-wrap text-slate-700">{{ state.selectedTask.report }}</div>
          </div>

          <div v-if="state.extractAttackSteps(state.selectedTask.report).length > 0" class="p-4 bg-rose-50 rounded-xl border border-rose-100">
            <h4 class="font-bold text-sm mb-3 text-rose-700">诈骗链路时间线</h4>
            <div class="space-y-2">
              <div v-for="(step, idx) in state.extractAttackSteps(state.selectedTask.report)" :key="'attack-'+idx" class="text-sm text-slate-700 leading-relaxed">
                {{ idx + 1 }}. {{ step }}
              </div>
            </div>
          </div>

          <div v-if="state.extractScamKeywordSentences(state.selectedTask.report).length > 0" class="p-4 bg-fuchsia-50 rounded-xl border border-fuchsia-100">
            <h4 class="font-bold text-sm mb-3 text-fuchsia-700">诈骗关键词句</h4>
            <div class="flex flex-wrap gap-2">
              <span v-for="(keyword, idx) in state.extractScamKeywordSentences(state.selectedTask.report)" :key="'keyword-'+idx" class="text-xs font-bold text-fuchsia-700 bg-white px-2.5 py-1 rounded-full border border-fuchsia-100">
                {{ keyword }}
              </span>
            </div>
          </div>
        </div>

        <div v-if="state.selectedTask.risk_level" class="flex items-center gap-2">
          <span class="text-sm font-bold text-slate-600">风险等级:</span>
          <span :class="['px-2 py-0.5 rounded text-xs font-bold border', state.getRiskClass(state.selectedTask.risk_level)]">{{ state.normalizeRiskLevelText(state.selectedTask.risk_level) }}</span>
        </div>

        <div class="space-y-3" v-if="state.selectedTask.payload">
          <h4 class="font-bold text-sm text-slate-700">输入概览</h4>
          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">文本</div>
              <div class="mt-1 text-sm font-extrabold text-slate-900">{{ state.selectedTask.payload.text ? '已提交' : '无' }}</div>
            </div>
            <div class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">图片</div>
              <div class="mt-1 text-sm font-extrabold text-slate-900">{{ state.selectedTask.payload.images ? state.selectedTask.payload.images.length : 0 }} 份</div>
            </div>
            <div class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">音频</div>
              <div class="mt-1 text-sm font-extrabold text-slate-900">{{ state.selectedTask.payload.audios ? state.selectedTask.payload.audios.length : 0 }} 份</div>
            </div>
            <div class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">视频</div>
              <div class="mt-1 text-sm font-extrabold text-slate-900">{{ state.selectedTask.payload.videos ? state.selectedTask.payload.videos.length : 0 }} 份</div>
            </div>
          </div>
          <div class="rounded-xl border border-dashed border-slate-200 bg-white px-4 py-3 text-xs leading-5 text-slate-400">
            手机端主详情页只展示分析结果与输入概览，原始多模态材料不在这里展开。
          </div>
        </div>
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
