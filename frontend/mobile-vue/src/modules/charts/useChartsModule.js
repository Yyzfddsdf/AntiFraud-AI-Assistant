import { computed, ref } from 'vue';

export function useChartsModule(deps) {
  const riskInterval = ref('day');
  const riskData = ref(null);
  const regionCaseStats = ref(null);
  const riskCache = {};
  let pieChartInstance = null;
  let lineChartInstance = null;
  let lineDetailChartInstance = null;
  let hasWarnedMissingChartLibrary = false;

  const todayRiskStatsSummary = computed(() => {
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
    const tomorrowStart = todayStart + 24 * 60 * 60 * 1000;
    const todayHistory = (Array.isArray(deps.history.value) ? deps.history.value : []).filter((item) => {
      const createdAt = new Date(item && item.created_at ? item.created_at : '').getTime();
      return Number.isFinite(createdAt) && createdAt >= todayStart && createdAt < tomorrowStart;
    });
    const countByRiskLevel = (level) => todayHistory.filter((item) => String(item && item.risk_level ? item.risk_level : '').trim() === level).length;
    return {
      total: todayHistory.length,
      high: countByRiskLevel('高'),
      medium: countByRiskLevel('中'),
      low: countByRiskLevel('低')
    };
  });

  const riskStatsSummary = computed(() => {
    const stats = riskData.value && riskData.value.stats ? riskData.value.stats : {};
    return {
      total: Number(stats.total) || 0,
      high: Number(stats.high) || 0,
      medium: Number(stats.medium) || 0,
      low: Number(stats.low) || 0
    };
  });

  const formatChartLabel = (label, intervalType) => {
    const interval = intervalType || riskInterval.value;
    if (interval === 'week' && label.includes('-W')) {
      try {
        const [yearStr, weekStr] = label.split('-W');
        const year = parseInt(yearStr, 10);
        const week = parseInt(weekStr, 10);
        const jan4 = new Date(year, 0, 4);
        const jan4Day = jan4.getDay() || 7;
        const week1Start = new Date(year, 0, 4 - jan4Day + 1);
        const start = new Date(week1Start.getTime() + (week - 1) * 7 * 86400000);
        const end = new Date(start.getTime() + 6 * 86400000);
        const fmt = (date) => `${date.getMonth() + 1}.${date.getDate()}`;
        return `${year}年第${week}周 (${fmt(start)}-${fmt(end)})`;
      } catch {
        return label;
      }
    }

    if (interval === 'month' && /^\d{4}-\d{2}$/.test(label)) {
      const [year, month] = label.split('-');
      return `${year}年${parseInt(month, 10)}月`;
    }

    return label;
  };

  const fillTrendGaps = (sparseTrend, interval) => {
    if (!sparseTrend || sparseTrend.length === 0) return [];

    const sorted = [...sparseTrend].sort((a, b) => a.time_bucket.localeCompare(b.time_bucket));
    const startBucket = sorted[0].time_bucket;
    const endBucket = sorted[sorted.length - 1].time_bucket;
    const filled = [];
    const dataMap = new Map(sorted.map((item) => [item.time_bucket, item]));
    let current = startBucket;
    let count = 0;

    while (current <= endBucket && count < 500) {
      count += 1;
      if (dataMap.has(current)) {
        filled.push(dataMap.get(current));
      } else {
        filled.push({
          time_bucket: current,
          total: 0,
          high: 0,
          medium: 0,
          low: 0
        });
      }

      if (interval === 'day') {
        const date = new Date(current);
        date.setDate(date.getDate() + 1);
        current = date.toISOString().split('T')[0];
      } else if (interval === 'week') {
        const [year, week] = current.split('-W').map(Number);
        let nextWeek = week + 1;
        let nextYear = year;
        if (nextWeek > 53) {
          nextWeek = 1;
          nextYear += 1;
        }
        current = `${nextYear}-W${String(nextWeek).padStart(2, '0')}`;
      } else if (interval === 'month') {
        const [year, month] = current.split('-').map(Number);
        let nextMonth = month + 1;
        let nextYear = year;
        if (nextMonth > 12) {
          nextMonth = 1;
          nextYear += 1;
        }
        current = `${nextYear}-${String(nextMonth).padStart(2, '0')}`;
      } else {
        break;
      }
    }

    return filled;
  };

  const buildRiskLineOption = (trend, compact = false) => ({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(15, 23, 42, 0.9)',
      textStyle: { color: '#fff' }
    },
    legend: compact ? undefined : { bottom: 0, textStyle: { fontSize: 11 } },
    grid: compact
      ? { left: '2%', right: '2%', top: '8%', bottom: '6%', containLabel: false }
      : { left: '3%', right: '4%', top: '10%', bottom: '15%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: trend.map((item) => formatChartLabel(item.time_bucket)),
      axisLabel: compact ? { show: false } : { color: '#64748b' },
      axisLine: compact ? { show: false } : undefined,
      axisTick: compact ? { show: false } : undefined
    },
    yAxis: {
      type: 'value',
      axisLabel: compact ? { show: false } : { color: '#64748b' },
      axisLine: compact ? { show: false } : undefined,
      axisTick: compact ? { show: false } : undefined,
      splitLine: compact ? { show: false } : { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.1)' } }
    },
    series: [
      {
        name: '高风险',
        type: 'line',
        smooth: true,
        showSymbol: !compact,
        symbolSize: compact ? 0 : 6,
        data: trend.map((item) => item.high),
        itemStyle: { color: '#ef4444' },
        lineStyle: { width: compact ? 2.5 : 3 },
        areaStyle: compact ? { color: 'rgba(239, 68, 68, 0.08)' } : undefined
      },
      {
        name: '中风险',
        type: 'line',
        smooth: true,
        showSymbol: !compact,
        symbolSize: compact ? 0 : 6,
        data: trend.map((item) => item.medium),
        itemStyle: { color: '#f59e0b' },
        lineStyle: { width: compact ? 2.5 : 3 },
        areaStyle: compact ? { color: 'rgba(245, 158, 11, 0.08)' } : undefined
      },
      {
        name: '低风险',
        type: 'line',
        smooth: true,
        showSymbol: !compact,
        symbolSize: compact ? 0 : 6,
        data: trend.map((item) => item.low),
        itemStyle: { color: '#10b981' },
        lineStyle: { width: compact ? 2.5 : 3 },
        areaStyle: compact ? { color: 'rgba(16, 185, 129, 0.08)' } : undefined
      }
    ]
  });

  const disposeCharts = () => {
    if (pieChartInstance && typeof pieChartInstance.dispose === 'function') pieChartInstance.dispose();
    if (lineChartInstance && typeof lineChartInstance.dispose === 'function') lineChartInstance.dispose();
    if (lineDetailChartInstance && typeof lineDetailChartInstance.dispose === 'function') lineDetailChartInstance.dispose();
    pieChartInstance = null;
    lineChartInstance = null;
    lineDetailChartInstance = null;
  };

  const renderCharts = () => {
    if (!riskData.value) {
      disposeCharts();
      return;
    }

    if (typeof window.echarts === 'undefined') {
      if (!hasWarnedMissingChartLibrary) {
        console.warn('ECharts 未加载，已跳过移动端图表渲染。');
        hasWarnedMissingChartLibrary = true;
      }
      return;
    }

    hasWarnedMissingChartLibrary = false;
    const stats = riskData.value.stats || { high: 0, medium: 0, low: 0, total: 0 };
    const trend = fillTrendGaps(riskData.value.trend, riskInterval.value);

    disposeCharts();

    const pieDom = document.getElementById('riskPieChart');
    if (pieDom) {
      pieChartInstance = window.echarts.init(pieDom);
      pieChartInstance.setOption({
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(15, 23, 42, 0.9)',
          textStyle: { color: '#fff' }
        },
        graphic: [
          {
            type: 'text',
            left: 'center',
            top: '43%',
            style: {
              text: String(Number(stats.total) || 0),
              fill: '#0f172a',
              fontSize: 26,
              fontWeight: '700'
            }
          },
          {
            type: 'text',
            left: 'center',
            top: '58%',
            style: {
              text: '总检测',
              fill: '#64748b',
              fontSize: 12
            }
          }
        ],
        series: [{
          name: '风险分布',
          type: 'pie',
          radius: ['56%', '82%'],
          avoidLabelOverlap: false,
          itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
          label: { show: false },
          emphasis: { label: { show: true, fontSize: 13, fontWeight: 'bold' } },
          data: [
            { value: Number(stats.high) || 0, name: '高风险', itemStyle: { color: '#ef4444' } },
            { value: Number(stats.medium) || 0, name: '中风险', itemStyle: { color: '#f59e0b' } },
            { value: Number(stats.low) || 0, name: '低风险', itemStyle: { color: '#10b981' } }
          ]
        }]
      });
    }

    const lineDom = document.getElementById('riskLineChart');
    if (lineDom) {
      lineChartInstance = window.echarts.init(lineDom);
      lineChartInstance.setOption(buildRiskLineOption(trend, true));
    }

    const lineDetailDom = document.getElementById('riskLineChartDetail');
    if (lineDetailDom) {
      lineDetailChartInstance = window.echarts.init(lineDetailDom);
      lineDetailChartInstance.setOption(buildRiskLineOption(trend, false));
    }
  };

  const fetchRiskTrend = async (forceRefresh = false) => {
    if (!deps.isAuthenticated.value) return;

    riskInterval.value = 'day';
    const interval = 'day';

    if (riskCache[interval]) {
      riskData.value = riskCache[interval];
      setTimeout(() => renderCharts(), 0);
    }

    try {
      const res = await deps.request(`/scam/multimodal/history/overview?interval=${interval}`, 'GET', null, { silent: true });
      if (res) {
        const cachedData = riskCache[interval];
        const hasChanged = !cachedData || deps.stableJSONStringify(cachedData) !== deps.stableJSONStringify(res);
        if (hasChanged || forceRefresh) {
          riskData.value = res;
          riskCache[interval] = res;
          setTimeout(() => renderCharts(), 100);
          if (forceRefresh) deps.showToast('数据已更新');
        }
      }
    } catch (error) {
      console.error('Fetch risk trend failed:', error);
    }
  };

  const fetchCurrentRegionCaseStats = async () => {
    if (!deps.isAuthenticated.value) return;
    try {
      const res = await deps.request('/regions/cases/stats/current', 'GET', null, { silent: true });
      if (res) {
        regionCaseStats.value = res;
      }
    } catch (error) {
      console.error('Fetch current region case stats failed:', error);
    }
  };

  const getRegionStatsSignal = (statsPayload) => {
    const summary = statsPayload && statsPayload.summary ? statsPayload.summary : null;
    if (!summary) {
      return {
        level: 'neutral',
        title: '地区风险态势待补充',
        detail: '当前统计样本不足，建议继续观察。'
      };
    }

    const todayCount = Number(summary.today_count) || 0;
    const last7dCount = Number(summary.last_7d_count) || 0;
    const totalCount = Number(summary.total_count) || 0;
    const highCount = Number(summary.high_count) || 0;
    const weeklyAverage = last7dCount > 0 ? last7dCount / 7 : 0;
    const highRiskRatio = totalCount > 0 ? highCount / totalCount : 0;

    if (todayCount >= Math.max(3, weeklyAverage * 1.5) || highRiskRatio >= 0.35) {
      return {
        level: 'high',
        title: '近期风险有抬升信号',
        detail: '今日增量或高风险占比偏高，建议减少陌生转账与验证码操作。'
      };
    }
    if (todayCount >= Math.max(1, weeklyAverage * 0.8) || highRiskRatio >= 0.2) {
      return {
        level: 'medium',
        title: '近期风险保持活跃',
        detail: '建议重点核验陌生来电、客服退款与投资荐股类话术。'
      };
    }
    return {
      level: 'low',
      title: '近期风险相对平稳',
      detail: '仍需保持警惕，遇到催转账和索要验证码请先核实。'
    };
  };

  const getRegionTopScamHint = (statsPayload) => {
    const topList = Array.isArray(statsPayload?.top_scam_types) ? statsPayload.top_scam_types : [];
    if (!topList.length) {
      return '近期未形成明显高发类型，注意通用防诈规则。';
    }
    const topItem = topList[0] || {};
    const scamType = String(topItem.scam_type || '').trim() || '当前高发类型';
    const count = Number(topItem.count) || 0;
    return `最近高发：${scamType}${count > 0 ? `（${count}起）` : ''}，同类话术请优先核验来源。`;
  };

  const getRegionSignalClass = (signalLevel) => {
    const level = String(signalLevel || '').trim();
    if (level === 'high') return 'bg-red-50 text-red-700 border-red-200';
    if (level === 'medium') return 'bg-amber-50 text-amber-700 border-amber-200';
    if (level === 'low') return 'bg-emerald-50 text-emerald-700 border-emerald-200';
    return 'bg-slate-50 text-slate-700 border-slate-200';
  };

  const getRecentRiskTrendRows = (limit = 7) => {
    if (!riskData.value || !Array.isArray(riskData.value.trend)) return [];
    const normalizedLimit = Number(limit) > 0 ? Number(limit) : 7;
    return fillTrendGaps(riskData.value.trend, riskInterval.value)
      .slice(-normalizedLimit)
      .reverse();
  };

  const getRiskTrendAnalysisClass = (trendText) => {
    switch (String(trendText || '').trim()) {
      case '上升':
        return 'bg-red-50 text-red-700 ring-1 ring-red-200';
      case '下降':
        return 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200';
      case '平稳':
        return 'bg-amber-50 text-amber-700 ring-1 ring-amber-200';
      default:
        return 'bg-slate-100 text-slate-600 ring-1 ring-slate-200';
    }
  };

  const formatRiskTrendDescriptor = (trendText, dimension = 'overall') => {
    const normalized = String(trendText || '').trim();
    if (dimension === 'high') {
      if (normalized === '上升') return '高危暴露增强';
      if (normalized === '下降') return '高危暴露收敛';
      if (normalized === '平稳') return '高危信号平稳';
      return '高危信号待观察';
    }

    if (normalized === '上升') return '风险热度抬升';
    if (normalized === '下降') return '风险热度回落';
    if (normalized === '平稳') return '风险热度平稳';
    return '风险热度待观察';
  };

  const getRiskTrendHeadline = (analysis) => {
    const overall = String(analysis?.overall_trend || '').trim();
    const high = String(analysis?.high_risk_trend || '').trim();
    if (overall === '上升' && high === '上升') return 'AI 研判：风险热度正在拉升';
    if (overall === '下降' && high === '下降') return 'AI 研判：风险热度出现回落';
    if (high === '上升') return 'AI 研判：高危暴露正在增强';
    if (overall === '上升') return 'AI 研判：整体风险有所抬头';
    if (overall === '下降') return 'AI 研判：整体风险趋于缓和';
    if (overall === '平稳' && high === '平稳') return 'AI 研判：当前波动较为平稳';
    return 'AI 研判：风险走势仍需持续观察';
  };

  const resizeCharts = () => {
    if (pieChartInstance && typeof pieChartInstance.resize === 'function') pieChartInstance.resize();
    if (lineChartInstance && typeof lineChartInstance.resize === 'function') lineChartInstance.resize();
    if (lineDetailChartInstance && typeof lineDetailChartInstance.resize === 'function') lineDetailChartInstance.resize();
  };

  return {
    riskInterval,
    riskData,
    regionCaseStats,
    todayRiskStatsSummary,
    riskStatsSummary,
    fetchRiskTrend,
    fetchCurrentRegionCaseStats,
    formatChartLabel,
    getRecentRiskTrendRows,
    getRiskTrendAnalysisClass,
    formatRiskTrendDescriptor,
    getRiskTrendHeadline,
    getRegionStatsSignal,
    getRegionTopScamHint,
    getRegionSignalClass,
    resizeCharts,
    disposeCharts
  };
}
