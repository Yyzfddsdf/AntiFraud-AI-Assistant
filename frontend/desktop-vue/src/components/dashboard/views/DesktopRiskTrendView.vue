<template>
  <div class="max-w-6xl mx-auto animate-fade-in">
    <div class="flex justify-between items-end mb-8">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight">风险趋势分析</h1>
        <p class="text-slate-500 mt-2">查看您的风险等级分布与历史变化趋势</p>
      </div>
      <div class="flex bg-slate-100 p-1 rounded-sm">
        <button @click="riskInterval = 'day'; fetchRiskTrend()" :class="['px-4 py-1.5 text-xs font-bold rounded-sm transition-all', riskInterval === 'day' ? 'bg-white text-brand-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']">日</button>
        <button @click="riskInterval = 'week'; fetchRiskTrend()" :class="['px-4 py-1.5 text-xs font-bold rounded-sm transition-all', riskInterval === 'week' ? 'bg-white text-brand-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']">周</button>
        <button @click="riskInterval = 'month'; fetchRiskTrend()" :class="['px-4 py-1.5 text-xs font-bold rounded-sm transition-all', riskInterval === 'month' ? 'bg-white text-brand-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']">月</button>
        <button @click="fetchRiskTrend(true)" class="px-2 py-1.5 text-slate-400 hover:text-brand-600 ml-1 transition-colors" title="刷新数据">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
        </button>
      </div>
    </div>

    <div v-if="riskData && riskData.analysis" class="bg-white p-6 rounded-sm shadow-sm border border-slate-200 mb-8">
      <div class="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
        <div>
          <h3 class="text-lg font-bold text-slate-800 mb-2 flex items-center gap-2">
            <span class="w-1.5 h-4 bg-rose-500 rounded-sm"></span>
            中文趋势判断
          </h3>
          <p class="text-sm text-slate-600 leading-7">{{ riskData.analysis.summary || '暂无趋势分析结论。' }}</p>
          <p v-if="riskData.analysis.current_bucket" class="text-xs text-slate-400 mt-2">
            对比窗口：{{ riskData.analysis.previous_bucket || '无上一窗口' }} → {{ riskData.analysis.current_bucket }}
          </p>
        </div>
        <div class="flex flex-wrap gap-3 lg:justify-end">
          <div class="px-4 py-3 rounded-sm min-w-[140px]" :class="getRiskTrendAnalysisClass(riskData.analysis.overall_trend)">
            <div class="text-xs opacity-80 mb-1">整体风险</div>
            <div class="text-base font-bold">{{ riskData.analysis.overall_trend }}</div>
          </div>
          <div class="px-4 py-3 rounded-sm min-w-[140px]" :class="getRiskTrendAnalysisClass(riskData.analysis.high_risk_trend)">
            <div class="text-xs opacity-80 mb-1">高风险案件</div>
            <div class="text-base font-bold">{{ riskData.analysis.high_risk_trend }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div class="bg-white p-6 rounded-sm shadow-sm border border-slate-200">
        <h3 class="text-lg font-bold text-slate-800 mb-6 flex items-center gap-2">
          <span class="w-1.5 h-4 bg-indigo-500 rounded-sm"></span>
          风险等级分布
        </h3>
        <div class="relative h-64 w-full flex items-center justify-center">
          <div id="riskPieChart" class="w-full h-full"></div>
        </div>
      </div>

      <div class="bg-white p-6 rounded-sm shadow-sm border border-slate-200">
        <h3 class="text-lg font-bold text-slate-800 mb-6 flex items-center gap-2">
          <span class="w-1.5 h-4 bg-brand-500 rounded-sm"></span>
          风险变化趋势
        </h3>
        <div class="relative h-64 w-full">
          <div id="riskLineChart" class="w-full h-full"></div>
        </div>
      </div>
    </div>

    <div v-if="regionCaseStatsNeedsSetup" class="mt-8 bg-amber-50 border border-amber-200 rounded-sm px-6 py-5 flex items-start justify-between gap-4">
      <div>
        <div class="text-sm font-bold text-amber-800">尚未设置所在地区</div>
        <div class="text-sm text-amber-700 mt-1">{{ regionCaseStatsMessage || '请先在“个人资料”中完善省/市/县（区）信息，才能查看所在地区案件统计。' }}</div>
      </div>
      <button @click="activeTab = 'profile'" class="shrink-0 px-4 py-2 text-xs font-bold rounded-sm bg-amber-600 text-white hover:bg-amber-700 transition-colors">去设置</button>
    </div>

    <div v-else-if="regionCaseStats && regionCaseStats.region" class="mt-8 bg-white p-6 rounded-sm shadow-sm border border-cyan-100">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h3 class="text-lg font-bold text-slate-800 mb-1 flex items-center gap-2">
            <span class="w-1.5 h-4 bg-cyan-500 rounded-sm"></span>
            所在{{ regionCaseStats.region.granularity_label || '市' }}案件统计
          </h3>
          <p class="text-sm text-slate-500">
            {{ ['county', 'district'].includes(regionCaseStats.region.granularity) ? `${regionCaseStats.region.province_name} / ${regionCaseStats.region.city_name} / ${regionCaseStats.region.district_name}` : `${regionCaseStats.region.province_name} / ${regionCaseStats.region.city_name}` }}
          </p>
        </div>
        <span class="text-xs bg-cyan-50 text-cyan-700 border border-cyan-100 px-3 py-1.5 rounded-sm font-medium">总计 {{ regionCaseStats.summary ? regionCaseStats.summary.total_count : 0 }}</span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-5">
        <div class="rounded-sm border border-slate-100 bg-slate-50 px-4 py-3">
          <div class="text-xs text-slate-400">今日</div>
          <div class="text-2xl font-black text-slate-900 mt-1">{{ regionCaseStats.summary ? regionCaseStats.summary.today_count : 0 }}</div>
        </div>
        <div class="rounded-sm border border-slate-100 bg-slate-50 px-4 py-3">
          <div class="text-xs text-slate-400">近7天</div>
          <div class="text-2xl font-black text-slate-900 mt-1">{{ regionCaseStats.summary ? regionCaseStats.summary.last_7d_count : 0 }}</div>
        </div>
        <div class="rounded-sm border border-slate-100 bg-slate-50 px-4 py-3">
          <div class="text-xs text-slate-400">近30天</div>
          <div class="text-2xl font-black text-slate-900 mt-1">{{ regionCaseStats.summary ? regionCaseStats.summary.last_30d_count : 0 }}</div>
        </div>
      </div>

      <div v-if="regionCaseStats.top_scam_types && regionCaseStats.top_scam_types.length" class="mt-6">
        <div class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-3">高发骗局 Top5</div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
          <div v-for="(item, idx) in regionCaseStats.top_scam_types" :key="`desktop-region-scam-${idx}-${item.scam_type}`" class="flex items-center justify-between bg-cyan-50/60 border border-cyan-100 rounded-sm px-3 py-2">
            <span class="text-sm font-medium text-slate-700">{{ item.scam_type }}</span>
            <span class="text-sm font-bold text-cyan-700">{{ item.count }}</span>
          </div>
        </div>
      </div>

      <div class="mt-5 rounded-sm border px-4 py-3" :class="getRegionSignalClass(getRegionStatsSignal(regionCaseStats).level)">
        <div class="text-sm font-bold">{{ getRegionStatsSignal(regionCaseStats).title }}</div>
        <div class="text-sm mt-1 opacity-90">{{ getRegionStatsSignal(regionCaseStats).detail }}</div>
        <div class="text-xs mt-2 font-medium">{{ getRegionTopScamHint(regionCaseStats) }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'DesktopRiskTrendView',
  props: {
    app: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    return props.app;
  }
};
</script>
