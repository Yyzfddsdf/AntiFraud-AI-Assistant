import { computed, ref, watch } from 'vue';

const CHINA_MAP_URL = 'https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json';
const PROVINCE_MAP_URL = (code) => `https://geo.datav.aliyun.com/areas_v3/bound/${code}_full.json`;

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
  const geoMapData = ref(null);
  const geoMapLoading = ref(false);
  const geoMapError = ref('');
  const geoSelectedWindow = ref('last_7d');
  const geoViewMode = ref('province');
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

  const geoMapCache = ref(null);
  const geoJSONCache = new Map();
  let geoChartInstance = null;

  const geoSelectedProvince = computed(() => {
    if (!geoMapData.value?.provinces || !geoSelectedProvinceCode.value) return null;
    return geoMapData.value.provinces.find((item) => item.region_code === geoSelectedProvinceCode.value) || null;
  });

  const geoCurrentEntries = computed(() => {
    if (!geoMapData.value?.provinces) return [];
    if (geoViewMode.value === 'district' && geoSelectedProvince.value && geoSelectedCityCode.value) {
      const city = (geoSelectedProvince.value.cities || []).find((item) => item.region_code === geoSelectedCityCode.value);
      return Array.isArray(city?.districts) ? city.districts : [];
    }
    if (geoViewMode.value === 'city' && geoSelectedProvince.value) {
      return Array.isArray(geoSelectedProvince.value.cities) ? geoSelectedProvince.value.cities : [];
    }
    return geoMapData.value.provinces;
  });

  const geoCurrentRanking = computed(() => {
    const field = geoSelectedWindow.value;
    return [...geoCurrentEntries.value].sort((a, b) => {
      const diff = Number(b?.stats?.[field]?.count || 0) - Number(a?.stats?.[field]?.count || 0);
      if (diff !== 0) return diff;
      return String(a?.region_code || '').localeCompare(String(b?.region_code || ''));
    }).slice(0, 12);
  });

  const geoCurrentWindowLabel = computed(() => geoWindowOptions.find((item) => item.value === geoSelectedWindow.value)?.label || '近7天');
  const geoMapTitle = computed(() => geoViewMode.value === 'city' && geoSelectedProvinceName.value
    ? `${geoSelectedProvinceName.value}城市态势图`
    : geoViewMode.value === 'district' && geoSelectedCityName.value
      ? `${geoSelectedCityName.value}县区态势图`
    : '全国省级态势图');
  const geoRankingTitle = computed(() => geoViewMode.value === 'city' && geoSelectedProvinceName.value
    ? `${geoSelectedProvinceName.value}高风险城市排行`
    : geoViewMode.value === 'district' && geoSelectedCityName.value
      ? `${geoSelectedCityName.value}高风险县区排行`
    : '全国省级风险排行');

  const setGeoWindow = (value) => {
    geoSelectedWindow.value = value;
  };

  const setGeoViewMode = (value) => {
    if (value === 'city' && !geoSelectedProvinceCode.value) return;
    if (value === 'district' && !geoSelectedCityCode.value) return;
    geoViewMode.value = value;
  };

  const drillIntoProvince = (provinceCode) => {
    const province = geoMapData.value?.provinces?.find((item) => item.region_code === provinceCode);
    if (!province) return;
    geoSelectedProvinceCode.value = province.region_code;
    geoSelectedProvinceName.value = province.region_name;
    geoSelectedCityCode.value = '';
    geoSelectedCityName.value = '';
    geoViewMode.value = 'city';
  };

  const drillIntoCity = (cityCode) => {
    if (!geoSelectedProvince.value) return;
    const city = (geoSelectedProvince.value.cities || []).find((item) => item.region_code === cityCode);
    if (!city) return;
    geoSelectedCityCode.value = city.region_code;
    geoSelectedCityName.value = city.region_name;
    geoViewMode.value = 'district';
  };

  const backToProvinceGeoMap = () => {
    geoSelectedCityCode.value = '';
    geoSelectedCityName.value = '';
    geoViewMode.value = 'province';
  };

  const backToCityGeoMap = () => {
    geoViewMode.value = 'city';
  };

  const formatGeoChange = (value) => {
    const numeric = Number(value);
    if (!Number.isFinite(numeric)) return '--';
    const prefix = numeric > 0 ? '+' : '';
    return `${prefix}${(numeric * 100).toFixed(1)}%`;
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

  const fetchGeoRiskMap = async (forceRefresh = false) => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    if (geoMapCache.value && !forceRefresh) {
      geoMapData.value = geoMapCache.value;
      setTimeout(() => {
        renderGeoRiskMap().catch((error) => {
          deps.showToast(error?.message || '地图渲染失败', 'error');
        });
      }, 120);
      return;
    }
    geoMapLoading.value = true;
    geoMapError.value = '';
    try {
      const res = await deps.request('/scam/case-library/cases/geo-map', 'GET', null, { silent: true });
      if (res) {
        geoMapData.value = res;
        geoMapCache.value = res;
        if (!geoSelectedProvinceCode.value && Array.isArray(res.provinces) && res.provinces.length > 0) {
          geoSelectedProvinceCode.value = res.provinces[0].region_code;
          geoSelectedProvinceName.value = res.provinces[0].region_name;
        }
        setTimeout(() => {
          renderGeoRiskMap().catch((error) => {
            deps.showToast(error?.message || '地图渲染失败', 'error');
          });
        }, 120);
      }
    } catch (error) {
      geoMapError.value = error?.message || '全国地理统计加载失败';
      deps.showToast(geoMapError.value, 'error');
    } finally {
      geoMapLoading.value = false;
    }
  };

  const loadGeoJSON = async (mapKey, url) => {
    if (geoJSONCache.has(mapKey)) return geoJSONCache.get(mapKey);
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error('地图边界数据加载失败');
    }
    const geoJSON = await response.json();
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
    if (!geoMapData.value || typeof window.echarts === 'undefined') return;
    const dom = document.getElementById('adminGeoRiskMapChart');
    if (!dom) return;
    if (geoChartInstance?.dispose) geoChartInstance.dispose();
    geoChartInstance = echarts.init(dom);

    const mapKey = geoViewMode.value === 'district' && geoSelectedCityCode.value
      ? `city-${geoSelectedCityCode.value}`
      : geoViewMode.value === 'city' && geoSelectedProvinceCode.value
        ? `province-${geoSelectedProvinceCode.value}`
        : 'china-country';
    const mapURL = geoViewMode.value === 'district' && geoSelectedCityCode.value
      ? PROVINCE_MAP_URL(geoSelectedCityCode.value)
      : geoViewMode.value === 'city' && geoSelectedProvinceCode.value
        ? PROVINCE_MAP_URL(geoSelectedProvinceCode.value)
        : CHINA_MAP_URL;
    const geoJSON = await loadGeoJSON(mapKey, mapURL);
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
        zoom: geoViewMode.value === 'city' ? 1.02 : 1.08,
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
      if (geoViewMode.value === 'province') {
        const matched = geoMapData.value?.provinces?.find((item) => item.region_name === params.name);
        if (matched) {
          drillIntoProvince(matched.region_code);
        }
        return;
      }
      if (geoViewMode.value === 'city' && geoSelectedProvince.value) {
        const matched = (geoSelectedProvince.value.cities || []).find((item) => item.region_name === params.name);
        if (matched) {
          drillIntoCity(matched.region_code);
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
    () => [geoMapData.value, geoSelectedWindow.value, geoViewMode.value, geoSelectedProvinceCode.value, geoSelectedCityCode.value],
    () => {
      if (!geoMapData.value) return;
      setTimeout(() => {
        renderGeoRiskMap().catch((error) => {
          deps.showToast(error?.message || '地图渲染失败', 'error');
        });
      }, 60);
    },
    { deep: false }
  );

  watch(
    () => deps.activeTab?.value,
    (tab) => {
      if ((tab === 'geo_risk_map' || tab === 'geo_risk_map_full') && !geoMapData.value && !geoMapLoading.value) {
        fetchGeoRiskMap();
        return;
      }
      if (tab === 'geo_risk_map' || tab === 'geo_risk_map_full') {
        setTimeout(() => {
          renderGeoRiskMap().catch(() => {
            // ignore first-frame retry failures; data/watchers will retry later
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
    fetchGeoRiskMap,
    resizeGeoMap,
    disposeGeoMap
  };
}
