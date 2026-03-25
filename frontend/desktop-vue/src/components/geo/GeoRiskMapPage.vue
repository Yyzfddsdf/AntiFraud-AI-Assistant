<template>
  <div class="h-dvh bg-[radial-gradient(circle_at_top_left,_rgba(6,182,212,0.14),_transparent_25%),radial-gradient(circle_at_top_right,_rgba(59,130,246,0.12),_transparent_20%),linear-gradient(180deg,_#020617,_#0f172a_38%,_#020617)] text-white overflow-hidden flex flex-col">
    <div class="absolute inset-0 pointer-events-none opacity-50" style="background-image: linear-gradient(rgba(148,163,184,0.08) 1px, transparent 1px), linear-gradient(90deg, rgba(148,163,184,0.08) 1px, transparent 1px); background-size: 34px 34px;"></div>
    <div class="relative z-10 flex-1 min-h-0 px-4 py-4 md:px-6 md:py-5 overflow-hidden">
      <div class="flex h-full min-h-0 flex-col gap-4">
        <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div class="flex items-start gap-3">
            <button @click="activeTab = 'admin_stats'" class="inline-flex items-center gap-2 rounded-sm border border-white/10 bg-white/5 px-3 py-2 text-xs font-bold text-slate-100 hover:bg-white/10 transition-colors shrink-0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
              返回全景分析
            </button>
            <div>
              <div class="text-[11px] font-bold uppercase tracking-[0.3em] text-cyan-200/90">Geo Risk Theater</div>
              <h1 class="mt-1.5 text-2xl md:text-3xl font-black tracking-tight">全国反诈案件地理态势</h1>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <button
              v-for="item in geoWindowOptions"
              :key="`geo-page-window-${item.value}`"
              @click="setGeoWindow(item.value)"
              :class="[
                'px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide transition-all',
                geoSelectedWindow === item.value
                  ? 'bg-cyan-400 text-slate-950 shadow-sm'
                  : 'bg-white/5 text-slate-300 hover:text-white hover:bg-white/10'
              ]">
              {{ item.label }}
            </button>
            <button @click="fetchGeoRiskMap(true)" class="px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide bg-white/5 text-slate-300 hover:text-white hover:bg-white/10 transition-all">
              刷新
            </button>
          </div>
        </div>

        <div class="grid flex-1 min-h-0 grid-cols-[minmax(0,1.85fr)_360px] gap-4">
          <section class="flex min-h-0 flex-col rounded-sm border border-cyan-400/15 bg-slate-950/55 backdrop-blur-none">
            <div class="flex items-center justify-between gap-3 px-4 py-4 border-b border-white/10">
              <div>
                <div class="text-[11px] font-bold uppercase tracking-[0.26em] text-cyan-200/80">实时地图</div>
                <h2 class="mt-1.5 text-lg font-extrabold text-white">{{ geoMapTitle }}</h2>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  @click="setGeoViewMode('province')"
                  :class="[
                    'px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide transition-all',
                    geoViewMode === 'province'
                      ? 'bg-indigo-400 text-slate-950'
                      : 'bg-white/5 text-slate-300 hover:bg-white/10'
                  ]">
                  省级总览
                </button>
                <button
                  @click="setGeoViewMode('city')"
                  :disabled="!geoSelectedProvinceCode"
                  :class="[
                    'px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide transition-all',
                    geoViewMode === 'city'
                      ? 'bg-fuchsia-400 text-slate-950'
                      : 'bg-white/5 text-slate-300 hover:bg-white/10 disabled:opacity-40 disabled:cursor-not-allowed'
                  ]">
                  城市钻取
                </button>
                <button
                  @click="setGeoViewMode('district')"
                  :disabled="!geoSelectedCityCode"
                  :class="[
                    'px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide transition-all',
                    geoViewMode === 'district'
                      ? 'bg-emerald-400 text-slate-950'
                      : 'bg-white/5 text-slate-300 hover:bg-white/10 disabled:opacity-40 disabled:cursor-not-allowed'
                  ]">
                  县区钻取
                </button>
                <button v-if="geoViewMode === 'district'" @click="backToCityGeoMap" class="px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide bg-white/5 text-slate-300 hover:bg-white/10 transition-all">
                  返回城市
                </button>
                <button v-if="geoViewMode === 'city'" @click="backToProvinceGeoMap" class="px-3 py-1.5 rounded-sm text-[11px] font-bold tracking-wide bg-white/5 text-slate-300 hover:bg-white/10 transition-all">
                  返回全国
                </button>
              </div>
            </div>
            <div class="flex-1 min-h-0 p-4">
              <div class="relative h-full min-h-[360px] rounded-sm border border-white/6 bg-[radial-gradient(circle_at_top,_rgba(56,189,248,0.08),_transparent_38%),linear-gradient(180deg,_rgba(2,6,23,0.96),_rgba(15,23,42,0.92))] overflow-hidden">
                <div id="adminGeoRiskMapChart" class="absolute inset-0"></div>
                <div v-if="geoMapLoading" class="absolute inset-0 flex flex-col items-center justify-center gap-4 bg-slate-950/85 text-cyan-100">
                  <div class="w-14 h-14 rounded-sm border-2 border-cyan-300/20 border-t-cyan-300 animate-spin"></div>
                  <div class="text-sm font-bold tracking-[0.24em] uppercase">地图数据加载中</div>
                </div>
                <div v-else-if="geoMapError" class="absolute inset-0 flex flex-col items-center justify-center gap-4 bg-slate-950/85 text-center px-8">
                  <div class="text-base font-extrabold text-rose-200">全国地理态势加载失败</div>
                  <div class="text-sm text-slate-300 max-w-md">{{ geoMapError }}</div>
                  <button @click="fetchGeoRiskMap(true)" class="px-4 py-2 rounded-sm bg-rose-500/15 border border-rose-300/20 text-rose-100 text-sm font-bold hover:bg-rose-500/25 transition-colors">
                    重新加载
                  </button>
                </div>
              </div>
            </div>
          </section>

          <aside class="flex min-h-0 flex-col gap-4 overflow-hidden">
            <div class="grid grid-cols-2 gap-3">
              <div class="rounded-sm border border-white/10 bg-white/5 p-4 backdrop-blur-none">
                <div class="text-[11px] font-bold uppercase tracking-[0.24em] text-slate-400">定位用户</div>
                <div class="mt-2 text-2xl font-black text-white">{{ geoMapData?.summary?.total_users_with_location || 0 }}</div>
              </div>
              <div class="rounded-sm border border-white/10 bg-white/5 p-4 backdrop-blur-none">
                <div class="text-[11px] font-bold uppercase tracking-[0.24em] text-slate-400">历史案件</div>
                <div class="mt-2 text-2xl font-black text-white">{{ geoMapData?.summary?.total_cases || 0 }}</div>
              </div>
            </div>

            <div class="flex-1 min-h-0 rounded-sm border border-white/10 bg-white/5 p-4 backdrop-blur-none overflow-hidden">
              <div class="text-[11px] font-bold uppercase tracking-[0.24em] text-slate-400">{{ geoRankingTitle }}</div>
              <div class="mt-3 space-y-2.5 h-[calc(100%-1.5rem)] overflow-y-auto pr-1">
                <button
                  v-for="(item, index) in geoCurrentRanking"
                  :key="`geo-page-rank-${item.region_code}`"
                  @click="geoViewMode === 'province' ? drillIntoProvince(item.region_code) : (geoViewMode === 'city' ? drillIntoCity(item.region_code) : null)"
                  class="w-full rounded-sm border border-white/8 bg-slate-950/55 px-3.5 py-3 text-left hover:border-cyan-300/25 transition-all">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <div class="text-[11px] font-bold uppercase tracking-[0.2em] text-slate-500">TOP {{ index + 1 }}</div>
                      <div class="mt-1.5 truncate text-sm font-extrabold text-white">{{ item.region_name }}</div>
                      <div class="mt-2 flex flex-wrap gap-1.5">
                        <span
                          v-for="scam in item.stats[geoSelectedWindow].top_scam_types"
                          :key="`${item.region_code}-${scam.scam_type}`"
                          class="rounded-sm border border-white/10 bg-white/5 px-2 py-0.5 text-[10px] font-bold text-cyan-100">
                          {{ scam.scam_type }} · {{ scam.count }}
                        </span>
                      </div>
                    </div>
                    <div class="shrink-0 text-right">
                      <div class="text-xl font-black text-white">{{ item.stats[geoSelectedWindow].count }}</div>
                      <div :class="geoRiskBadgeClass(item.stats[geoSelectedWindow].risk_level)" class="mt-1.5 inline-flex rounded-sm px-2 py-0.5 text-[10px] font-bold">
                        {{ item.stats[geoSelectedWindow].risk_level }}风险
                      </div>
                      <div class="mt-1.5 text-[10px] text-slate-400">{{ item.stats[geoSelectedWindow].trend }} · {{ formatGeoChange(item.stats[geoSelectedWindow].change_rate) }}</div>
                    </div>
                  </div>
                </button>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'GeoRiskMapPage',
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
