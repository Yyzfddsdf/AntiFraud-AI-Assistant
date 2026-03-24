import { computed, ref } from 'vue';

export function useChartsModule(deps) {
  const riskInterval = ref('day');
  const riskData = ref(null);
  const riskCache = ref({});
  const adminStatsInterval = ref('day');
  const adminStatsData = ref(null);
  const adminStatsCache = ref({});
  const adminGraphData = ref(null);
  const adminGraphCache = ref(null);
  const adminTargetGroupChartData = ref(null);
  const adminTargetGroupChartCache = ref({});
  const selectedGraphTargetGroup = ref('');
  const showGraphModal = ref(false);
  const selectedGraphProfile = ref('');

  let pieChartInstance = null;
  let lineChartInstance = null;
  let adminTrendChart = null;
  let adminTypeChart = null;
  let adminTargetChart = null;
  let adminTargetGroupBarChart = null;
  let adminNetworkInstance = null;

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
        const fmt = d => `${d.getMonth() + 1}.${d.getDate()}`;
        return `${year}年第${week}周 (${fmt(start)}-${fmt(end)})`;
      } catch {
        return label;
      }
    }
    if (interval === 'month' && /^\d{4}-\d{2}$/.test(label)) {
      const [y, m] = label.split('-');
      return `${y}年${parseInt(m, 10)}月`;
    }
    return label;
  };

  const fillTrendGaps = (sparseTrend, interval) => {
    if (!sparseTrend || sparseTrend.length === 0) return [];
    const sorted = [...sparseTrend].sort((a, b) => a.time_bucket.localeCompare(b.time_bucket));
    const startBucket = sorted[0].time_bucket;
    const endBucket = sorted[sorted.length - 1].time_bucket;
    const filled = [];
    const dataMap = new Map(sorted.map(item => [item.time_bucket, item]));
    let current = startBucket;
    let count = 0;

    while (current <= endBucket && count < 500) {
      count += 1;
      if (dataMap.has(current)) {
        filled.push(dataMap.get(current));
      } else {
        const zeroPoint = { time_bucket: current, total: 0 };
        if ('high' in sorted[0]) {
          zeroPoint.high = 0;
          zeroPoint.medium = 0;
          zeroPoint.low = 0;
        }
        if ('count' in sorted[0]) {
          zeroPoint.count = 0;
        }
        filled.push(zeroPoint);
      }

      if (interval === 'day') {
        const date = new Date(current);
        date.setDate(date.getDate() + 1);
        current = date.toISOString().split('T')[0];
      } else if (interval === 'week') {
        const [y, w] = current.split('-W').map(Number);
        let nextW = w + 1;
        let nextY = y;
        if (nextW > 53) { nextW = 1; nextY += 1; }
        current = `${nextY}-W${String(nextW).padStart(2, '0')}`;
      } else if (interval === 'month') {
        const [y, m] = current.split('-').map(Number);
        let nextM = m + 1;
        let nextY = y;
        if (nextM > 12) { nextM = 1; nextY += 1; }
        current = `${nextY}-${String(nextM).padStart(2, '0')}`;
      } else {
        break;
      }
    }
    return filled;
  };

  const renderCharts = () => {
    if (!riskData.value || typeof window.echarts === 'undefined') return;
    const stats = riskData.value.stats;
    const trend = fillTrendGaps(riskData.value.trend, riskInterval.value);
    if (pieChartInstance?.dispose) pieChartInstance.dispose();
    if (lineChartInstance?.dispose) lineChartInstance.dispose();

    const pieDom = document.getElementById('riskPieChart');
    if (pieDom) {
      pieChartInstance = echarts.init(pieDom);
      pieChartInstance.setOption({
        tooltip: { trigger: 'item', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
        series: [{
          name: '风险分布',
          type: 'pie',
          radius: ['50%', '80%'],
          itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
          label: { show: false },
          emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
          data: [
            { value: stats.high, name: '高风险', itemStyle: { color: '#ef4444' } },
            { value: stats.medium, name: '中风险', itemStyle: { color: '#f59e0b' } },
            { value: stats.low, name: '低风险', itemStyle: { color: '#10b981' } }
          ]
        }]
      });
    }

    const lineDom = document.getElementById('riskLineChart');
    if (lineDom) {
      lineChartInstance = echarts.init(lineDom);
      lineChartInstance.setOption({
        tooltip: { trigger: 'axis', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
        legend: { bottom: 0, textStyle: { fontSize: 11 } },
        grid: { left: '3%', right: '4%', top: '10%', bottom: '15%', containLabel: true },
        xAxis: { type: 'category', boundaryGap: false, data: trend.map(item => formatChartLabel(item.time_bucket)), axisLabel: { color: '#64748b' } },
        yAxis: { type: 'value', axisLabel: { color: '#64748b' }, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.1)' } } },
        series: [
          { name: '高风险', type: 'line', smooth: true, data: trend.map(item => item.high), itemStyle: { color: '#ef4444' }, lineStyle: { width: 3 } },
          { name: '中风险', type: 'line', smooth: true, data: trend.map(item => item.medium), itemStyle: { color: '#f59e0b' }, lineStyle: { width: 3 } },
          { name: '低风险', type: 'line', smooth: true, data: trend.map(item => item.low), itemStyle: { color: '#10b981' }, lineStyle: { width: 3 } }
        ]
      });
    }
  };

  const fetchRiskTrend = async (forceRefresh = false) => {
    if (!deps.isAuthenticated.value) return;
    const interval = riskInterval.value;
    if (riskCache.value[interval]) {
      riskData.value = riskCache.value[interval];
      setTimeout(() => renderCharts(), 0);
    }
    try {
      const res = await deps.request(`/scam/multimodal/history/overview?interval=${interval}`, 'GET', null, { silent: true });
      if (res) {
        const cachedData = riskCache.value[interval];
        const hasChanged = !cachedData || deps.stableJSONStringify(cachedData) !== deps.stableJSONStringify(res);
        if (hasChanged || forceRefresh) {
          riskData.value = res;
          riskCache.value[interval] = res;
          setTimeout(() => renderCharts(), 100);
          if (forceRefresh) deps.showToast('数据已更新');
        }
      }
    } catch (e) {
      console.error('Fetch risk trend failed:', e);
    }
  };

  const getRiskTrendAnalysisClass = (trendText) => {
    switch ((trendText || '').trim()) {
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

  const formatAdminChartLabel = (label) => formatChartLabel(label, adminStatsInterval.value);

  const fetchAdminStats = async (forceRefresh = false) => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    const interval = adminStatsInterval.value;
    if (adminStatsCache.value[interval]) {
      adminStatsData.value = adminStatsCache.value[interval];
      setTimeout(() => renderAdminCharts(), 0);
    }
    try {
      const res = await deps.request(`/scam/case-library/cases/overview?interval=${interval}`, 'GET', null, { silent: true });
      if (res) {
        const cachedData = adminStatsCache.value[interval];
        const hasChanged = !cachedData || deps.stableJSONStringify(cachedData) !== deps.stableJSONStringify(res);
        if (hasChanged || forceRefresh) {
          adminStatsData.value = res;
          adminStatsCache.value[interval] = res;
          setTimeout(() => renderAdminCharts(), 100);
          if (forceRefresh) deps.showToast('全景数据已更新');
        }
      }
    } catch (e) {
      console.error('Fetch admin stats failed:', e);
    }
    await fetchAdminCaseGraph(forceRefresh);
  };

  const fetchAdminCaseGraph = async (forceRefresh = false) => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    if (adminGraphCache.value) {
      adminGraphData.value = adminGraphCache.value;
    }
    try {
      const res = await deps.request('/scam/case-library/cases/graph?top_k=5', 'GET', null, { silent: true });
      if (res) {
        const cachedData = adminGraphCache.value;
        const hasChanged = !cachedData || deps.stableJSONStringify(cachedData) !== deps.stableJSONStringify(res);
        if (hasChanged || forceRefresh) {
          adminGraphData.value = res;
          adminGraphCache.value = res;
          
          if (res.profiles && res.profiles.length > 0) {
            selectedGraphProfile.value = res.profiles[0].scam_type;
          } else {
            selectedGraphProfile.value = '';
          }
        }
      }
      if (selectedGraphTargetGroup.value) {
        await fetchAdminTargetGroupChart(selectedGraphTargetGroup.value, forceRefresh);
      }
    } catch (e) {
      console.error('Fetch admin graph failed:', e);
    }
  };

  const clearAdminTargetGroupFocus = () => {
    selectedGraphTargetGroup.value = '';
    adminTargetGroupChartData.value = null;
    if (adminTargetGroupBarChart?.dispose) {
      adminTargetGroupBarChart.dispose();
      adminTargetGroupBarChart = null;
    }
  };

  const fetchAdminTargetGroupChart = async (targetGroup, forceRefresh = false) => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    const normalizedTargetGroup = String(targetGroup || '').trim();
    if (!normalizedTargetGroup) {
      clearAdminTargetGroupFocus();
      return;
    }

    selectedGraphTargetGroup.value = normalizedTargetGroup;
    if (adminTargetGroupChartCache.value[normalizedTargetGroup] && !forceRefresh) {
      adminTargetGroupChartData.value = adminTargetGroupChartCache.value[normalizedTargetGroup];
      setTimeout(() => renderAdminTargetGroupBarChart(), 0);
      return;
    }

    try {
      const query = `/scam/case-library/cases/graph?top_k=5&focus_group=${encodeURIComponent(normalizedTargetGroup)}`;
      const res = await deps.request(query, 'GET', null, { silent: true });
      const nextData = res && Array.isArray(res.target_group_top_scam_types) ? (res.target_group_top_scam_types[0] || null) : null;
      adminTargetGroupChartData.value = nextData;
      if (nextData) {
        adminTargetGroupChartCache.value[normalizedTargetGroup] = nextData;
        setTimeout(() => renderAdminTargetGroupBarChart(), 0);
      } else if (adminTargetGroupBarChart?.dispose) {
        adminTargetGroupBarChart.dispose();
        adminTargetGroupBarChart = null;
      }
    } catch (e) {
      console.error('Fetch admin target group chart failed:', e);
    }
  };

  const openGraphModal = () => {
    showGraphModal.value = true;
    setTimeout(() => renderAdminGraphNetwork(), 300);
  };

  const renderAdminGraphNetwork = () => {
    if (!adminGraphData.value?.graph) return;
    const container = document.getElementById('adminGraphNetwork');
    if (!container) return;
    const { nodes, edges } = adminGraphData.value.graph;

    const visNodes = new vis.DataSet(nodes.map((node) => {
      let background = '#6366f1';
      let border = '#818cf8';
      let size = 20;
      if (node.node_type === 'scam_type') {
        background = '#d946ef';
        border = '#f0abfc';
        size = 35;
      } else if (node.node_type === 'target_group') {
        background = '#10b981';
        border = '#6ee7b7';
        size = 42;
      } else if (node.node_type === 'keyword') {
        background = '#818cf8';
        border = '#c7d2fe';
        size = 18;
      }
      return {
        id: node.id,
        label: node.label,
        nodeType: node.node_type,
        title: `类型: ${node.node_type}\n名称: ${node.label}${node.properties?.case_count ? `\n案件数: ${node.properties.case_count}` : ''}`,
        color: {
          background,
          border,
          highlight: { background, border: '#000' },
          hover: { background, border }
        },
        font: { color: '#475569', size: 14, face: 'Plus Jakarta Sans', weight: '800' },
        size,
        shape: node.node_type === 'target_group' ? 'diamond' : 'dot',
        borderWidth: node.node_type === 'target_group' ? 6 : 4,
        shadow: { enabled: true, color: 'rgba(0,0,0,0.1)', size: 12, x: 0, y: 4 }
      };
    }));

    const visEdges = new vis.DataSet(edges.map((edge) => {
      const relation = String(edge.relation || edge.relation_type || '').trim();
      let label = '';
      let dashes = false;
      let color = '#e2e8f0';
      if (relation === 'similar' || relation === 'similar_to') { label = '相似'; dashes = true; color = '#fbcfe8'; }
      else if (relation === 'targets' || relation === 'target_of') { label = '针对'; color = '#d1fae5'; }
      else if (relation === 'keyword' || relation === 'has_keyword') { label = '关键词'; color = '#e0e7ff'; }

      return {
        from: edge.source,
        to: edge.target,
        label,
        arrows: { to: { enabled: true, scaleFactor: 0.5, type: 'arrow' } },
        font: { size: 11, align: 'middle', color: '#94a3b8', strokeWidth: 0 },
        color: { color, highlight: color, hover: color },
        width: edge.weight ? Math.max(1.5, edge.weight * 4) : 1.5,
        dashes,
        smooth: { type: 'curvedCW', roundness: 0.2 }
      };
    }));

    if (adminNetworkInstance) adminNetworkInstance.destroy();
    adminNetworkInstance = new vis.Network(container, { nodes: visNodes, edges: visEdges }, {
      nodes: { borderWidthSelected: 2 },
      edges: { hoverWidth: 1.5, selectionWidth: 2 },
      physics: {
        forceAtlas2Based: { gravitationalConstant: -80, centralGravity: 0.005, springLength: 180, springConstant: 0.04, avoidOverlap: 1 },
        maxVelocity: 45,
        solver: 'forceAtlas2Based',
        timestep: 0.35,
        stabilization: { iterations: 200, updateInterval: 25 }
      },
      interaction: { hover: true, tooltipDelay: 100, navigationButtons: false, keyboard: true, zoomView: true, dragView: true }
    });

    adminNetworkInstance.on('click', (params) => {
      if (params.nodes.length > 0) {
        const nodeId = params.nodes[0];
        const node = visNodes.get(nodeId);
        if (node && node.nodeType === 'target_group') {
          const connectedNodeIds = adminNetworkInstance.getConnectedNodes(nodeId);
          adminNetworkInstance.selectNodes([nodeId, ...connectedNodeIds]);
          fetchAdminTargetGroupChart(node.label);
          deps.showToast(`已切换至【${node.label}】人群视角，关联 ${connectedNodeIds.length} 种诈骗手法`, 'success');
        }
      }
    });
  };

  const resetGraphZoom = () => {
    if (adminNetworkInstance) adminNetworkInstance.fit({ animation: true });
  };

  const formatGraphScore = (score) => {
    const value = Number(score);
    if (!Number.isFinite(value)) return '--';
    return `${(value * 100).toFixed(1)}%`;
  };

  const availableGraphTargetGroups = computed(() => {
    if (!adminGraphData.value || !Array.isArray(adminGraphData.value.target_group_top_scam_types)) {
      return [];
    }
    return adminGraphData.value.target_group_top_scam_types;
  });

  const renderAdminTargetGroupBarChart = () => {
    if (!adminTargetGroupChartData.value || typeof window.echarts === 'undefined') return;
    const items = Array.isArray(adminTargetGroupChartData.value.top_scam_types) ? adminTargetGroupChartData.value.top_scam_types : [];
    const rawScores = items.map(item => Number((Number(item.score || 0) * 100).toFixed(2)));
    const maxScore = rawScores.length > 0 ? Math.max(...rawScores) : 100;
    const displayMax = Math.ceil(maxScore * 1.15 / 10) * 10;
    const targetDom = document.getElementById('adminTargetGroupBarChart');
    if (!targetDom) return;
    if (adminTargetGroupBarChart?.dispose) adminTargetGroupBarChart.dispose();
    adminTargetGroupBarChart = echarts.init(targetDom);

    adminTargetGroupBarChart.setOption({
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        backgroundColor: 'rgba(15, 23, 42, 0.9)',
        borderColor: 'rgba(255, 255, 255, 0.1)',
        borderWidth: 1,
        textStyle: { color: '#fff', fontSize: 12 },
        formatter: (params) => {
          const data = params.find(p => p.seriesName === '案件占比');
          if (!data) return '';
          return `<div class="font-bold mb-1">${data.name}</div><div class="flex items-center gap-2"><span class="w-2 h-2 rounded-full" style="background:${data.color}"></span><span>占比: ${data.value}%</span></div>`;
        }
      },
      grid: { left: '3%', right: '15%', bottom: '3%', top: '3%', containLabel: true },
      xAxis: { type: 'value', max: displayMax, axisLabel: { formatter: '{value}%', color: '#64748b', fontWeight: 'bold' }, splitLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.1)' } } },
      yAxis: { type: 'category', data: items.map(item => item.scam_type).reverse(), axisLabel: { color: '#334155', fontWeight: 'bold', fontSize: 12 }, axisLine: { show: false }, axisTick: { show: false } },
      series: [
        { name: 'placeholder', type: 'bar', itemStyle: { color: 'rgba(148, 163, 184, 0.05)', borderRadius: [0, 20, 20, 0] }, barGap: '-100%', barWidth: 32, data: items.map(() => displayMax), animation: false, tooltip: { show: false } },
        {
          name: '案件占比',
          type: 'bar',
          data: items.map((item, index) => {
            const gradients = [
              ['#3b82f6', '#6366f1', '#d946ef'],
              ['#10b981', '#3b82f6', '#6366f1'],
              ['#f59e0b', '#ef4444', '#d946ef'],
              ['#6366f1', '#a855f7', '#ec4899'],
              ['#0ea5e9', '#2dd4bf', '#10b981']
            ];
            const colors = gradients[index % gradients.length];
            return {
              value: Number((Number(item.score || 0) * 100).toFixed(2)),
              itemStyle: {
                borderRadius: [0, 20, 20, 0],
                color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
                  { offset: 0, color: colors[0] },
                  { offset: 0.5, color: colors[1] },
                  { offset: 1, color: colors[2] }
                ])
              }
            };
          }).reverse(),
          barWidth: 32,
          label: { show: true, position: 'right', formatter: '{c}%', color: '#475569', fontWeight: 'bold', distance: 10 }
        }
      ]
    });
  };

  const renderAdminCharts = () => {
    if (!adminStatsData.value || typeof window.echarts === 'undefined') return;
    const trend = fillTrendGaps(adminStatsData.value.trend, adminStatsInterval.value);
    const { by_scam_type, by_target_group } = adminStatsData.value;

    if (adminTrendChart?.dispose) adminTrendChart.dispose();
    if (adminTypeChart?.dispose) adminTypeChart.dispose();
    if (adminTargetChart?.dispose) adminTargetChart.dispose();

    const trendDom = document.getElementById('adminTrendChart');
    if (trendDom) {
      adminTrendChart = echarts.init(trendDom);
      adminTrendChart.setOption({
        tooltip: { trigger: 'axis', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', boundaryGap: false, data: trend.map(item => formatAdminChartLabel(item.time_bucket)), axisLabel: { color: '#64748b' } },
        yAxis: { type: 'value', axisLabel: { color: '#64748b' }, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.1)' } } },
        series: [{
          name: '新增案件数',
          type: 'line',
          smooth: true,
          data: trend.map(item => item.count),
          symbol: 'circle',
          symbolSize: 8,
          itemStyle: { color: '#6366f1' },
          areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: 'rgba(99, 102, 241, 0.3)' }, { offset: 1, color: 'rgba(99, 102, 241, 0)' }]) },
          lineStyle: { width: 3 }
        }]
      });
    }

    const typeDom = document.getElementById('adminTypeChart');
    if (typeDom) {
      adminTypeChart = echarts.init(typeDom);
      adminTypeChart.setOption({
        tooltip: { trigger: 'item', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
        legend: { orient: 'vertical', right: 10, top: 'center', itemWidth: 10, itemHeight: 10, textStyle: { fontSize: 11, color: '#64748b' } },
        series: [{ name: '诈骗类型', type: 'pie', radius: ['40%', '70%'], center: ['40%', '50%'], itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 }, label: { show: false }, emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } }, data: by_scam_type.map(i => ({ value: i.count, name: i.name })) }]
      });
    }

    const targetDom = document.getElementById('adminTargetChart');
    if (targetDom) {
      adminTargetChart = echarts.init(targetDom);
      adminTargetChart.setOption({
        tooltip: { trigger: 'item', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
        legend: { orient: 'vertical', right: 10, top: 'center', itemWidth: 10, itemHeight: 10, textStyle: { fontSize: 11, color: '#64748b' } },
        series: [{ name: '目标人群', type: 'pie', radius: '70%', center: ['40%', '50%'], itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 }, label: { show: false }, data: by_target_group.map(i => ({ value: i.count, name: i.name })) }]
      });
    }
  };

  const resizeCharts = () => {
    if (adminTargetGroupBarChart?.resize) adminTargetGroupBarChart.resize();
    if (adminTrendChart?.resize) adminTrendChart.resize();
    if (adminTypeChart?.resize) adminTypeChart.resize();
    if (adminTargetChart?.resize) adminTargetChart.resize();
    if (pieChartInstance?.resize) pieChartInstance.resize();
    if (lineChartInstance?.resize) lineChartInstance.resize();
  };

  const disposeCharts = () => {
    if (adminNetworkInstance) {
      adminNetworkInstance.destroy();
      adminNetworkInstance = null;
    }
    if (adminTargetGroupBarChart?.dispose) adminTargetGroupBarChart.dispose();
    if (adminTrendChart?.dispose) adminTrendChart.dispose();
    if (adminTypeChart?.dispose) adminTypeChart.dispose();
    if (adminTargetChart?.dispose) adminTargetChart.dispose();
    if (pieChartInstance?.dispose) pieChartInstance.dispose();
    if (lineChartInstance?.dispose) lineChartInstance.dispose();
  };

  return {
    riskInterval,
    riskData,
    adminStatsInterval,
    adminStatsData,
    adminGraphData,
    adminTargetGroupChartData,
    selectedGraphTargetGroup,
    showGraphModal,
    selectedGraphProfile,
    availableGraphTargetGroups,
    fetchRiskTrend,
    getRiskTrendAnalysisClass,
    fetchAdminStats,
    formatGraphScore,
    openGraphModal,
    resetGraphZoom,
    fetchAdminTargetGroupChart,
    clearAdminTargetGroupFocus,
    resizeCharts,
    disposeCharts
  };
}
