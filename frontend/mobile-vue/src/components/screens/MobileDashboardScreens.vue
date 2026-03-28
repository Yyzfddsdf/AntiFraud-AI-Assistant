<template>
  <div v-show="state.activeTab === 'tasks'" class="bg-white pb-[24px]">
    <section class="px-5 mt-[16px]">
      <div class="bg-slate-50 border border-slate-100 rounded-[32px] px-[20px] py-[20px] flex items-center justify-between overflow-hidden relative">
        <div class="relative z-10">
          <div class="flex items-center gap-1.5 mb-1">
            <span class="w-1.5 h-1.5 bg-primary rounded-full"></span>
            <span class="text-[10px] font-bold text-primary uppercase tracking-tight">System Secured</span>
          </div>
          <h2 class="text-[20px] font-bold text-slate-800 leading-tight">全域实时守护中</h2>
          <p class="text-[12px] text-slate-400 mt-1">累计检测 {{ state.riskStatsSummary.total }} 次风险内容</p>
        </div>

        <div class="relative w-20 h-20 flex items-center justify-center">
          <div class="absolute inset-0 bg-primary/10 rounded-full safe-pulse"></div>
          <div class="w-14 h-14 bg-white rounded-full shadow-lg flex items-center justify-center border-2 border-primary/20">
            <i data-lucide="check" class="text-primary" size="28"></i>
          </div>
        </div>
      </div>
    </section>

    <section class="px-5 mt-[24px]">
      <div class="grid grid-cols-5 gap-2">
        <div class="flex flex-col items-center gap-2 cursor-pointer" @click="state.activeTab = 'chat'">
          <div class="w-12 h-12 bg-emerald-50 text-emerald-600 rounded-[16px] flex items-center justify-center shadow-sm border border-emerald-100/50">
            <i data-lucide="bot" size="20"></i>
          </div>
          <span class="text-[11px] font-medium text-slate-600">AI助手</span>
        </div>

        <div class="flex flex-col items-center gap-2 cursor-pointer" @click="state.activeTab = 'history'">
          <div class="w-12 h-12 bg-blue-50 text-blue-600 rounded-[16px] flex items-center justify-center shadow-sm border border-blue-100/50">
            <i data-lucide="calendar" size="20"></i>
          </div>
          <span class="text-[11px] font-medium text-slate-600">历史</span>
        </div>

        <div class="flex flex-col items-center gap-2 cursor-pointer" @click="state.activeTab = 'risk_trend'">
          <div class="w-12 h-12 bg-amber-50 text-amber-600 rounded-[16px] flex items-center justify-center shadow-sm border border-amber-100/50">
            <i data-lucide="line-chart" size="20"></i>
          </div>
          <span class="text-[11px] font-medium text-slate-600">趋势</span>
        </div>

        <div class="flex flex-col items-center gap-2 cursor-pointer" @click="state.activeTab = 'family'">
          <div class="w-12 h-12 bg-indigo-50 text-indigo-600 rounded-[16px] flex items-center justify-center shadow-sm border border-indigo-100/50">
            <i data-lucide="heart" size="20"></i>
          </div>
          <span class="text-[11px] font-medium text-slate-600">守护</span>
        </div>

        <div class="flex flex-col items-center gap-2" @click="state.activeTab = 'simulation_quiz'">
          <div class="w-12 h-12 bg-rose-50 text-rose-600 rounded-[16px] flex items-center justify-center shadow-sm border border-rose-100/50 cursor-pointer active:scale-95 transition-transform">
            <i data-lucide="play-circle" size="20"></i>
          </div>
          <span class="text-[11px] font-bold text-slate-600">演练</span>
        </div>
      </div>
    </section>

    <section class="px-5 mt-[24px] grid grid-cols-2 gap-3">
      <div class="px-[16px] py-[16px] bg-white border border-slate-100 rounded-[24px] shadow-sm">
        <div class="w-7 h-7 bg-emerald-50 text-emerald-600 rounded-lg flex items-center justify-center mb-2">
          <i data-lucide="search" size="14"></i>
        </div>
        <div class="text-[22px] font-bold text-slate-900 leading-none">{{ state.riskStatsSummary.total }}</div>
        <div class="text-[11px] text-slate-400 mt-1">总检测数量</div>
      </div>

      <div class="px-[16px] py-[16px] bg-white border border-slate-100 rounded-[24px] shadow-sm">
        <div class="w-7 h-7 bg-rose-50 text-rose-600 rounded-lg flex items-center justify-center mb-2">
          <i data-lucide="alert-octagon" size="14"></i>
        </div>
        <div class="text-[22px] font-bold text-slate-900 leading-none">{{ state.riskStatsSummary.high }}</div>
        <div class="text-[11px] text-slate-400 mt-1">高危预警</div>
      </div>
    </section>

    <section class="px-5 mt-[24px]">
      <div class="flex justify-between items-center mb-3">
        <h3 class="text-[15px] font-bold text-slate-800 flex items-center gap-2">
          <span class="w-1 h-4 bg-primary rounded-full"></span>
          风险趋势
        </h3>
        <button type="button" class="text-[11px] font-bold text-primary" @click="state.activeTab = 'risk_trend'">详情 ></button>
      </div>

      <div class="premium-gradient rounded-[28px] px-[20px] py-[20px] text-white relative overflow-hidden">
        <div class="flex items-center gap-2 mb-3">
          <div class="px-2 py-0.5 bg-indigo-500/30 rounded-md text-[9px] font-bold uppercase tracking-wider border border-indigo-500/40">AI Insight</div>
        </div>
        <h4 class="text-[16px] font-bold mb-2">AI 研判：整体风险平稳</h4>
        <p class="text-[11px] text-slate-400 leading-relaxed mb-4">基于大数据分析，当前欺诈指数处于低位，建议保持日常防范习惯。</p>
        
        <div class="grid grid-cols-2 gap-4 pt-4 border-t border-white/5">
          <div>
            <p class="text-[9px] text-slate-500 uppercase font-bold mb-0.5">整体走势</p>
            <p class="text-[12px] font-semibold text-emerald-400">持续回落</p>
          </div>
          <div>
            <p class="text-[9px] text-slate-500 uppercase font-bold mb-0.5">当前风险</p>
            <p class="text-[12px] font-semibold text-white">低风险</p>
          </div>
        </div>
      </div>
    </section>

    <section class="px-5 mt-[24px]">
      <div class="flex justify-between items-center mb-3">
        <h3 class="text-[15px] font-bold text-slate-800 flex items-center gap-2">
          <span class="w-1 h-4 bg-primary rounded-full"></span>
          最近任务
        </h3>
        <button type="button" class="text-[11px] font-bold text-slate-400" @click="state.activeTab = 'history'">全部</button>
      </div>
      <div class="bg-slate-50 rounded-[24px] py-[40px] flex flex-col items-center justify-center border-2 border-dashed border-slate-200/60">
        <div class="w-12 h-12 bg-white rounded-full flex items-center justify-center text-slate-300 shadow-sm mb-3">
          <i data-lucide="inbox" size="24"></i>
        </div>
        <p class="text-[12px] text-slate-400 font-medium tracking-tight">暂无检测记录</p>
      </div>
    </section>
  </div>

  <div v-if="state.activeTab === 'history'" class="bg-slate-50 pb-24">
    <!-- Header -->
    <div class="sticky top-0 z-50 bg-white/80 backdrop-blur-lg border-b border-slate-100/80 pt-safe">
      <div class="flex items-center justify-between px-4 h-14">
        <button @click="state.activeTab = 'tasks'" class="w-10 h-10 flex items-center justify-center -ml-2 active:opacity-50 transition-opacity">
          <svg class="w-6 h-6 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <h2 class="text-[17px] font-bold text-slate-900 tracking-tight">历史档案</h2>
        <div class="w-10 flex justify-end">
          <span class="text-[10px] font-black bg-slate-100 text-slate-500 px-2 py-1 rounded-full">{{ state.history.length }}</span>
        </div>
      </div>
    </div>
    
    <!-- List -->
    <div class="px-4 py-5 space-y-4">
      <div v-for="item in state.history" :key="item.record_id" class="bg-white rounded-3xl p-4 shadow-sm border border-slate-100/50 active:scale-[0.98] transition-transform cursor-pointer relative overflow-hidden" @click="state.viewHistoryDetail(item)">
        <div class="absolute top-0 right-0 w-16 h-16 opacity-10 pointer-events-none rounded-bl-full" :class="state.getRiskClass(item.risk_level).includes('red') ? 'bg-red-500' : (state.getRiskClass(item.risk_level).includes('yellow') ? 'bg-amber-500' : (state.getRiskClass(item.risk_level).includes('green') ? 'bg-emerald-500' : 'bg-slate-400'))"></div>
        <div class="flex items-start justify-between gap-3 relative z-10">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-2.5">
              <span v-if="item.risk_level" :class="['px-2 py-0.5 rounded-lg text-[10px] font-black tracking-widest uppercase', state.getRiskClass(item.risk_level).replace('rounded-full', '')]">{{ state.normalizeRiskLevelText(item.risk_level) }}</span>
              <span class="text-[10px] text-slate-400 font-mono bg-slate-50 px-1.5 py-0.5 rounded-md">{{ String(item.record_id || '').slice(0, 8) }}</span>
            </div>
            <h3 class="text-[15px] font-bold text-slate-900 leading-snug line-clamp-2 pr-2">{{ item.title || '无标题检测记录' }}</h3>
            <div class="mt-3 flex items-center gap-2.5 text-[11px] text-slate-500 font-medium">
              <span class="flex items-center gap-1"><svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>{{ state.formatTime(item.created_at).slice(5, 16) }}</span>
              <div class="w-1 h-1 rounded-full bg-slate-300"></div>
              <span class="text-slate-600 bg-slate-50 px-1.5 py-0.5 rounded-md">{{ item.scam_type || '未知类型' }}</span>
            </div>
          </div>
          <button @click.stop="state.deleteHistoryCase(item)" class="w-8 h-8 rounded-full bg-slate-50 text-slate-400 flex items-center justify-center shrink-0 active:bg-rose-50 active:text-rose-500 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
          </button>
        </div>
      </div>
      
      <div v-if="state.history.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
        <div class="w-20 h-20 bg-white shadow-sm border border-slate-100 rounded-full flex items-center justify-center mb-5">
          <svg class="w-8 h-8 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg>
        </div>
        <p class="text-[16px] font-bold text-slate-900">暂无历史档案</p>
        <p class="text-xs text-slate-400 mt-1.5">您的检测记录将在这里安全保存</p>
      </div>
    </div>
  </div>

  <div v-if="state.activeTab === 'risk_trend'" class="bg-slate-50 pb-24">
    <!-- Header -->
    <div class="sticky top-0 z-50 bg-white/80 backdrop-blur-lg border-b border-slate-100/80 pt-safe">
      <div class="flex items-center justify-between px-4 h-14">
        <button @click="state.activeTab = 'tasks'" class="w-10 h-10 flex items-center justify-center -ml-2 active:opacity-50 transition-opacity">
          <svg class="w-6 h-6 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <h2 class="text-[17px] font-bold text-slate-900 tracking-tight">风险趋势分析</h2>
        <div class="w-10"></div>
      </div>
    </div>

    <div class="px-4 py-5 space-y-4">
      <!-- Top Stats -->
      <div class="grid grid-cols-2 gap-3">
        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60 relative overflow-hidden group">
          <div class="absolute -top-4 -right-4 w-20 h-20 bg-emerald-50 rounded-full"></div>
          <div class="relative z-10 flex flex-col h-full justify-between">
            <div class="w-8 h-8 rounded-full bg-emerald-100 flex items-center justify-center text-emerald-600 mb-4 shadow-sm border border-emerald-100/50">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
            </div>
            <div>
              <div class="text-[11px] font-bold text-emerald-600 tracking-widest uppercase mb-1">总检测次数</div>
              <div class="text-[32px] font-black leading-none text-slate-900 tracking-tight">{{ state.riskStatsSummary.total }}</div>
            </div>
          </div>
        </div>
        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60 relative overflow-hidden group">
          <div class="absolute -top-4 -right-4 w-20 h-20 bg-rose-50 rounded-full"></div>
          <div class="relative z-10 flex flex-col h-full justify-between">
            <div class="w-8 h-8 rounded-full bg-rose-100 flex items-center justify-center text-rose-500 mb-4 shadow-sm border border-rose-100/50">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
            </div>
            <div>
              <div class="text-[11px] font-bold text-rose-500 tracking-widest uppercase mb-1">高风险预警</div>
              <div class="text-[32px] font-black leading-none text-slate-900 tracking-tight">{{ state.riskStatsSummary.high }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Region Stats -->
      <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60" v-if="state.regionCaseStats && state.regionCaseStats.region">
        <div class="flex items-start justify-between mb-4">
          <div>
            <div class="text-[10px] font-bold uppercase tracking-[0.2em] text-cyan-600 mb-1 flex items-center gap-1">
              <span class="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse"></span>
              所在{{ state.regionCaseStats.region.granularity_label || '地区' }}态势
            </div>
            <div class="text-[15px] font-extrabold text-slate-900 tracking-tight">{{ ['county', 'district'].includes(state.regionCaseStats.region.granularity) ? `${state.regionCaseStats.region.province_name} / ${state.regionCaseStats.region.city_name} / ${state.regionCaseStats.region.district_name}` : `${state.regionCaseStats.region.province_name} / ${state.regionCaseStats.region.city_name}` }}</div>
          </div>
          <div class="text-[10px] font-bold text-slate-400 bg-slate-50 px-2 py-1 rounded-md border border-slate-100">总计 {{ state.regionCaseStats.summary ? state.regionCaseStats.summary.total_count : 0 }} 起</div>
        </div>
        
        <div class="grid grid-cols-3 gap-2 mb-4">
          <div class="rounded-2xl bg-slate-50 border border-slate-100/50 p-2.5 flex flex-col items-center justify-center">
            <span class="text-[10px] text-slate-400 font-bold mb-0.5">今日</span>
            <span class="text-[15px] font-black text-slate-900">{{ state.regionCaseStats.summary ? state.regionCaseStats.summary.today_count : 0 }}</span>
          </div>
          <div class="rounded-2xl bg-slate-50 border border-slate-100/50 p-2.5 flex flex-col items-center justify-center">
            <span class="text-[10px] text-slate-400 font-bold mb-0.5">近7天</span>
            <span class="text-[15px] font-black text-slate-900">{{ state.regionCaseStats.summary ? state.regionCaseStats.summary.last_7d_count : 0 }}</span>
          </div>
          <div class="rounded-2xl bg-slate-50 border border-slate-100/50 p-2.5 flex flex-col items-center justify-center">
            <span class="text-[10px] text-slate-400 font-bold mb-0.5">近30天</span>
            <span class="text-[15px] font-black text-slate-900">{{ state.regionCaseStats.summary ? state.regionCaseStats.summary.last_30d_count : 0 }}</span>
          </div>
        </div>

        <div v-if="state.regionCaseStats.top_scam_types && state.regionCaseStats.top_scam_types.length">
          <div class="text-[10px] font-bold text-slate-400 tracking-[0.2em] uppercase mb-2">高发骗局 Top5</div>
          <div class="space-y-1.5">
            <div v-for="(item, idx) in state.regionCaseStats.top_scam_types" :key="`region-scam-${idx}-${item.scam_type}`" class="flex items-center justify-between text-[11px] font-bold bg-white border border-slate-100 rounded-xl px-3 py-2 shadow-sm">
              <div class="flex items-center gap-2">
                <span class="w-4 h-4 rounded-full bg-cyan-50 text-cyan-600 flex items-center justify-center text-[9px]">{{ idx + 1 }}</span>
                <span class="text-slate-700">{{ item.scam_type }}</span>
              </div>
              <span class="text-cyan-600 bg-cyan-50 px-1.5 rounded-md">{{ item.count }}</span>
            </div>
          </div>
        </div>
        
        <div class="mt-4 rounded-2xl border px-4 py-3" :class="state.getRegionSignalClass(state.getRegionStatsSignal(state.regionCaseStats).level)">
          <div class="text-[13px] font-black tracking-tight flex items-center gap-1.5"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>{{ state.getRegionStatsSignal(state.regionCaseStats).title }}</div>
          <div class="text-[11px] mt-1.5 leading-relaxed font-medium opacity-90">{{ state.getRegionStatsSignal(state.regionCaseStats).detail }}</div>
          <div class="text-[11px] mt-1.5 font-bold pt-1.5 border-t border-current border-opacity-10">{{ state.getRegionTopScamHint(state.regionCaseStats) }}</div>
        </div>
      </div>

      <!-- Trend Window -->
      <div>
        <div class="flex items-center justify-between px-1 mb-3 mt-2">
          <div class="flex items-center gap-2">
            <h3 class="text-[15px] font-extrabold text-slate-900 tracking-tight">近期检测走势</h3>
            <span class="text-[10px] font-bold text-slate-400 bg-slate-200/50 px-1.5 py-0.5 rounded-md">近7天</span>
          </div>
        </div>
        
        <div class="space-y-3">
          <div v-for="item in state.getRecentRiskTrendRows(7)" :key="`trend-row-${item.time_bucket}`" class="bg-white rounded-[20px] p-4 shadow-sm border border-slate-100/60 flex flex-col gap-3">
            <div class="flex items-center justify-between border-b border-slate-50 pb-2">
              <div class="flex items-center gap-2">
                <div class="w-1.5 h-4 bg-slate-800 rounded-full"></div>
                <span class="text-[13px] font-black text-slate-900 tracking-tight">{{ state.formatChartLabel(item.time_bucket) }}</span>
              </div>
              <div class="text-[10px] font-bold text-slate-500 bg-slate-50 px-2 py-1 rounded-md border border-slate-100">总计 {{ item.total || 0 }} 笔</div>
            </div>
            <div class="grid grid-cols-3 gap-2">
              <div class="flex flex-col p-2.5 rounded-2xl bg-rose-50/50 border border-rose-100/50">
                <span class="text-[10px] font-bold text-rose-400 uppercase tracking-widest mb-0.5">高危</span>
                <span class="text-[17px] font-black text-rose-600">{{ item.high || 0 }}</span>
              </div>
              <div class="flex flex-col p-2.5 rounded-2xl bg-amber-50/50 border border-amber-100/50">
                <span class="text-[10px] font-bold text-amber-500 uppercase tracking-widest mb-0.5">中危</span>
                <span class="text-[17px] font-black text-amber-500">{{ item.medium || 0 }}</span>
              </div>
              <div class="flex flex-col p-2.5 rounded-2xl bg-emerald-50/50 border border-emerald-100/50">
                <span class="text-[10px] font-bold text-emerald-500 uppercase tracking-widest mb-0.5">安全</span>
                <span class="text-[17px] font-black text-emerald-500">{{ item.low || 0 }}</span>
              </div>
            </div>
          </div>
          
          <div v-if="state.getRecentRiskTrendRows(7).length === 0" class="flex flex-col items-center justify-center py-16 bg-white rounded-[24px] border border-slate-100/60 shadow-sm">
            <div class="w-16 h-16 bg-slate-50 border border-slate-100 rounded-full flex items-center justify-center mb-4">
              <svg class="w-8 h-8 text-slate-300" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
            </div>
            <span class="text-[13px] font-bold text-slate-400">近期暂无检测数据记录</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { nextTick, onMounted, onUpdated } from 'vue';

defineProps({
  state: {
    type: Object,
    required: true
  }
});

const refreshLucideIcons = () => {
  nextTick(() => {
    if (window.lucide && typeof window.lucide.createIcons === 'function') {
      window.lucide.createIcons();
    }
  });
};

onMounted(refreshLucideIcons);
onUpdated(refreshLucideIcons);
</script>
