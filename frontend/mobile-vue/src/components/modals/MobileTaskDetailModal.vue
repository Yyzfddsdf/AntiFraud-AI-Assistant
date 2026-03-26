<template>
  <div v-if="state.selectedTask" class="fixed inset-0 z-[1100] flex flex-col justify-end bg-black/40 backdrop-blur-sm transition-all" @click.self="state.close">
    <div class="bg-slate-50 w-full h-[90vh] rounded-t-[32px] overflow-hidden flex flex-col animate-slide-up shadow-2xl relative">
      <!-- Handle & Header -->
      <div class="bg-white/80 backdrop-blur-md sticky top-0 z-20 border-b border-slate-100/80 pt-3 pb-3 px-4 flex flex-col items-center justify-center">
        <div class="w-12 h-1.5 bg-slate-200 rounded-full mb-3"></div>
        <div class="w-full flex justify-between items-center">
          <div class="flex-1">
            <h3 class="font-extrabold text-[17px] text-slate-900 tracking-tight">案件分析详情</h3>
            <p class="text-[11px] text-slate-400 font-mono mt-0.5">ID: {{ state.selectedTask.task_id }}</p>
          </div>
          <button @click="state.close" class="w-8 h-8 flex items-center justify-center bg-slate-100 text-slate-500 rounded-full active:scale-95 transition-transform"><svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"></path></svg></button>
        </div>
      </div>
      
      <!-- Content -->
      <div class="flex-1 overflow-y-auto px-4 py-5 space-y-4 pb-32 pb-safe" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <!-- Status & Time -->
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-2">
            <span :class="state.getStatusClass(state.selectedTask.status).replace('rounded-full', 'rounded-lg tracking-widest uppercase')">{{ state.getStatusLabel(state.selectedTask.status) }}</span>
            <div class="w-1 h-1 rounded-full bg-slate-300"></div>
            <span class="text-[11px] text-slate-500 font-medium">{{ state.formatTime(state.selectedTask.created_at) }}</span>
          </div>
          <div class="m-dropdown" data-custom-dropdown>
            <button type="button" @click="state.toggleDropdown('task-export-menu')" :class="['w-8 h-8 rounded-full bg-white shadow-sm border border-slate-100 flex items-center justify-center transition-all', state.openDropdownKey === 'task-export-menu' ? 'text-emerald-600 bg-emerald-50 border-emerald-100' : 'text-slate-600']">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path></svg>
            </button>
            <transition name="fade">
              <div v-if="state.openDropdownKey === 'task-export-menu'" class="m-dropdown-menu !right-0 !left-auto !w-44 !mt-1 shadow-lg rounded-2xl border-slate-100">
                <button type="button" @click="state.exportData('md'); state.closeDropdown()" class="m-dropdown-option px-4 py-3 border-b border-slate-50">
                  <div class="text-left">
                    <div class="text-[13px] font-bold text-slate-900">Markdown</div>
                    <div class="text-[10px] text-slate-500 mt-0.5">导出结构化文本报告</div>
                  </div>
                </button>
                <button type="button" @click="state.exportData('json'); state.closeDropdown()" class="m-dropdown-option px-4 py-3">
                  <div class="text-left">
                    <div class="text-[13px] font-bold text-slate-900">JSON</div>
                    <div class="text-[10px] text-slate-500 mt-0.5">导出完整结构化数据</div>
                  </div>
                </button>
              </div>
            </transition>
          </div>
        </div>

        <!-- Risk Score Card -->
        <div v-if="state.selectedTask.risk_score || state.selectedTask.risk_summary" class="bg-white rounded-[24px] shadow-sm border border-slate-100/60 p-5 relative overflow-hidden">
          <div class="absolute top-0 right-0 w-32 h-32 opacity-10 pointer-events-none rounded-bl-full" :class="state.getRiskClass(state.selectedTask.risk_level).includes('red') ? 'bg-red-500' : (state.getRiskClass(state.selectedTask.risk_level).includes('yellow') ? 'bg-amber-500' : (state.getRiskClass(state.selectedTask.risk_level).includes('green') ? 'bg-emerald-500' : 'bg-slate-400'))"></div>
          
          <div class="flex items-start justify-between relative z-10 mb-4">
            <div>
              <div class="text-[11px] font-bold text-slate-400 tracking-[0.2em] uppercase mb-1">风险评估</div>
              <div class="flex items-end gap-2">
                <span class="text-4xl font-black leading-none tracking-tighter text-slate-900">{{ state.selectedTask.risk_score || 0 }}</span>
                <span v-if="state.selectedTask.risk_level" class="text-xs font-bold mb-1" :class="['px-2 py-0.5 rounded-md border', state.getRiskClass(state.selectedTask.risk_level)]">{{ state.normalizeRiskLevelText(state.selectedTask.risk_level) }}</span>
              </div>
            </div>
          </div>

          <div v-if="state.parseRiskSummary(state.selectedTask.risk_summary)" class="space-y-4 relative z-10">
            <div class="grid grid-cols-2 gap-2">
              <div class="bg-slate-50 rounded-2xl p-3 flex flex-col">
                <span class="text-[10px] text-slate-400 font-bold mb-1">社工话术</span>
                <span class="text-[15px] font-black text-slate-800">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.social_engineering || 0 }}</span>
              </div>
              <div class="bg-slate-50 rounded-2xl p-3 flex flex-col">
                <span class="text-[10px] text-slate-400 font-bold mb-1">诱导动作</span>
                <span class="text-[15px] font-black text-slate-800">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.requested_actions || 0 }}</span>
              </div>
              <div class="bg-slate-50 rounded-2xl p-3 flex flex-col">
                <span class="text-[10px] text-slate-400 font-bold mb-1">证据强度</span>
                <span class="text-[15px] font-black text-slate-800">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.evidence_strength || 0 }}</span>
              </div>
              <div class="bg-slate-50 rounded-2xl p-3 flex flex-col">
                <span class="text-[10px] text-slate-400 font-bold mb-1">受害暴露</span>
                <span class="text-[15px] font-black text-slate-800">{{ state.parseRiskSummary(state.selectedTask.risk_summary).dimensions?.loss_exposure || 0 }}</span>
              </div>
            </div>
            <div v-if="state.parseRiskSummary(state.selectedTask.risk_summary).hit_rules && state.parseRiskSummary(state.selectedTask.risk_summary).hit_rules.length" class="flex flex-wrap gap-1.5 pt-1 border-t border-slate-50">
              <span v-for="item in state.parseRiskSummary(state.selectedTask.risk_summary).hit_rules" :key="`risk-rule-${item}`" class="px-2.5 py-1 rounded-lg bg-rose-50/50 text-rose-600 text-[10px] font-bold border border-rose-100/50">{{ item }}</span>
            </div>
          </div>
          <div v-else class="text-xs text-slate-600 whitespace-pre-wrap leading-relaxed relative z-10">{{ state.selectedTask.risk_summary }}</div>
        </div>

        <!-- Summary -->
        <div v-if="state.selectedTask.summary" class="bg-white rounded-[24px] shadow-sm border border-slate-100/60 p-5">
          <div class="text-[11px] font-bold text-slate-400 tracking-[0.2em] uppercase mb-3 flex items-center gap-2">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            案件摘要
          </div>
          <div class="text-[13px] leading-relaxed whitespace-pre-wrap text-slate-700 font-medium">{{ state.selectedTask.summary }}</div>
        </div>

        <!-- Report Timeline / Details -->
        <div v-if="state.selectedTask.report" class="space-y-4">
          <div v-if="state.extractAttackSteps(state.selectedTask.report).length > 0" class="bg-gradient-to-br from-rose-50 to-white rounded-[24px] shadow-sm border border-rose-100/60 p-5">
            <div class="text-[11px] font-bold text-rose-400 tracking-[0.2em] uppercase mb-4 flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
              诈骗链路时间线
            </div>
            <div class="space-y-4 relative before:absolute before:inset-0 before:ml-[9px] before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-rose-200 before:to-transparent">
              <div v-for="(step, idx) in state.extractAttackSteps(state.selectedTask.report)" :key="'attack-'+idx" class="relative flex items-start gap-4">
                <div class="w-5 h-5 rounded-full bg-white border-2 border-rose-400 shrink-0 flex items-center justify-center shadow-sm z-10 mt-0.5">
                  <div class="w-1.5 h-1.5 bg-rose-500 rounded-full"></div>
                </div>
                <div class="text-[13px] text-slate-800 leading-relaxed font-medium bg-white/60 px-3 py-2 rounded-xl border border-rose-100/30">
                  {{ step }}
                </div>
              </div>
            </div>
          </div>

          <div v-if="state.extractScamKeywordSentences(state.selectedTask.report).length > 0" class="bg-gradient-to-br from-fuchsia-50 to-white rounded-[24px] shadow-sm border border-fuchsia-100/60 p-5">
            <div class="text-[11px] font-bold text-fuchsia-400 tracking-[0.2em] uppercase mb-3 flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"></path></svg>
              高危关键词句
            </div>
            <div class="flex flex-wrap gap-2">
              <span v-for="(keyword, idx) in state.extractScamKeywordSentences(state.selectedTask.report)" :key="'keyword-'+idx" class="text-[11px] font-bold text-fuchsia-700 bg-white px-3 py-1.5 rounded-xl border border-fuchsia-100/60 shadow-sm">
                {{ keyword }}
              </span>
            </div>
          </div>

          <div class="bg-white rounded-[24px] shadow-sm border border-slate-100/60 p-5">
            <div class="text-[11px] font-bold text-slate-400 tracking-[0.2em] uppercase mb-4 flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
              综合分析报告
            </div>
            <div v-for="section in state.parseReport(state.selectedTask.report)" :key="section.id" class="mb-4 last:mb-0">
              <div class="text-[13px] font-extrabold text-slate-900 mb-1.5">{{ section.title }}</div>
              <div class="text-[13px] text-slate-600 leading-relaxed whitespace-pre-wrap font-medium">{{ section.content }}</div>
            </div>
            <div v-if="state.parseReport(state.selectedTask.report).length === 0" class="text-[13px] leading-relaxed whitespace-pre-wrap text-slate-600 font-medium">{{ state.selectedTask.report }}</div>
          </div>
        </div>

        <!-- Input Overview -->
        <div class="bg-white rounded-[24px] shadow-sm border border-slate-100/60 p-5" v-if="state.selectedTask.payload">
          <div class="text-[11px] font-bold text-slate-400 tracking-[0.2em] uppercase mb-4 flex items-center gap-2">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
            输入概览
          </div>
          <div class="grid grid-cols-2 gap-2 mb-3">
            <div class="rounded-2xl border border-slate-100 bg-slate-50/50 p-3 flex justify-between items-center">
              <span class="text-[11px] font-bold text-slate-500">文本</span>
              <span class="text-[13px] font-black text-slate-800">{{ state.selectedTask.payload.text ? '已提交' : '无' }}</span>
            </div>
            <div class="rounded-2xl border border-slate-100 bg-slate-50/50 p-3 flex justify-between items-center">
              <span class="text-[11px] font-bold text-slate-500">图片</span>
              <span class="text-[13px] font-black text-slate-800">{{ state.selectedTask.payload.images ? state.selectedTask.payload.images.length : 0 }} 份</span>
            </div>
            <div class="rounded-2xl border border-slate-100 bg-slate-50/50 p-3 flex justify-between items-center">
              <span class="text-[11px] font-bold text-slate-500">音频</span>
              <span class="text-[13px] font-black text-slate-800">{{ state.selectedTask.payload.audios ? state.selectedTask.payload.audios.length : 0 }} 份</span>
            </div>
            <div class="rounded-2xl border border-slate-100 bg-slate-50/50 p-3 flex justify-between items-center">
              <span class="text-[11px] font-bold text-slate-500">视频</span>
              <span class="text-[13px] font-black text-slate-800">{{ state.selectedTask.payload.videos ? state.selectedTask.payload.videos.length : 0 }} 份</span>
            </div>
          </div>
          <div class="rounded-xl border border-dashed border-slate-200 bg-slate-50/30 px-4 py-3 text-[11px] font-medium leading-relaxed text-slate-400 text-center">
            原始多模态材料受限于移动端屏幕尺寸，不在此处展开展示。
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
