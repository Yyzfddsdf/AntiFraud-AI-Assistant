import { computed, ref, watch } from 'vue';

const DEFAULT_GEO_BOUNDARY_CODE = '100000';
const GEO_LEVEL_PROVINCE = 'province';
const GEO_LEVEL_CITY = 'city';
const GEO_LEVEL_DISTRICT = 'district';

const buildGeoBoundaryPath = (code) => `/scam/case-library/maps/geojson?code=${encodeURIComponent(String(code || DEFAULT_GEO_BOUNDARY_CODE).trim() || DEFAULT_GEO_BOUNDARY_CODE)}`;

const riskColors = {
  高: '#fb7185',
  中: '#f59e0b',
  低: '#22c55e'
};

const riskAreaColors = {
  高: 'rgba(248, 113, 113, 0.68)',
  中: 'rgba(245, 158, 11, 0.54)',
  低: 'rgba(34, 197, 94, 0.4)'
};

export function useGeoRiskMapModule(deps) {
  const geoOverviewSummary = ref(null);
  const geoMapLoading = ref(false);
  const geoMapError = ref('');
  const geoRegionCasesVisible = ref(false);
  const geoRegionCasesLoading = ref(false);
  const geoRegionCasesError = ref('');
  const geoRegionCasesData = ref(null);
  const geoRegionCasesPage = ref(1);
  const geoRegionCasesPageSize = ref(10);
  const geoRegionCasesPageSizeOptions = [10, 20, 50];
  const geoSearchKeyword = ref('');
  const geoSelectedWindow = ref('last_7d');
  const geoViewMode = ref(GEO_LEVEL_PROVINCE);
  const geoSelectedProvinceCode = ref('');
  const geoSelectedProvinceName = ref('');
  const geoSelectedCityCode = ref('');
  const geoSelectedCityName = ref('');
  const geoWindowOptions = [
    { value: 'today', label: '今日新增' },
    { value: 'last_7d', label: '近7天' },
    { value: 'last_30d', label: '近30天' },
    { value: 'all_time', label: '全历史' }
  ];

  const geoProvinceEntries = ref([]);
  const geoCurrentEntriesState = ref([]);
  const geoChildCache = {
    [GEO_LEVEL_CITY]: new Map(),
    [GEO_LEVEL_DISTRICT]: new Map()
  };
  const geoJSONCache = new Map();
  let geoChartInstance = null;

  const geoMapData = computed(() => ({
    summary: geoOverviewSummary.value,
    level: geoViewMode.value,
    regions: geoViewMode.value === GEO_LEVEL_PROVINCE ? geoProvinceEntries.value : geoCurrentEntriesState.value
  }));

  const geoCurrentEntries = computed(() => geoMapData.value?.regions || []);
  const geoSearchPlaceholder = computed(() => geoViewMode.value === GEO_LEVEL_CITY
    ? '搜索城市'
    : geoViewMode.value === GEO_LEVEL_DISTRICT
      ? '搜索区县'
      : '搜索省份');

  const geoFilteredEntries = computed(() => {
    const keyword = String(geoSearchKeyword.value || '').trim().toLowerCase();
    if (!keyword) return geoCurrentEntries.value;
    return geoCurrentEntries.value.filter((item) => String(item?.region_name || '').trim().toLowerCase().includes(keyword));
  });

  const geoCurrentRanking = computed(() => {
    const field = geoSelectedWindow.value;
    return [...geoFilteredEntries.value].sort((a, b) => {
      const diff = Number(b?.stats?.[field]?.count || 0) - Number(a?.stats?.[field]?.count || 0);
      if (diff !== 0) return diff;
      return String(a?.region_code || '').localeCompare(String(b?.region_code || ''));
    });
  });

  const geoCurrentWindowLabel = computed(() => geoWindowOptions.find((item) => item.value === geoSelectedWindow.value)?.label || '近7天');
  const geoMapTitle = computed(() => geoViewMode.value === GEO_LEVEL_CITY && geoSelectedProvinceName.value
    ? `${geoSelectedProvinceName.value}城市案件统计`
    : geoViewMode.value === GEO_LEVEL_DISTRICT && geoSelectedCityName.value
      ? `${geoSelectedCityName.value}区县案件统计`
      : '全国省级案件统计');
  const geoRankingTitle = computed(() => geoViewMode.value === GEO_LEVEL_CITY && geoSelectedProvinceName.value
    ? `${geoSelectedProvinceName.value}城市案件排行`
    : geoViewMode.value === GEO_LEVEL_DISTRICT && geoSelectedCityName.value
      ? `${geoSelectedCityName.value}区县案件排行`
      : '全国省级案件排行');

  const findProvinceByCode = (code) => (geoProvinceEntries.value || []).find((item) => item.region_code === code) || null;
  const findCityByCode = (code) => (geoCurrentEntriesState.value || []).find((item) => item.region_code === code) || null;

  const setGeoWindow = (value) => {
    geoSelectedWindow.value = value;
  };

  const geoRiskBadgeClass = (riskLevel) => {
    switch (String(riskLevel || '').trim()) {
      case '高':
        return 'bg-rose-400/15 text-rose-200 border border-rose-400/30';
      case '中':
        return 'bg-amber-400/15 text-amber-100 border border-amber-300/30';
      default:
        return 'bg-emerald-400/15 text-emerald-100 border border-emerald-300/30';
    }
  };

  const formatGeoChange = (value) => {
    const numeric = Number(value);
    if (!Number.isFinite(numeric)) return '--';
    const prefix = numeric > 0 ? '+' : '';
    return `${prefix}${(numeric * 100).toFixed(1)}%`;
  };

  const loadGeoChildren = async (level, parentCode, { parentName = '', forceRefresh = false } = {}) => {
    const normalizedLevel = level === GEO_LEVEL_DISTRICT ? GEO_LEVEL_DISTRICT : GEO_LEVEL_CITY;
    const normalizedParentCode = String(parentCode || '').trim();
    if (!normalizedParentCode) return [];

    const cacheKey = `${normalizedLevel}:${normalizedParentCode}`;
    if (!forceRefresh && geoChildCache[normalizedLevel].has(cacheKey)) {
      const cached = geoChildCache[normalizedLevel].get(cacheKey);
      geoCurrentEntriesState.value = Array.isArray(cached?.regions) ? cached.regions : [];
      if (normalizedLevel === GEO_LEVEL_CITY) {
        geoSelectedProvinceName.value = String(cached?.parent_name || parentName || geoSelectedProvinceName.value || '').trim();
      } else {
        geoSelectedCityName.value = String(cached?.parent_name || parentName || geoSelectedCityName.value || '').trim();
      }
      return geoCurrentEntriesState.value;
    }

    geoMapLoading.value = true;
    geoMapError.value = '';
    try {
      const res = await deps.request(
        `/scam/case-library/cases/geo-map/children?parent_code=${encodeURIComponent(normalizedParentCode)}&level=${encodeURIComponent(normalizedLevel)}`,
        'GET',
        null,
        { silent: true, throwOnError: true }
      );
      const nextList = Array.isArray(res?.regions) ? res.regions : [];
      geoChildCache[normalizedLevel].set(cacheKey, {
        parent_name: String(res?.parent_name || parentName || '').trim(),
        regions: nextList
      });
      geoCurrentEntriesState.value = nextList;
      if (normalizedLevel === GEO_LEVEL_CITY) {
        geoSelectedProvinceName.value = String(res?.parent_name || parentName || geoSelectedProvinceName.value || '').trim();
      } else {
        geoSelectedCityName.value = String(res?.parent_name || parentName || geoSelectedCityName.value || '').trim();
      }
      return nextList;
    } catch (error) {
      geoMapError.value = error?.message || '子级地区统计加载失败';
      deps.showToast(geoMapError.value, 'error');
      return [];
    } finally {
      geoMapLoading.value = false;
    }
  };

  const fetchGeoRiskMap = async (forceRefresh = false) => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    geoMapLoading.value = true;
    geoMapError.value = '';
    try {
      const res = await deps.request('/scam/case-library/cases/geo-map', 'GET', null, {
        silent: true,
        throwOnError: true
      });
      geoOverviewSummary.value = res?.summary || null;
      geoProvinceEntries.value = Array.isArray(res?.regions) ? res.regions : [];
      if (forceRefresh) {
        geoChildCache[GEO_LEVEL_CITY].clear();
        geoChildCache[GEO_LEVEL_DISTRICT].clear();
      }
      if (geoViewMode.value === GEO_LEVEL_PROVINCE) {
        geoCurrentEntriesState.value = [];
      } else if (geoViewMode.value === GEO_LEVEL_CITY && geoSelectedProvinceCode.value) {
        await loadGeoChildren(GEO_LEVEL_CITY, geoSelectedProvinceCode.value, {
          parentName: geoSelectedProvinceName.value,
          forceRefresh
        });
      } else if (geoViewMode.value === GEO_LEVEL_DISTRICT && geoSelectedCityCode.value) {
        await loadGeoChildren(GEO_LEVEL_DISTRICT, geoSelectedCityCode.value, {
          parentName: geoSelectedCityName.value,
          forceRefresh
        });
      }
      setTimeout(() => {
        renderGeoRiskMap().catch((error) => {
          deps.showToast(error?.message || '地图渲染失败', 'error');
        });
      }, 120);
    } catch (error) {
      geoMapError.value = error?.message || '全国地理统计加载失败';
      deps.showToast(geoMapError.value, 'error');
    } finally {
      geoMapLoading.value = false;
    }
  };

  const drillIntoProvince = async (provinceCode) => {
    const province = findProvinceByCode(provinceCode);
    if (!province) return;
    geoSelectedProvinceCode.value = province.region_code;
    geoSelectedProvinceName.value = province.region_name;
    geoSelectedCityCode.value = '';
    geoSelectedCityName.value = '';
    geoViewMode.value = GEO_LEVEL_CITY;
    await loadGeoChildren(GEO_LEVEL_CITY, province.region_code, { parentName: province.region_name });
  };

  const drillIntoCity = async (cityCode) => {
    const city = findCityByCode(cityCode);
    if (!city) return;
    geoSelectedCityCode.value = city.region_code;
    geoSelectedCityName.value = city.region_name;
    geoViewMode.value = GEO_LEVEL_DISTRICT;
    await loadGeoChildren(GEO_LEVEL_DISTRICT, city.region_code, { parentName: city.region_name });
  };

  const setGeoViewMode = async (value) => {
    if (value === GEO_LEVEL_CITY) {
      if (!geoSelectedProvinceCode.value) return;
      geoViewMode.value = GEO_LEVEL_CITY;
      await loadGeoChildren(GEO_LEVEL_CITY, geoSelectedProvinceCode.value, { parentName: geoSelectedProvinceName.value });
      return;
    }
    if (value === GEO_LEVEL_DISTRICT) {
      if (!geoSelectedCityCode.value) return;
      geoViewMode.value = GEO_LEVEL_DISTRICT;
      await loadGeoChildren(GEO_LEVEL_DISTRICT, geoSelectedCityCode.value, { parentName: geoSelectedCityName.value });
      return;
    }
    geoViewMode.value = GEO_LEVEL_PROVINCE;
    geoCurrentEntriesState.value = [];
  };

  const backToProvinceGeoMap = () => {
    geoSelectedCityCode.value = '';
    geoSelectedCityName.value = '';
    geoViewMode.value = GEO_LEVEL_PROVINCE;
    geoCurrentEntriesState.value = [];
  };

  const backToCityGeoMap = async () => {
    if (!geoSelectedProvinceCode.value) return;
    geoViewMode.value = GEO_LEVEL_CITY;
    await loadGeoChildren(GEO_LEVEL_CITY, geoSelectedProvinceCode.value, { parentName: geoSelectedProvinceName.value });
  };

  const openGeoRegionCases = async (region, options = {}) => {
    const regionCode = String(region?.region_code || region?.regionCode || '').trim();
    if (!regionCode) return;
    const requestedPage = Number(options.page);
    const requestedPageSize = Number(options.pageSize);
    if (Number.isFinite(requestedPage) && requestedPage > 0) {
      geoRegionCasesPage.value = requestedPage;
    }
    if (Number.isFinite(requestedPageSize) && requestedPageSize > 0) {
      geoRegionCasesPageSize.value = requestedPageSize;
    }
    geoRegionCasesVisible.value = true;
    geoRegionCasesLoading.value = true;
    geoRegionCasesError.value = '';
    try {
      const res = await deps.request(
        `/scam/case-library/cases/geo-map/region-cases?region_code=${encodeURIComponent(regionCode)}&window=${encodeURIComponent(geoSelectedWindow.value)}&page=${encodeURIComponent(geoRegionCasesPage.value)}&page_size=${encodeURIComponent(geoRegionCasesPageSize.value)}`,
        'GET',
        null,
        { silent: true, throwOnError: true }
      );
      geoRegionCasesData.value = res;
      geoRegionCasesPage.value = Number(res?.page) || geoRegionCasesPage.value;
      geoRegionCasesPageSize.value = Number(res?.page_size) || geoRegionCasesPageSize.value;
    } catch (error) {
      geoRegionCasesData.value = null;
      geoRegionCasesError.value = error?.message || '地区案件摘要加载失败';
      deps.showToast(geoRegionCasesError.value, 'error');
    } finally {
      geoRegionCasesLoading.value = false;
    }
  };

  const closeGeoRegionCases = () => {
    geoRegionCasesVisible.value = false;
  };

  const changeGeoRegionCasesPage = (page) => {
    const regionCode = String(geoRegionCasesData.value?.region_code || '').trim();
    if (!regionCode) return;
    openGeoRegionCases({ region_code: regionCode }, { page });
  };

  const changeGeoRegionCasesPageSize = (pageSize) => {
    const regionCode = String(geoRegionCasesData.value?.region_code || '').trim();
    if (!regionCode) return;
    openGeoRegionCases({ region_code: regionCode }, { page: 1, pageSize });
  };

  const loadGeoJSON = async (mapKey, boundaryCode) => {
    if (geoJSONCache.has(mapKey)) return geoJSONCache.get(mapKey);
    const geoJSON = await deps.request(buildGeoBoundaryPath(boundaryCode), 'GET', null, {
      silent: true,
      throwOnError: true
    });
    if (!geoJSON || typeof geoJSON !== 'object') {
      throw new Error('地图边界数据加载失败');
    }
    geoJSONCache.set(mapKey, geoJSON);
    return geoJSON;
  };

  const normalizeFeatureName = (name) => String(name || '').trim();

  const extractFeatureCenter = (feature) => {
    const properties = feature?.properties || {};
    const cp = Array.isArray(properties.cp) ? properties.cp : [];
    if (cp.length === 2) return cp;
    const center = Array.isArray(properties.center) ? properties.center : [];
    if (center.length === 2) return center;
    return null;
  };

  const buildScatterData = (features, entries) => {
    const centerMap = new Map(features.map((feature) => [normalizeFeatureName(feature?.properties?.name), extractFeatureCenter(feature)]));
    const field = geoSelectedWindow.value;
    return entries
      .map((item) => {
        const name = normalizeFeatureName(item.region_name);
        const center = centerMap.get(name);
        if (!center) return null;
        return {
          name,
          value: [...center, Number(item?.stats?.[field]?.count || 0)],
          payload: item
        };
      })
      .filter(Boolean);
  };

  const buildTooltipHTML = (payload) => {
    const field = geoSelectedWindow.value;
    const stats = payload?.stats?.[field];
    const topScams = Array.isArray(stats?.top_scam_types) ? stats.top_scam_types : [];
    const topScamText = topScams.length
      ? topScams.map((item) => `${item.scam_type} ${item.count}`).join(' / ')
      : '暂无高发骗局';
    return [
      `<div class="font-bold text-white text-sm mb-2">${payload.region_name}</div>`,
      `<div>案件数量：${stats?.count || 0}</div>`,
      `<div>风险等级：${stats?.risk_level || '低'}</div>`,
      `<div>高发骗局：${topScamText}</div>`,
      `<div>环比趋势：${stats?.trend || '持平'} (${formatGeoChange(stats?.change_rate || 0)})</div>`
    ].join('');
  };

  const renderGeoRiskMap = async () => {
    if (!geoCurrentEntries.value.length || typeof window.echarts === 'undefined') return;
    const dom = document.getElementById('adminGeoRiskMapChart');
    if (!dom) return;
    if (geoChartInstance?.dispose) geoChartInstance.dispose();
    geoChartInstance = echarts.init(dom);

    const mapKey = geoViewMode.value === GEO_LEVEL_DISTRICT && geoSelectedCityCode.value
      ? `city-${geoSelectedCityCode.value}`
      : geoViewMode.value === GEO_LEVEL_CITY && geoSelectedProvinceCode.value
        ? `province-${geoSelectedProvinceCode.value}`
        : 'china-country';
    const boundaryCode = geoViewMode.value === GEO_LEVEL_DISTRICT && geoSelectedCityCode.value
      ? geoSelectedCityCode.value
      : geoViewMode.value === GEO_LEVEL_CITY && geoSelectedProvinceCode.value
        ? geoSelectedProvinceCode.value
        : DEFAULT_GEO_BOUNDARY_CODE;
    const geoJSON = await loadGeoJSON(mapKey, boundaryCode);
    window.echarts.registerMap(mapKey, geoJSON);

    const entries = geoCurrentEntries.value;
    const features = Array.isArray(geoJSON?.features) ? geoJSON.features : [];
    const field = geoSelectedWindow.value;
    const maxCount = Math.max(1, ...entries.map((item) => Number(item?.stats?.[field]?.count || 0)));
    const mapData = entries.map((item) => ({
      name: item.region_name,
      value: Number(item?.stats?.[field]?.count || 0),
      payload: item,
      itemStyle: {
        areaColor: riskAreaColors[item?.stats?.[field]?.risk_level] || riskAreaColors['低']
      }
    }));
    const scatterData = buildScatterData(features, entries);

    geoChartInstance.setOption({
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(2, 6, 23, 0.94)',
        borderColor: 'rgba(56, 189, 248, 0.35)',
        borderWidth: 1,
        textStyle: { color: '#e2e8f0', fontSize: 12 },
        formatter: (params) => {
          const payload = params?.data?.payload;
          if (!payload) return params?.name || '';
          return buildTooltipHTML(payload);
        }
      },
      geo: {
        map: mapKey,
        roam: false,
        zoom: geoViewMode.value === GEO_LEVEL_CITY ? 1.02 : 1.08,
        itemStyle: {
          areaColor: 'rgba(15, 23, 42, 0.88)',
          borderColor: 'rgba(125, 211, 252, 0.25)',
          borderWidth: 1.1
        },
        emphasis: {
          label: { color: '#f8fafc' },
          itemStyle: {
            areaColor: 'rgba(34, 211, 238, 0.28)',
            borderColor: 'rgba(56, 189, 248, 0.8)'
          }
        }
      },
      series: [
        {
          name: '区域态势',
          type: 'map',
          map: mapKey,
          geoIndex: 0,
          data: mapData,
          silent: false
        },
        {
          name: '案件热度',
          type: 'scatter',
          coordinateSystem: 'geo',
          data: scatterData,
          silent: true,
          symbolSize: (value) => {
            const count = Array.isArray(value) ? Number(value[2] || 0) : 0;
            return Math.max(8, Math.min(24, 8 + (count / maxCount) * 16));
          },
          itemStyle: {
            color: (params) => riskColors[params?.data?.payload?.stats?.[field]?.risk_level] || riskColors['低'],
            shadowBlur: 18,
            shadowColor: 'rgba(56, 189, 248, 0.45)'
          },
          emphasis: { scale: 1.2 }
        }
      ]
    });

    geoChartInstance.off('click');
    geoChartInstance.on('click', (params) => {
      if (geoViewMode.value === GEO_LEVEL_PROVINCE) {
        const matched = findProvinceByCode(params?.data?.payload?.region_code) || geoProvinceEntries.value.find((item) => item.region_name === params.name);
        if (matched) {
          drillIntoProvince(matched.region_code);
        }
        return;
      }
      if (geoViewMode.value === GEO_LEVEL_CITY) {
        const matched = findCityByCode(params?.data?.payload?.region_code) || geoCurrentEntriesState.value.find((item) => item.region_name === params.name);
        if (matched) {
          drillIntoCity(matched.region_code);
        }
        return;
      }
      if (geoViewMode.value === GEO_LEVEL_DISTRICT) {
        const matched = geoCurrentEntriesState.value.find((item) => item.region_name === params.name);
        if (matched) {
          openGeoRegionCases(matched);
        }
      }
    });
  };

  const resizeGeoMap = () => {
    if (geoChartInstance?.resize) geoChartInstance.resize();
  };

  const disposeGeoMap = () => {
    if (geoChartInstance?.dispose) geoChartInstance.dispose();
    geoChartInstance = null;
  };

  watch(
    () => [geoCurrentEntries.value, geoSelectedWindow.value, geoViewMode.value, geoSelectedProvinceCode.value, geoSelectedCityCode.value],
    () => {
      if (!geoCurrentEntries.value.length) return;
      setTimeout(() => {
        renderGeoRiskMap().catch((error) => {
          deps.showToast(error?.message || '地图渲染失败', 'error');
        });
      }, 60);
    },
    { deep: false }
  );

  watch(
    () => geoViewMode.value,
    () => {
      geoSearchKeyword.value = '';
    }
  );

  watch(geoSelectedWindow, () => {
    const regionCode = String(geoRegionCasesData.value?.region_code || '').trim();
    if (!geoRegionCasesVisible.value || !regionCode) return;
    openGeoRegionCases({ region_code: regionCode }, { page: 1 });
  });

  watch(
    () => deps.activeTab?.value,
    (tab) => {
      if ((tab === 'geo_risk_map' || tab === 'geo_risk_map_full') && !geoProvinceEntries.value.length && !geoMapLoading.value) {
        fetchGeoRiskMap();
        return;
      }
      if (tab === 'geo_risk_map' || tab === 'geo_risk_map_full') {
        setTimeout(() => {
          renderGeoRiskMap().catch(() => {
            // ignore first-frame retry failures
          });
        }, 180);
      }
    },
    { immediate: true }
  );

  return {
    geoMapData,
    geoMapLoading,
    geoMapError,
    geoRegionCasesVisible,
    geoRegionCasesLoading,
    geoRegionCasesError,
    geoRegionCasesData,
    geoRegionCasesPage,
    geoRegionCasesPageSize,
    geoRegionCasesPageSizeOptions,
    geoSearchKeyword,
    geoSearchPlaceholder,
    geoSelectedWindow,
    geoWindowOptions,
    geoViewMode,
    geoSelectedProvinceCode,
    geoSelectedProvinceName,
    geoSelectedCityCode,
    geoSelectedCityName,
    geoCurrentRanking,
    geoCurrentWindowLabel,
    geoMapTitle,
    geoRankingTitle,
    setGeoWindow,
    setGeoViewMode,
    drillIntoProvince,
    drillIntoCity,
    backToProvinceGeoMap,
    backToCityGeoMap,
    formatGeoChange,
    geoRiskBadgeClass,
    openGeoRegionCases,
    closeGeoRegionCases,
    changeGeoRegionCasesPage,
    changeGeoRegionCasesPageSize,
    fetchGeoRiskMap,
    resizeGeoMap,
    disposeGeoMap
  };
}
