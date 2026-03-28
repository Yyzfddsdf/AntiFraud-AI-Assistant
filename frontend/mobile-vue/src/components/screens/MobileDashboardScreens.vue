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
        <div class="text-[22px] font-bold text-slate-900 leading-none">{{ state.todayRiskStatsSummary.total }}</div>
        <div class="text-[11px] text-slate-400 mt-1">今日检测数量</div>
      </div>

      <div class="px-[16px] py-[16px] bg-white border border-slate-100 rounded-[24px] shadow-sm">
        <div class="w-7 h-7 bg-rose-50 text-rose-600 rounded-lg flex items-center justify-center mb-2">
          <i data-lucide="alert-octagon" size="14"></i>
        </div>
        <div class="text-[22px] font-bold text-slate-900 leading-none">{{ state.todayRiskStatsSummary.high }}</div>
        <div class="text-[11px] text-slate-400 mt-1">今日高危预警</div>
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
        <h4 class="text-[16px] font-bold mb-2">{{ homepageTrendHeadline }}</h4>
        <p class="text-[11px] text-slate-400 leading-relaxed mb-4">{{ homepageTrendSummary }}</p>
        
        <div class="grid grid-cols-2 gap-4 pt-4 border-t border-white/5">
          <div>
            <p class="text-[9px] text-slate-500 uppercase font-bold mb-0.5">整体走势</p>
            <p class="text-[12px] font-semibold" :class="homepageOverallTrendClass">{{ homepageOverallTrendLabel }}</p>
          </div>
          <div>
            <p class="text-[9px] text-slate-500 uppercase font-bold mb-0.5">当前风险</p>
            <p class="text-[12px] font-semibold" :class="homepageCurrentRiskClass">{{ homepageCurrentRiskLabel }}</p>
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

  <div v-if="state.activeTab === 'history'" class="bg-slate-50 pb-8">
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

  <div v-if="state.activeTab === 'risk_trend'" class="risk-trend-screen min-h-full pb-0">
    <div class="sticky top-0 z-50 bg-white border-b border-slate-100 pt-safe">
      <div class="flex items-center justify-between px-4 h-14">
        <button @click="state.activeTab = 'tasks'" class="w-10 h-10 flex items-center justify-center -ml-2 active:opacity-50 transition-opacity">
          <svg class="w-6 h-6 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <h2 class="text-[17px] font-bold text-slate-900 tracking-tight" style="font-family: Outfit, 'Plus Jakarta Sans', sans-serif;">风险趋势分析</h2>
        <div class="w-10"></div>
      </div>
    </div>

    <div class="relative px-3.5 pt-4 pb-0 space-y-4 overflow-hidden min-h-full">
      <section class="risk-hero-card" :class="heroToneClass">
        <div class="flex items-start gap-3">
          <div class="risk-hero-icon">
            <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" stroke-width="2.1" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 3l7 4v5c0 4.7-2.9 7.8-7 9-4.1-1.2-7-4.3-7-9V7l7-4z"></path><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v4"></path><path stroke-linecap="round" stroke-linejoin="round" d="M12 16h.01"></path></svg>
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-[18px] font-black text-slate-900 tracking-tight leading-tight" style="font-family: Outfit, 'Plus Jakarta Sans', sans-serif;">{{ riskHeroTitle }}</div>
            <p class="mt-1.5 text-[12px] leading-6 text-slate-600">{{ riskHeroDetail }}</p>
          </div>
        </div>
        <div class="risk-hero-footer">
          <span class="text-slate-400">最近高发:</span>
          <span class="text-slate-700">{{ riskHeroFooter }}</span>
        </div>
      </section>

      <section class="grid grid-cols-2 gap-3">
        <article class="risk-stat-card">
          <div class="risk-stat-icon risk-stat-icon--mint">
            <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13v6m7-10v10m7-14v14M3 19h18"></path></svg>
          </div>
          <div class="text-[12px] font-bold text-slate-400 mt-4">总检测次数</div>
          <div class="flex items-end gap-1.5 mt-2">
            <span class="text-[27px] font-black text-slate-950 leading-none">{{ state.riskStatsSummary.total }}</span>
            <span class="text-[11px] font-bold leading-none mb-1" :class="summaryDelta.total === '--' ? 'text-slate-400' : 'text-emerald-500'">{{ summaryDelta.total }}</span>
          </div>
        </article>

        <article class="risk-stat-card">
          <div class="risk-stat-icon risk-stat-icon--rose">
            <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3"></path><path stroke-linecap="round" stroke-linejoin="round" d="M12 16h.01"></path><path stroke-linecap="round" stroke-linejoin="round" d="M10.29 3.86l-7.2 12.47A2 2 0 004.82 19h14.36a2 2 0 001.73-2.67l-7.2-12.47a2 2 0 00-3.42 0z"></path></svg>
          </div>
          <div class="text-[12px] font-bold text-slate-400 mt-4">高风险预警</div>
          <div class="flex items-end gap-1.5 mt-2">
            <span class="text-[27px] font-black text-slate-950 leading-none">{{ state.riskStatsSummary.high }}</span>
            <span class="text-[11px] font-bold leading-none mb-1" :class="summaryDelta.high === '--' ? 'text-slate-400' : 'text-rose-500'">{{ summaryDelta.high }}</span>
          </div>
        </article>
      </section>

      <section class="risk-section-card">
        <div class="flex items-center justify-between">
          <h3 class="risk-section-title">所在区态势</h3>
          <span class="risk-pill-badge">总计 {{ regionSummaryTotal }} 起</span>
        </div>

        <div class="risk-location-card">
          <div class="risk-location-icon">
            <svg class="w-[17px] h-[17px]" fill="none" stroke="currentColor" stroke-width="2.1" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 21s6-5.33 6-11a6 6 0 10-12 0c0 5.67 6 11 6 11z"></path><circle cx="12" cy="10" r="2.5"></circle></svg>
          </div>
          <div class="text-[15px] font-extrabold text-slate-900 tracking-tight leading-snug">{{ regionPathLabel }}</div>
        </div>

        <div class="grid grid-cols-3 gap-3 mt-4">
          <button
            v-for="item in regionWindowCards"
            :key="item.key"
            type="button"
            class="risk-window-chip"
            :class="{ 'is-active': selectedRegionWindow === item.key }"
            @click="selectedRegionWindow = item.key"
          >
            <span class="text-[11px] font-bold text-slate-400">{{ item.label }}</span>
            <span class="text-[29px] font-black text-slate-950 leading-none mt-2">{{ item.value }}</span>
          </button>
        </div>

        <div class="risk-window-note">
          <span class="risk-window-note__dot"></span>
          <span>{{ activeRegionWindowNote }}</span>
        </div>

        <div v-if="topRegionScamTypes.length" class="mt-5">
          <div class="text-[12px] font-black text-slate-400 tracking-tight mb-3">高发骗局 TOP 5</div>
          <div class="space-y-3">
            <article
              v-for="(item, idx) in topRegionScamTypes"
              :key="`region-scam-${idx}-${item.scam_type}`"
              class="risk-top-scam-row"
              :style="{ animationDelay: `${idx * 90}ms` }"
            >
              <div class="flex items-center gap-3 min-w-0">
                <div class="risk-top-scam-rank">{{ idx + 1 }}</div>
                <span class="text-[15px] font-extrabold text-slate-900 truncate">{{ item.scam_type }}</span>
              </div>
              <span class="text-[18px] font-black text-indigo-600 leading-none">{{ item.count }}</span>
            </article>
          </div>
        </div>
      </section>

      <section>
        <div class="flex items-center justify-between px-1 mb-3">
          <h3 class="text-[22px] font-black text-slate-950 tracking-tight" style="font-family: Outfit, 'Plus Jakarta Sans', sans-serif;">近期检测走势</h3>
          <span class="risk-chip-tag">近7天</span>
        </div>

        <div v-if="recentTrendCards.length" class="space-y-3">
          <article
            v-for="(item, cardIndex) in recentTrendCards"
            :key="`trend-row-${item.time_bucket}`"
            class="risk-trend-card"
            :class="{ 'is-featured': cardIndex === 0 }"
            :style="{ animationDelay: `${cardIndex * 120}ms` }"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="risk-trend-dot" :class="item.dotToneClass"></span>
                  <span class="text-[17px] font-black text-slate-950 tracking-tight">{{ item.label }}</span>
                </div>
                <div class="text-[12px] text-slate-500 mt-1">{{ item.headline }}</div>
              </div>
              <span class="risk-pill-badge shrink-0">总计 {{ item.total }} 笔</span>
            </div>

            <div class="risk-trend-divider"></div>

            <div class="grid grid-cols-3 gap-2 mt-4">
              <div class="risk-metric-chip risk-metric-chip--rose">
                <span class="text-[10px] font-bold tracking-wide">高危</span>
                <span class="text-[18px] font-black">{{ item.high }}</span>
              </div>
              <div class="risk-metric-chip risk-metric-chip--amber">
                <span class="text-[10px] font-bold tracking-wide">中危</span>
                <span class="text-[18px] font-black">{{ item.medium }}</span>
              </div>
              <div class="risk-metric-chip risk-metric-chip--mint">
                <span class="text-[10px] font-bold tracking-wide">安全</span>
                <span class="text-[18px] font-black">{{ item.low }}</span>
              </div>
            </div>
          </article>
        </div>

        <div v-else class="risk-empty-card">
          <div class="w-14 h-14 rounded-full bg-slate-100 flex items-center justify-center text-slate-300 mb-4">
            <svg class="w-7 h-7" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M12 5v14"></path></svg>
          </div>
          <span class="text-[13px] font-bold text-slate-400">近期暂无检测数据记录</span>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUpdated, ref } from 'vue';

const props = defineProps({
  state: {
    type: Object,
    required: true
  }
});

const state = props.state;
const selectedRegionWindow = ref('week');

const homepageTrendAnalysis = computed(() => state.riskData?.analysis || null);

const homepageTrendHeadline = computed(() => {
  if (homepageTrendAnalysis.value && typeof state.getRiskTrendHeadline === 'function') {
    return state.getRiskTrendHeadline(homepageTrendAnalysis.value);
  }
  return 'AI 研判：风险走势仍需持续观察';
});

const homepageTrendSummary = computed(() => {
  const summary = String(homepageTrendAnalysis.value?.summary || '').trim();
  return summary || '后端趋势分析暂未返回摘要，请继续关注最新检测结果。';
});

const homepageOverallTrendLabel = computed(() => {
  const overall = String(homepageTrendAnalysis.value?.overall_trend || '').trim();
  if (typeof state.formatRiskTrendDescriptor === 'function') {
    return state.formatRiskTrendDescriptor(overall);
  }
  return '风险热度待观察';
});

const homepageOverallTrendClass = computed(() => {
  const overall = String(homepageTrendAnalysis.value?.overall_trend || '').trim();
  if (overall === '上升') return 'text-rose-400';
  if (overall === '下降') return 'text-emerald-400';
  if (overall === '平稳') return 'text-amber-300';
  return 'text-slate-300';
});

const homepageCurrentRisk = computed(() => {
  const stats = state.riskData?.stats || {};
  const total = Number(stats.total) || 0;
  const high = Number(stats.high) || 0;
  const medium = Number(stats.medium) || 0;
  if (total === 0) return '待观察';
  if (high > 0 || high / Math.max(total, 1) >= 0.2) return '高风险';
  if (medium > 0) return '中风险';
  return '低风险';
});

const homepageCurrentRiskLabel = computed(() => homepageCurrentRisk.value);

const homepageCurrentRiskClass = computed(() => {
  if (homepageCurrentRisk.value === '高风险') return 'text-rose-300';
  if (homepageCurrentRisk.value === '中风险') return 'text-amber-300';
  if (homepageCurrentRisk.value === '低风险') return 'text-white';
  return 'text-slate-300';
});

const fallbackRegionSignal = {
  level: 'neutral',
  title: '近期风险相对平稳',
  detail: '仍需保持警惕，遇到催转账和索要验证码请先核实。'
};

const regionSignal = computed(() => {
  if (state.regionCaseStats && typeof state.getRegionStatsSignal === 'function') {
    return state.getRegionStatsSignal(state.regionCaseStats);
  }
  return fallbackRegionSignal;
});

const regionPathLabel = computed(() => {
  const region = state.regionCaseStats?.region;
  if (!region) return '浙江省 / 杭州市 / 钱塘区';
  if (['county', 'district'].includes(region.granularity)) {
    return `${region.province_name} / ${region.city_name} / ${region.district_name}`;
  }
  return `${region.province_name} / ${region.city_name}`;
});

const regionSummaryTotal = computed(() => Number(state.regionCaseStats?.summary?.total_count) || 0);

const topRegionScamTypes = computed(() => {
  const topList = state.regionCaseStats?.top_scam_types;
  return Array.isArray(topList) ? topList.slice(0, 5) : [];
});

const riskHeroTitle = computed(() => {
  if (homepageTrendAnalysis.value) {
    return homepageTrendHeadline.value;
  }
  return regionSignal.value.title || fallbackRegionSignal.title;
});

const riskHeroDetail = computed(() => {
  if (homepageTrendAnalysis.value) {
    return homepageTrendSummary.value;
  }
  return regionSignal.value.detail || fallbackRegionSignal.detail;
});

const riskHeroFooter = computed(() => {
  if (state.regionCaseStats && typeof state.getRegionTopScamHint === 'function') {
    return state.getRegionTopScamHint(state.regionCaseStats).replace(/^最近高发[:：]\s*/, '');
  }
  if (topRegionScamTypes.value.length) {
    const topItem = topRegionScamTypes.value[0];
    return `${topItem.scam_type}（${topItem.count}起）`;
  }
  return '暂未形成明显高发类型';
});

const heroToneClass = computed(() => {
  const overallTrend = String(homepageTrendAnalysis.value?.overall_trend || '').trim();
  if (overallTrend === '上升') {
    return {
      'is-high': true,
      'is-medium': false,
      'is-low': false,
      'is-neutral': false
    };
  }
  if (overallTrend === '下降') {
    return {
      'is-high': false,
      'is-medium': false,
      'is-low': true,
      'is-neutral': false
    };
  }
  if (overallTrend === '平稳') {
    return {
      'is-high': false,
      'is-medium': true,
      'is-low': false,
      'is-neutral': false
    };
  }
  return {
    'is-high': regionSignal.value.level === 'high',
    'is-medium': regionSignal.value.level === 'medium',
    'is-low': regionSignal.value.level === 'low',
    'is-neutral': !['high', 'medium', 'low'].includes(regionSignal.value.level)
  };
});

const regionWindowCards = computed(() => {
  const summary = state.regionCaseStats?.summary || {};
  return [
    { key: 'day', label: '今日', value: Number(summary.today_count) || 0, note: '今日新增态势，用于观察短时抬升。' },
    { key: 'week', label: '近7天', value: Number(summary.last_7d_count) || 0, note: '近7天波动更能反映当前风险热度。' },
    { key: 'month', label: '近30天', value: Number(summary.last_30d_count) || 0, note: '近30天累计样本适合判断长期变化。' }
  ];
});

const activeRegionWindowNote = computed(() => {
  const selectedCard = regionWindowCards.value.find((item) => item.key === selectedRegionWindow.value);
  return selectedCard ? selectedCard.note : '暂无地区统计说明。';
});

const recentTrendRows = computed(() => {
  if (typeof state.getRecentRiskTrendRows !== 'function') return [];
  const rows = state.getRecentRiskTrendRows(7);
  return Array.isArray(rows) ? rows : [];
});

const buildDeltaText = (current, previous) => {
  const currentValue = Number(current) || 0;
  const previousValue = Number(previous) || 0;
  if (!previousValue) return currentValue > 0 ? '+100%' : '--';
  const delta = Math.round(((currentValue - previousValue) / previousValue) * 100);
  if (!Number.isFinite(delta) || delta === 0) return '--';
  return `${delta > 0 ? '+' : ''}${delta}%`;
};

const summaryDelta = computed(() => {
  const latest = recentTrendRows.value[0] || {};
  const previous = recentTrendRows.value[1] || {};
  return {
    total: buildDeltaText(latest.total, previous.total),
    high: buildDeltaText(latest.high, previous.high)
  };
});

const getTrendTone = (item) => {
  const high = Number(item?.high) || 0;
  const medium = Number(item?.medium) || 0;
  if (high > 0) return 'trend-dot--rose';
  if (medium > 0) return 'trend-dot--amber';
  return 'trend-dot--indigo';
};

const buildTrendHeadline = (item) => {
  const total = Number(item?.total) || 0;
  const high = Number(item?.high) || 0;
  const medium = Number(item?.medium) || 0;
  const low = Number(item?.low) || 0;
  if (total === 0) return '当日暂无新增检测记录';
  if (high > 0) return `${high} 笔高危信号，建议优先核验来源`;
  if (medium > 0 && medium >= low) return '中风险样本占比更高，注意退款客服与陌生链接';
  return '以常规预警样本为主，整体波动较稳';
};

const recentTrendCards = computed(() => {
  const rows = recentTrendRows.value.slice(0, 3);
  return rows.map((item, cardIndex) => {
    return {
      time_bucket: item.time_bucket,
      label: state.formatChartLabel(item.time_bucket),
      total: Number(item.total) || 0,
      high: Number(item.high) || 0,
      medium: Number(item.medium) || 0,
      low: Number(item.low) || 0,
      headline: buildTrendHeadline(item),
      dotToneClass: getTrendTone(item),
      animationDelay: `${cardIndex * 120}ms`
    };
  });
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

<style scoped>
.risk-trend-screen {
  background: #f8fafc;
}

.risk-hero-card {
  position: relative;
  overflow: hidden;
  padding: 18px 18px 16px;
  border-radius: 26px;
  border: 1px solid rgba(226, 232, 240, 0.48);
  background: rgba(255, 255, 255, 0.84);
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.04);
  backdrop-filter: blur(12px);
  animation: riskCardEnter 0.55s ease both;
}

.risk-hero-card.is-high {
  border-color: rgba(226, 232, 240, 0.48);
  background: #ffffff;
}

.risk-hero-card.is-medium {
  border-color: rgba(226, 232, 240, 0.48);
  background: #ffffff;
}

.risk-hero-card.is-low,
.risk-hero-card.is-neutral {
  border-color: rgba(226, 232, 240, 0.48);
  background: #ffffff;
}

.risk-hero-icon {
  width: 38px;
  height: 38px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #10b981;
  background: linear-gradient(180deg, #e8fff7 0%, #d8faee 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.risk-hero-footer {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  font-size: 12px;
  font-weight: 800;
  display: flex;
  align-items: center;
  gap: 8px;
}

.risk-stat-card,
.risk-section-card,
.risk-trend-card,
.risk-empty-card {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(226, 232, 240, 0.46);
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.04);
  backdrop-filter: blur(12px);
}

.risk-stat-card {
  min-height: 152px;
  padding: 18px;
  animation: riskCardEnter 0.55s ease both;
}

.risk-stat-icon {
  position: relative;
  width: 42px;
  height: 42px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.86);
}

.risk-stat-icon--mint {
  color: #14b8a6;
  background: linear-gradient(180deg, #ecfdf5 0%, #dcfce7 100%);
}

.risk-stat-icon--rose {
  color: #fb7185;
  background: linear-gradient(180deg, #fff1f2 0%, #ffe4e6 100%);
}

.risk-section-card {
  padding: 18px;
  animation: riskCardEnter 0.65s ease both;
}

.risk-section-title {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 17px;
  font-weight: 900;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.risk-section-title::before {
  content: '';
  width: 5px;
  height: 24px;
  border-radius: 999px;
  background: linear-gradient(180deg, #4f46e5 0%, #7367ff 100%);
}

.risk-pill-badge,
.risk-chip-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 800;
}

.risk-pill-badge {
  color: #94a3b8;
  background: #f1f5f9;
}

.risk-chip-tag {
  color: #5b4cf0;
  background: rgba(99, 102, 241, 0.1);
}

.risk-location-card {
  margin-top: 14px;
  padding: 16px 14px;
  border-radius: 20px;
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.96) 0%, rgba(255, 255, 255, 0.96) 100%);
  border: 1px solid rgba(226, 232, 240, 0.42);
  display: flex;
  align-items: center;
  gap: 12px;
}

.risk-location-icon {
  width: 34px;
  height: 34px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4f46e5;
  background: rgba(99, 102, 241, 0.1);
}

.risk-window-chip {
  border-radius: 18px;
  padding: 14px 12px;
  text-align: center;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  border: 1px solid rgba(226, 232, 240, 0.38);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
  transition: transform 0.25s ease, box-shadow 0.25s ease, background 0.25s ease, border-color 0.25s ease;
}

.risk-window-chip.is-active {
  background: linear-gradient(180deg, #5e4ef7 0%, #4f46e5 100%);
  border-color: rgba(99, 102, 241, 0.5);
  box-shadow: 0 14px 24px rgba(79, 70, 229, 0.24);
  transform: translateY(-2px);
}

.risk-window-chip.is-active span {
  color: #ffffff;
}

.risk-window-note {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #64748b;
  font-size: 12px;
  font-weight: 700;
}

.risk-window-note__dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #4f46e5;
  box-shadow: 0 0 0 6px rgba(99, 102, 241, 0.08);
  animation: riskPulseDot 2.1s ease-in-out infinite;
}

.risk-top-scam-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 250, 252, 0.94) 100%);
  border: 1px solid rgba(226, 232, 240, 0.38);
  box-shadow: 0 10px 20px rgba(15, 23, 42, 0.028);
  animation: riskCardEnter 0.6s ease both;
}

.risk-top-scam-rank {
  width: 24px;
  height: 24px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  font-size: 11px;
  font-weight: 900;
  background: #eef2ff;
}

.risk-trend-card {
  padding: 16px;
  animation: riskCardEnter 0.7s ease both;
}

.risk-trend-dot {
  width: 9px;
  height: 9px;
  border-radius: 999px;
  box-shadow: 0 0 0 6px rgba(99, 102, 241, 0.08);
}

.trend-dot--indigo {
  background: #4f46e5;
}

.trend-dot--rose {
  background: #f43f5e;
  box-shadow: 0 0 0 6px rgba(244, 63, 94, 0.08);
}

.trend-dot--amber {
  background: #f59e0b;
  box-shadow: 0 0 0 6px rgba(245, 158, 11, 0.08);
}

.risk-trend-divider {
  margin-top: 16px;
  height: 1px;
  background: linear-gradient(90deg, rgba(226, 232, 240, 0.2) 0%, rgba(226, 232, 240, 1) 28%, rgba(226, 232, 240, 1) 72%, rgba(226, 232, 240, 0.2) 100%);
}

.risk-metric-chip {
  border-radius: 16px;
  padding: 11px 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.risk-metric-chip--rose {
  color: #e11d48;
  background: rgba(255, 241, 242, 0.92);
}

.risk-metric-chip--amber {
  color: #d97706;
  background: rgba(255, 251, 235, 0.96);
}

.risk-metric-chip--mint {
  color: #059669;
  background: rgba(236, 253, 245, 0.96);
}

.risk-empty-card {
  padding: 44px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

@keyframes riskCardEnter {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes riskPulseDot {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.12);
  }
}

@media (prefers-reduced-motion: reduce) {
  .risk-window-note__dot,
  .risk-hero-card,
  .risk-stat-card,
  .risk-section-card,
  .risk-top-scam-row,
  .risk-trend-card {
    animation: none !important;
  }

  .risk-window-chip {
    transition: none !important;
  }
}
</style>
