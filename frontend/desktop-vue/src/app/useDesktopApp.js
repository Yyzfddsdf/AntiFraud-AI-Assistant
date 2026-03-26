import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { createDesktopTabRouter } from '../modules/router/tabConfig';
import { createDesktopSessionHandlers } from '../modules/session/desktopSession';
import { createDesktopTabChangeHandler } from '../modules/tabs/desktopTabEffects';
import { useAlertsModule } from '../modules/alerts/useAlertsModule';
import { useFamilyModule } from '../modules/family/useFamilyModule';
import { useCaseLibraryModule } from '../modules/case-library/useCaseLibraryModule';
import { useChartsModule } from '../modules/charts/useChartsModule';
import { useGeoRiskMapModule } from '../modules/geo/useGeoRiskMapModule';
import { useChatModule } from '../modules/chat/useChatModule';

export function useDesktopApp() {
  const adminAnalyticsSectionKeys = new Set(['overview', 'graph', 'target_groups', 'profiles']);
  const familySectionKeysWithoutGroup = new Set(['received', 'create', 'join']);
  const familySectionKeysWithGroup = new Set(['overview', 'invite', 'guardian', 'members', 'records', 'notifications']);
  const taskCenterTabKeys = new Set(['submit', 'tasks', 'history', 'risk_trend']);
  const isAuthenticated = ref(false);
  const authReady = ref(false);
  const token = ref(localStorage.getItem('token') || '');
  const user = ref({});
  const authMode = ref('login');
  const loginMethod = ref('password');
  const activeTab = ref('tasks');
  const tabRouter = createDesktopTabRouter();
  let stopTabRouter = null;

  const loading = ref(false);
  const analyzing = ref(false);
  const inviteCode = ref('');
  const captchaImage = ref('');
  const captchaId = ref('');
  const toasts = ref([]);
  const tasks = ref([]);
  const history = ref([]);
  const users = ref([]);
  const selectedTask = ref(null);
  const userSearch = ref('');
  const deletingHistory = reactive({});
  const simulationGenerating = ref(false);
  const simulationSubmitting = ref(false);
  const simulationForm = reactive({
    caseType: '冒充客服',
    targetPersona: '普通居民',
    difficulty: 'medium',
    locale: 'zh-CN'
  });
  const simulationPackId = ref('');
  const simulationPack = ref(null);
  const simulationCurrentStep = ref(null);
  const simulationCurrentScore = ref(60);
  const simulationStatus = ref('idle');
  const simulationAnswers = ref([]);
  const simulationResult = ref(null);
  const simulationPackList = ref([]);
  const simulationSessionList = ref([]);

  const analyzeForm = reactive({
    text: '',
    videos: [],
    audios: [],
    images: []
  });

  const form = reactive({
    account: '',
    username: '',
    email: '',
    phone: '',
    password: '',
    captchaCode: '',
    smsCode: ''
  });

  const ageForm = reactive({ age: 28 });
  const profileForm = reactive({
    occupation: '',
    recentTagsText: '',
    provinceCode: '',
    provinceName: '',
    cityCode: '',
    cityName: '',
    districtCode: '',
    districtName: '',
    locationSource: 'manual'
  });
  const occupationOptions = ref([]);
  const provinceOptions = ref([]);
  const cityOptions = ref([]);
  const districtOptions = ref([]);
  const profileSaving = ref(false);
  const locationResolving = ref(false);
  const demoSMSCode = '000000';
  const smsCodeCooldown = ref(0);
  const smsCodeSending = ref(false);
  let smsCodeCooldownTimer = null;

  const chatPosition = reactive({ left: 0, top: 0 });
  const isDragging = ref(false);
  const hasMoved = ref(false);
  const isSidebarCollapsed = ref(false);
  const adminAnalyticsSection = ref('overview');
  const familyCenterSection = ref('received');
  const toggleSidebar = () => { isSidebarCollapsed.value = !isSidebarCollapsed.value; };
  const setAdminAnalyticsSection = (nextSection) => {
    const normalized = String(nextSection || '').trim().toLowerCase();
    adminAnalyticsSection.value = adminAnalyticsSectionKeys.has(normalized) ? normalized : 'overview';
  };
  const resolveFamilyCenterSection = (nextSection, hasGroup) => {
    const normalized = String(nextSection || '').trim().toLowerCase();
    if (hasGroup) {
      return familySectionKeysWithGroup.has(normalized) ? normalized : 'overview';
    }
    return familySectionKeysWithoutGroup.has(normalized) ? normalized : 'received';
  };
  const setFamilyCenterSection = (nextSection) => {
    familyCenterSection.value = resolveFamilyCenterSection(nextSection, familyModule?.familyHasGroup?.value);
  };
  const isTaskCenterTab = (tab) => taskCenterTabKeys.has(String(tab || '').trim());
  const openTaskCenter = () => {
    if (isTaskCenterTab(activeTab.value)) return;
    activeTab.value = 'tasks';
  };
  const setTaskCenterSection = (nextSection) => {
    const normalized = String(nextSection || '').trim().toLowerCase();
    activeTab.value = taskCenterTabKeys.has(normalized) ? normalized : 'tasks';
  };
  const taskCenterSection = computed(() => isTaskCenterTab(activeTab.value) ? activeTab.value : 'tasks');

  const showToast = (message, type = 'success') => {
    const id = Date.now();
    toasts.value.push({ id, message, type });
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id);
    }, 3000);
  };

  const stableJSONStringify = (value) => {
    try {
      return JSON.stringify(value);
    } catch {
      return '';
    }
  };

  const replaceListIfChanged = (targetRef, nextList) => {
    const normalized = Array.isArray(nextList) ? nextList : [];
    if (stableJSONStringify(targetRef.value) !== stableJSONStringify(normalized)) {
      targetRef.value = normalized;
    }
  };

  const parseRecentTagsInput = (raw) => {
    return String(raw || '')
      .split(/\r?\n|,|，|;|；/)
      .map(item => item.trim())
      .filter(Boolean);
  };

  const syncProfileForm = (profile) => {
    const normalizedAge = Number(profile && profile.age);
    ageForm.age = Number.isFinite(normalizedAge) && normalizedAge > 0 ? normalizedAge : 28;
    profileForm.occupation = String(profile?.occupation || '').trim();
    profileForm.provinceCode = String(profile?.province_code || '').trim();
    profileForm.provinceName = String(profile?.province_name || '').trim();
    profileForm.cityCode = String(profile?.city_code || '').trim();
    profileForm.cityName = String(profile?.city_name || '').trim();
    profileForm.districtCode = String(profile?.district_code || '').trim();
    profileForm.districtName = String(profile?.district_name || '').trim();
    profileForm.locationSource = String(profile?.location_source || '').trim() || 'manual';
    profileForm.recentTagsText = Array.isArray(profile?.recent_tags)
      ? profile.recent_tags.filter(item => typeof item === 'string' && item.trim()).join('\n')
      : '';
  };

  let logout = () => {};

  const request = async (path, method = 'GET', body = null, options = {}) => {
    const { silent = false, throwOnError = false } = options || {};
    const headers = { Accept: 'application/json', 'Content-Type': 'application/json' };
    if (token.value) headers.Authorization = `Bearer ${token.value}`;

    try {
      const res = await fetch(`/api${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined
      });

      if (res.status === 401 && isAuthenticated.value) {
        logout();
        return null;
      }

      const data = await res.json();
      if (!res.ok) {
        const error = new Error(data.error || data.message || 'Request failed');
        error.status = res.status;
        error.data = data;
        throw error;
      }
      return data;
    } catch (e) {
      if (!silent) {
        showToast(e.message, 'error');
      }
      if (throwOnError) throw e;
      return null;
    }
  };

  const fetchOccupationOptions = async () => {
    const res = await request('/user/profile/options/occupations', 'GET', null, { silent: true });
    if (res && Array.isArray(res.occupations)) {
      occupationOptions.value = res.occupations.filter(item => typeof item === 'string' && item.trim());
    }
  };

  const toOptionMap = (items) => new Map((Array.isArray(items) ? items : []).map(item => [String(item.code), item]));

  const fetchProvinceOptions = async () => {
    const res = await request('/regions/provinces', 'GET', null, { silent: true });
    provinceOptions.value = Array.isArray(res?.provinces) ? res.provinces : [];
  };

  const fetchCityOptions = async (provinceCode) => {
    if (!provinceCode) {
      cityOptions.value = [];
      return;
    }
    const res = await request(`/regions/cities?province_code=${encodeURIComponent(provinceCode)}`, 'GET', null, { silent: true });
    cityOptions.value = Array.isArray(res?.cities) ? res.cities : [];
  };

  const fetchDistrictOptions = async (cityCode) => {
    if (!cityCode) {
      districtOptions.value = [];
      return;
    }
    const res = await request(`/regions/districts?city_code=${encodeURIComponent(cityCode)}`, 'GET', null, { silent: true });
    districtOptions.value = Array.isArray(res?.districts) ? res.districts : [];
  };

  const applyRegionSelection = async (selection, source = 'manual') => {
    profileForm.provinceCode = String(selection?.province_code || '').trim();
    profileForm.provinceName = String(selection?.province_name || '').trim();
    await fetchCityOptions(profileForm.provinceCode);
    profileForm.cityCode = String(selection?.city_code || '').trim();
    profileForm.cityName = String(selection?.city_name || '').trim();
    await fetchDistrictOptions(profileForm.cityCode);
    profileForm.districtCode = String(selection?.district_code || '').trim();
    profileForm.districtName = String(selection?.district_name || '').trim();
    profileForm.locationSource = source;
  };

  const handleProvinceChange = async () => {
    const option = toOptionMap(provinceOptions.value).get(String(profileForm.provinceCode));
    profileForm.provinceName = option?.name || '';
    profileForm.cityCode = '';
    profileForm.cityName = '';
    profileForm.districtCode = '';
    profileForm.districtName = '';
    districtOptions.value = [];
    await fetchCityOptions(profileForm.provinceCode);
  };

  const handleCityChange = async () => {
    const option = toOptionMap(cityOptions.value).get(String(profileForm.cityCode));
    profileForm.cityName = option?.name || '';
    profileForm.districtCode = '';
    profileForm.districtName = '';
    await fetchDistrictOptions(profileForm.cityCode);
  };

  const handleDistrictChange = () => {
    const option = toOptionMap(districtOptions.value).get(String(profileForm.districtCode));
    profileForm.districtName = option?.name || '';
    profileForm.locationSource = 'manual';
  };

  const hydrateRegionOptionsFromProfile = async () => {
    if (profileForm.provinceCode) {
      await fetchCityOptions(profileForm.provinceCode);
    }
    if (profileForm.cityCode) {
      await fetchDistrictOptions(profileForm.cityCode);
    }
  };

  const requestCurrentRegion = async () => {
    if (!navigator.geolocation) {
      showToast('当前浏览器不支持定位', 'error');
      return;
    }
    locationResolving.value = true;
    try {
      const position = await new Promise((resolve, reject) => {
        navigator.geolocation.getCurrentPosition(resolve, reject, {
          enableHighAccuracy: true,
          timeout: 10000,
          maximumAge: 300000
        });
      });
      const lat = position?.coords?.latitude;
      const lng = position?.coords?.longitude;
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
        throw new Error('浏览器未返回有效定位坐标');
      }
      const geoRes = await fetch(`https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=${encodeURIComponent(lat)}&longitude=${encodeURIComponent(lng)}&localityLanguage=zh`);
      const geoData = await geoRes.json();
      if (!geoRes.ok) {
        throw new Error(geoData?.description || '当前位置解析失败');
      }
      const administrative = Array.isArray(geoData?.localityInfo?.administrative) ? geoData.localityInfo.administrative : [];
      const districtCandidates = administrative
        .map(item => String(item?.name || '').trim())
        .filter(name => name && name !== geoData.countryName && name !== geoData.principalSubdivision)
        .reverse();
      const districtName = String(districtCandidates[0] || geoData.locality || geoData.city || '').trim();
      const resolvePayload = {
        province_name: String(geoData?.principalSubdivision || '').trim(),
        city_name: String(geoData?.city || '').trim(),
        district_name: districtName,
        district_candidates: districtCandidates
      };
      const resolveRes = await request('/regions/resolve', 'POST', resolvePayload, { throwOnError: true });
      if (!resolveRes?.region) {
        throw new Error('当前位置未匹配到标准行政区');
      }
      await applyRegionSelection(resolveRes.region, 'auto');
      showToast('已自动识别当前位置');
    } catch (error) {
      showToast(error?.message || '定位获取失败', 'error');
    } finally {
      locationResolving.value = false;
    }
  };

  const requiresGraphCaptcha = computed(() => authMode.value === 'register' || loginMethod.value === 'password');
  const shouldShowSMSCodeSection = computed(() => authMode.value === 'register' || loginMethod.value === 'sms');
  const authSubmitLabel = computed(() => authMode.value === 'register' ? '创建账户' : (loginMethod.value === 'sms' ? '短信登录' : '立即登录'));
  const smsCodeButtonLabel = computed(() => authMode.value === 'register' ? '发送注册短信码' : '发送登录短信码');
  const smsCodeButtonText = computed(() => smsCodeCooldown.value > 0 ? `${smsCodeCooldown.value}s后重试` : smsCodeButtonLabel.value);
  const canSendSMSCode = computed(() => !smsCodeSending.value && smsCodeCooldown.value === 0);

  const getUserDisplayName = (userInfo) => String(userInfo?.username || '').trim() || String(userInfo?.email || '').trim() || String(userInfo?.phone || '').trim() || '未设置';
  const getUserEmailText = (userInfo) => String(userInfo?.email || '').trim();
  const getUserPhoneText = (userInfo) => String(userInfo?.phone || '').trim();
  const getUserAvatarText = (userInfo) => getUserDisplayName(userInfo).slice(0, 2).toUpperCase();

  const clearSMSCodeCooldownTimer = () => {
    if (smsCodeCooldownTimer) {
      clearInterval(smsCodeCooldownTimer);
      smsCodeCooldownTimer = null;
    }
  };

  const startSMSCodeCooldown = () => {
    clearSMSCodeCooldownTimer();
    smsCodeCooldown.value = 60;
    smsCodeCooldownTimer = setInterval(() => {
      if (smsCodeCooldown.value <= 1) {
        smsCodeCooldown.value = 0;
        clearSMSCodeCooldownTimer();
        return;
      }
      smsCodeCooldown.value -= 1;
    }, 1000);
  };

  const getRouteContext = () => ({
    isAuthenticated: isAuthenticated.value,
    userRole: user.value?.role || ''
  });

  const applyResolvedRoute = (resolved) => {
    if (!resolved || !resolved.isAuthenticated) return;
    if (resolved.activeTab !== activeTab.value) {
      activeTab.value = resolved.activeTab;
    }
  };

  const reconcileRouteState = ({ replace = false } = {}) => {
    applyResolvedRoute(tabRouter.reconcile(getRouteContext(), { replace }));
  };

  const syncRouteFromActiveTab = ({ replace = false } = {}) => {
    const resolvedTab = tabRouter.sync(getRouteContext(), activeTab.value, { replace });
    if (isAuthenticated.value && resolvedTab !== activeTab.value) {
      activeTab.value = resolvedTab;
    }
  };

  let fetchHistory = async () => {};
  let viewTaskDetail = async () => {};
  let fetchAdminStats = async () => {};

  const alertsModule = useAlertsModule({
    token,
    isAuthenticated,
    history,
    fetchHistory: (...args) => fetchHistory(...args),
    viewTaskDetail: (...args) => viewTaskDetail(...args),
    activeTab,
    showToast
  });

  const familyModule = useFamilyModule({
    isAuthenticated,
    token,
    request,
    replaceListIfChanged,
    showToast,
    activeTab
  });

  const chartsModule = useChartsModule({
    isAuthenticated,
    user,
    request,
    stableJSONStringify,
    showToast
  });
  const geoRiskMapModule = useGeoRiskMapModule({
    isAuthenticated,
    user,
    request,
    showToast
  });
  fetchAdminStats = chartsModule.fetchAdminStats;

  const caseLibraryModule = useCaseLibraryModule({
    isAuthenticated,
    user,
    request,
    replaceListIfChanged,
    showToast,
    refreshAdminStats: (...args) => fetchAdminStats(...args)
  });

  const fileToBase64 = (file) => new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result);
    reader.onerror = error => reject(error);
  });

  const chatModule = useChatModule({
    isAuthenticated,
    token,
    request,
    showToast,
    fileToBase64,
    selectedTask,
    hasMoved
  });

  const fetchCaptcha = async () => {
    try {
      const res = await fetch('/api/auth/captcha');
      const data = await res.json();
      captchaId.value = data.captchaId;
      captchaImage.value = data.captchaImage;
    } catch {
      showToast('验证码获取失败', 'error');
    }
  };

  const sendSMSCode = async () => {
    if (!canSendSMSCode.value) return;
    const phone = form.phone.trim();
    if (!phone) {
      showToast('请输入手机号', 'error');
      return;
    }
    smsCodeSending.value = true;
    const res = await request('/auth/sms-code', 'POST', { phone });
    smsCodeSending.value = false;
    if (res) {
      showToast(res.message || `短信验证码已发送，请使用 ${demoSMSCode}`);
      startSMSCodeCooldown();
    }
  };

  let startPolling = () => {};
  let stopPolling = () => {};

  const sessionHandlers = createDesktopSessionHandlers({
    form,
    request,
    showToast,
    fetchCaptcha,
    fetchOccupationOptions,
    syncProfileForm,
    startPolling: () => startPolling(),
    stopPolling: () => stopPolling(),
    reconcileRouteState,
    resetAlerts: alertsModule.resetAlertState,
    resetFamily: familyModule.resetFamilyState,
    authMode: () => authMode.value,
    setAuthMode: (value) => { authMode.value = value; },
    loginMethod: () => loginMethod.value,
    setLoginMethod: (value) => { loginMethod.value = value; },
    captchaId: () => captchaId.value,
    requiresGraphCaptcha: () => requiresGraphCaptcha.value,
    setLoading: (value) => { loading.value = value; },
    setToken: (value) => { token.value = value; },
    setAuthenticated: (value) => { isAuthenticated.value = value; },
    setUser: (value) => { user.value = value; }
  });

  const handleAuth = sessionHandlers.handleAuth;
  const getUserInfo = sessionHandlers.getUserInfo;
  logout = sessionHandlers.logout;

  const updateUserProfile = async () => {
    const normalizedAge = Number(ageForm.age);
    if (!Number.isFinite(normalizedAge) || normalizedAge < 1 || normalizedAge > 150) {
      showToast('年龄请输入 1 到 150 之间的数字', 'error');
      return;
    }
    profileSaving.value = true;
    const res = await request('/user/profile', 'PUT', {
      age: normalizedAge,
      occupation: profileForm.occupation,
      province_code: profileForm.provinceCode,
      province_name: profileForm.provinceName,
      city_code: profileForm.cityCode,
      city_name: profileForm.cityName,
      district_code: profileForm.districtCode,
      district_name: profileForm.districtName,
      location_source: String(profileForm.locationSource || '').trim() || 'manual'
    });
    profileSaving.value = false;
    if (res && res.user) {
      user.value = res.user;
      syncProfileForm(res.user);
      await chartsModule.fetchCurrentRegionCaseStats();
      showToast(res.message || '用户画像更新成功');
    }
  };
  const updateAge = updateUserProfile;

  const upgradeAccount = async () => {
    if (!inviteCode.value) return;
    loading.value = true;
    const res = await request('/upgrade', 'POST', { invite_code: inviteCode.value });
    loading.value = false;
    if (res) {
      showToast(res.message);
      user.value = res.user;
      inviteCode.value = '';
    }
  };

  watch([authMode, loginMethod], () => {
    if (requiresGraphCaptcha.value) {
      fetchCaptcha();
    } else {
      form.captchaCode = '';
    }
  });

  const deleteAccount = async () => {
    if (!confirm('确定要删除账户吗？此操作不可逆！')) return;
    const res = await request('/user', 'DELETE');
    if (res) {
      showToast('账户已删除');
      logout();
    }
  };

  const handleFileSelect = async (event, type) => {
    const files = Array.from(event.target.files);
    if (files.length === 0) return;
    try {
      const results = await Promise.all(files.map(file => fileToBase64(file)));
      analyzeForm[type] = [...analyzeForm[type], ...results];
      showToast(`已添加 ${files.length} 个文件`);
    } catch {
      showToast('文件读取失败', 'error');
    }
  };

  const submitAnalysis = async () => {
    if (!analyzeForm.text && analyzeForm.videos.length === 0 && analyzeForm.audios.length === 0 && analyzeForm.images.length === 0) {
      showToast('请至少输入文本或上传一个文件', 'warning');
      return;
    }
    analyzing.value = true;
    const res = await request('/scam/multimodal/analyze', 'POST', analyzeForm);
    analyzing.value = false;
    if (res && res.task_id) {
      showToast('分析任务已提交');
      analyzeForm.text = '';
      analyzeForm.videos = [];
      analyzeForm.audios = [];
      analyzeForm.images = [];
      activeTab.value = 'tasks';
      fetchTasks();
    }
  };

  const fetchTasks = async ({ silent = false } = {}) => {
    const res = await request('/scam/multimodal/tasks', 'GET', null, { silent });
    if (res && res.tasks) {
      replaceListIfChanged(tasks, res.tasks);
    }
  };

  fetchHistory = async ({ silent = false } = {}) => {
    const res = await request('/scam/multimodal/history', 'GET', null, { silent });
    if (res && res.history) {
      replaceListIfChanged(history, res.history);
    }
  };

  viewTaskDetail = async (taskId) => {
    const res = await request(`/scam/multimodal/tasks/${taskId}`);
    if (res && res.task) {
      selectedTask.value = res.task;
    }
  };

  const viewHistoryDetail = (item) => viewTaskDetail(item.record_id);

  const deleteHistoryCase = async (item) => {
    if (!item || !item.record_id) return;
    if (!confirm(`确定删除记录 ${item.title} 吗？`)) return;
    deletingHistory[item.record_id] = true;
    try {
      const res = await request(`/scam/multimodal/history/${item.record_id}`, 'DELETE');
      if (res) {
        showToast(res.message || '历史记录已删除');
        fetchHistory({ silent: true });
      }
    } finally {
      deletingHistory[item.record_id] = false;
    }
  };

  const resetSimulation = () => {
    simulationPackId.value = '';
    simulationPack.value = null;
    simulationCurrentStep.value = null;
    simulationCurrentScore.value = 60;
    simulationStatus.value = 'idle';
    simulationAnswers.value = [];
    simulationResult.value = null;
  };

  const fetchSimulationPacks = async () => {
    const res = await request('/scam/simulation/packs?limit=50', 'GET', null, { silent: true });
    simulationPackList.value = Array.isArray(res?.packs) ? res.packs : [];
  };

  const fetchSimulationSessions = async () => {
    const res = await request('/scam/simulation/sessions?limit=50', 'GET', null, { silent: true });
    simulationSessionList.value = Array.isArray(res?.sessions) ? res.sessions : [];
  };

  const deleteSimulationSession = async (sessionID) => {
    const id = String(sessionID || '').trim();
    if (!id) return;
    if (!confirm('确定删除该模拟报告吗？')) return;
    const res = await request(`/scam/simulation/sessions/${encodeURIComponent(id)}`, 'DELETE');
    if (!res) return;
    showToast(res.message || '报告删除成功');
    await fetchSimulationSessions();
    await fetchSimulationPacks();
  };

  const generateSimulationPack = async () => {
    simulationGenerating.value = true;
    try {
      const res = await request('/scam/simulation/packs/generate', 'POST', {
        case_type: simulationForm.caseType,
        target_persona: simulationForm.targetPersona,
        difficulty: simulationForm.difficulty,
        locale: simulationForm.locale
      });

      if (!res) return;
      showToast(res.message || '题目生成任务已提交，请稍候查看题目列表');

      const maxPoll = 30;
      let picked = false;
      for (let i = 0; i < maxPoll; i++) {
        await new Promise((resolve) => setTimeout(resolve, 15000));
        await fetchSimulationPacks();
        const firstPack = simulationPackList.value[0] || null;
        if (firstPack) {
          simulationPack.value = firstPack;
          simulationPackId.value = String(firstPack.pack_id || '').trim();
          simulationStatus.value = 'pack_ready';
          simulationCurrentScore.value = 60;
          simulationAnswers.value = [];
          simulationResult.value = null;
          picked = true;
          break;
        }
      }
      await fetchSimulationSessions();
      if (picked) {
        showToast('模拟题包生成完成');
      } else {
        showToast('题目仍在生成，请稍后刷新题目列表', 'warning');
      }
    } finally {
      simulationGenerating.value = false;
    }
  };

  const startSimulationSession = async (packIDOverride = '') => {
    const packID = String(packIDOverride || simulationPackId.value || simulationPack.value?.pack_id || '').trim();
    if (!packID) {
      showToast('请先生成题包', 'warning');
      return;
    }
    simulationSubmitting.value = true;
    try {
      const res = await request('/scam/simulation/sessions/answer', 'POST', {
        pack_id: packID
      });

      if (!res) {
        const statusRes = await request(`/scam/simulation/packs/${encodeURIComponent(packID)}/ongoing`, 'GET', null, { silent: true });
        if (statusRes && String(statusRes.status || '').trim() === 'in_progress') {
          simulationPackId.value = String(statusRes.pack_id || packID).trim();
          simulationPack.value = statusRes.pack || simulationPack.value;
          simulationCurrentStep.value = statusRes.next_step || null;
          simulationCurrentScore.value = Number(statusRes.current_score) || simulationCurrentScore.value;
          simulationStatus.value = 'in_progress';
          await fetchSimulationSessions();
        }
        return;
      }
      simulationPackId.value = packID;
      simulationStatus.value = String(res?.status || '').trim() || 'in_progress';
      simulationPack.value = res?.pack || simulationPack.value;
      simulationCurrentScore.value = Number(res?.current_score) || 60;
      const steps = Array.isArray(simulationPack.value?.steps) ? simulationPack.value.steps : [];
      simulationCurrentStep.value = steps.length > 0 ? steps[0] : null;
      simulationAnswers.value = [];
      simulationResult.value = null;
      await fetchSimulationSessions();
      showToast('答题会话已开始');
    } finally {
      simulationSubmitting.value = false;
    }
  };

  const submitSimulationAnswer = async (optionKey) => {
    if (!simulationPackId.value || !simulationCurrentStep.value || simulationStatus.value !== 'in_progress') return;
    simulationSubmitting.value = true;
    try {
      const res = await request('/scam/simulation/sessions/answer', 'POST', {
        pack_id: simulationPackId.value,
        step_id: simulationCurrentStep.value.step_id,
        option_key: optionKey
      });

      if (!res) return;
      simulationStatus.value = String(res.status || '').trim() || simulationStatus.value;
      simulationCurrentScore.value = Number(res?.current_score) || simulationCurrentScore.value;
      simulationCurrentStep.value = res?.next_step || null;
      simulationResult.value = res?.result || simulationResult.value;

      if (simulationStatus.value === 'completed') {
        await fetchSimulationSessions();
      const doneSession = simulationSessionList.value.find((item) => String(item.pack_id || '').trim() === simulationPackId.value);
      if (doneSession) {
        simulationAnswers.value = Array.isArray(doneSession.answers) ? doneSession.answers : simulationAnswers.value;
        simulationResult.value = doneSession.result || simulationResult.value;
        simulationPack.value = doneSession.pack || simulationPack.value;
      }
        await fetchSimulationPacks();
        showToast('模拟答题完成');
        return;
      }

      await fetchSimulationSessions();
      const progressingSession = simulationSessionList.value.find((item) => String(item.pack_id || '').trim() === simulationPackId.value);
      if (progressingSession && Array.isArray(progressingSession.answers)) {
        simulationAnswers.value = progressingSession.answers;
      }
    } finally {
      simulationSubmitting.value = false;
    }
  };

  const resumeOngoingSimulationSession = async () => {
    const packID = String(simulationPackId.value || simulationPack.value?.pack_id || '').trim();
    if (!packID) return false;
    const res = await request(`/scam/simulation/packs/${encodeURIComponent(packID)}/ongoing`, 'GET', null, { silent: true });
    if (!res) return false;
    simulationPackId.value = String(res.pack_id || '').trim();
    simulationPack.value = res.pack || simulationPack.value;
    simulationStatus.value = String(res.status || 'in_progress').trim();
    simulationCurrentScore.value = Number(res.current_score) || simulationCurrentScore.value;
    simulationCurrentStep.value = res.next_step || null;
    simulationResult.value = res.result || simulationResult.value;
    await fetchSimulationSessions();
    const ongoingSession = simulationSessionList.value.find((item) => String(item.pack_id || '').trim() === simulationPackId.value);
    if (ongoingSession && Array.isArray(ongoingSession.answers)) {
      simulationAnswers.value = ongoingSession.answers;
      simulationResult.value = ongoingSession.result || simulationResult.value;
      simulationPack.value = ongoingSession.pack || simulationPack.value;
    }
    return true;
  };

  const fetchUsers = async () => {
    if (!isAuthenticated.value || user.value.role !== 'admin') return;
    const res = await request('/users');
    if (res && res.users) {
      replaceListIfChanged(users, res.users);
    }
  };

  const debouncedFetchUsers = () => {
    fetchUsers();
  };

  const focusAdminTargetGroup = async (targetGroup) => {
    setAdminAnalyticsSection('target_groups');
    await chartsModule.fetchAdminTargetGroupChart(targetGroup);
  };

  const focusAdminProfile = (scamType) => {
    const normalized = String(scamType || '').trim();
    if (!normalized) return;
    setAdminAnalyticsSection('profiles');
    chartsModule.selectedGraphProfile.value = normalized;
    chartsModule.refreshVisibleAdminCharts();
  };

  const handleActiveTabChange = createDesktopTabChangeHandler({
    syncRouteFromActiveTab,
    fetchPendingReviews: caseLibraryModule.fetchPendingReviews,
    fetchCaseLibrary: caseLibraryModule.fetchCaseLibrary,
    fetchCaseOptionLists: caseLibraryModule.fetchCaseOptionLists,
    fetchFamilyOverview: familyModule.fetchFamilyOverview,
    familyHasGroup: () => familyModule.familyHasGroup.value,
    connectFamilyNotificationWebSocket: familyModule.connectFamilyNotificationWebSocket,
    fetchReceivedFamilyInvitations: familyModule.fetchReceivedFamilyInvitations,
    fetchUsers,
    fetchHistory,
    fetchTasks,
    fetchRiskTrend: chartsModule.fetchRiskTrend,
    fetchCurrentRegionCaseStats: chartsModule.fetchCurrentRegionCaseStats,
    fetchAdminStats,
    fetchGeoRiskMap: geoRiskMapModule.fetchGeoRiskMap,
    resetSimulation,
    fetchSimulationPacks,
    fetchSimulationSessions,
    resumeOngoingSimulationSession
  });
  watch(activeTab, handleActiveTabChange);
  watch([activeTab, adminAnalyticsSection], ([currentTab]) => {
    if (currentTab !== 'admin_stats') return;
    chartsModule.refreshVisibleAdminCharts();
  });
  watch(() => familyModule.familyHasGroup.value, (hasGroup) => {
    familyCenterSection.value = resolveFamilyCenterSection(familyCenterSection.value, hasGroup);
  }, { immediate: true });

  let pollInterval = null;
  startPolling = () => {
    fetchTasks({ silent: true });
    fetchHistory({ silent: true });
    chartsModule.fetchCurrentRegionCaseStats();
    familyModule.fetchFamilyOverview({ silent: true });
    alertsModule.connectAlertWebSocket();
    familyModule.connectFamilyNotificationWebSocket();
    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(() => {
      if (isAuthenticated.value && activeTab.value === 'tasks') fetchTasks({ silent: true });
      if (isAuthenticated.value && activeTab.value === 'family') {
        familyModule.fetchFamilyOverview({ silent: true }).then(() => {
          if (!familyModule.familyHasGroup.value) {
            familyModule.fetchReceivedFamilyInvitations({ silent: true });
          }
        });
      }
    }, 5000);
  };

  stopPolling = () => {
    if (pollInterval) clearInterval(pollInterval);
    alertsModule.disconnectAlertWebSocket();
    familyModule.disconnectFamilyNotificationWebSocket();
  };

  const initChatPosition = () => {
    chatPosition.left = window.innerWidth - 100;
    chatPosition.top = window.innerHeight - 100;
  };

  const handleResize = () => {
    const maxX = window.innerWidth - 60;
    const maxY = window.innerHeight - 60;
    if (chatPosition.left > maxX) chatPosition.left = maxX;
    if (chatPosition.top > maxY) chatPosition.top = maxY;
  };

  const resizeChartsOnWindow = () => {
    chartsModule.resizeCharts();
    geoRiskMapModule.resizeGeoMap();
  };

  const startDrag = (e) => {
    if (e.button !== 0) return;
    isDragging.value = true;
    hasMoved.value = false;

    const startX = e.clientX;
    const startY = e.clientY;
    const initialLeft = chatPosition.left;
    const initialTop = chatPosition.top;

    const onMouseMove = (moveEvent) => {
      const dx = moveEvent.clientX - startX;
      const dy = moveEvent.clientY - startY;
      if (Math.abs(dx) > 5 || Math.abs(dy) > 5) {
        hasMoved.value = true;
      }
      let newLeft = initialLeft + dx;
      let newTop = initialTop + dy;
      const maxX = window.innerWidth - 60;
      const maxY = window.innerHeight - 60;
      if (newLeft < 0) newLeft = 0;
      if (newLeft > maxX) newLeft = maxX;
      if (newTop < 0) newTop = 0;
      if (newTop > maxY) newTop = maxY;
      chatPosition.left = newLeft;
      chatPosition.top = newTop;
    };

    const onMouseUp = () => {
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
      isDragging.value = false;
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  };

  onMounted(async () => {
    stopTabRouter = tabRouter.mount({ getContext: getRouteContext, onResolve: applyResolvedRoute });
    fetchProvinceOptions();
    fetchCaptcha();
    if (token.value) {
      await getUserInfo();
      await hydrateRegionOptionsFromProfile();
      await chartsModule.fetchCurrentRegionCaseStats();
      await resumeOngoingSimulationSession();
    } else {
      reconcileRouteState({ replace: true });
    }
    authReady.value = true;
    initChatPosition();
    window.addEventListener('resize', handleResize);
    window.addEventListener('resize', resizeChartsOnWindow);
  });

  onUnmounted(() => {
    stopPolling();
    clearSMSCodeCooldownTimer();
    if (stopTabRouter) stopTabRouter();
    chartsModule.disposeCharts();
    geoRiskMapModule.disposeGeoMap();
    window.removeEventListener('resize', handleResize);
    window.removeEventListener('resize', resizeChartsOnWindow);
  });

  const formatTime = (iso) => new Date(iso).toLocaleString('zh-CN', { hour12: false });
  const getStatusLabel = (status) => ({ pending: '等待中', processing: '分析中', completed: '已完成', failed: '失败' }[status] || status);
  const getStatusClass = (status) => ({
    pending: 'bg-yellow-100 text-yellow-800 px-2 py-1 rounded-full text-xs font-bold',
    processing: 'bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs font-bold',
    completed: 'bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-bold',
    failed: 'bg-red-100 text-red-800 px-2 py-1 rounded-full text-xs font-bold'
  }[status] || 'bg-gray-100 text-gray-800 px-2 py-1 rounded-full text-xs font-bold');

  const normalizeRiskLevelText = (level) => {
    const value = String(level || '').trim();
    if (value === '高') return '高';
    if (value === '低') return '低';
    return '中';
  };
  const getRiskClass = (level) => {
    const normalized = normalizeRiskLevelText(level);
    if (normalized === '高') return 'bg-red-100 text-red-800';
    if (normalized === '中') return 'bg-yellow-100 text-yellow-800';
    return 'bg-green-100 text-green-800';
  };

  const openImage = (src) => {
    const win = window.open('', '_blank');
    win.document.write(`<img src="${src}" style="max-width:100%; height:auto;">`);
  };

  return {
    isAuthenticated,
    authReady,
    user,
    authMode,
    loginMethod,
    form,
    ageForm,
    profileForm,
    occupationOptions,
    provinceOptions,
    cityOptions,
    districtOptions,
    profileSaving,
    locationResolving,
    analyzeForm,
    captchaImage,
    requiresGraphCaptcha,
    shouldShowSMSCodeSection,
    authSubmitLabel,
    smsCodeButtonText,
    canSendSMSCode,
    demoSMSCode,
    fetchCaptcha,
    sendSMSCode,
    handleAuth,
    logout,
    loading,
    activeTab,
    simulationGenerating,
    simulationSubmitting,
    simulationForm,
    simulationPack,
    simulationCurrentStep,
    simulationCurrentScore,
    simulationStatus,
    simulationAnswers,
    simulationResult,
    simulationPackList,
    simulationSessionList,
    tasks,
    history,
    users,
    selectedTask,
    userSearch,
    toasts,
    analyzing,
    deletingHistory,
    handleFileSelect,
    submitAnalysis,
    viewTaskDetail,
    viewHistoryDetail,
    deleteHistoryCase,
    debouncedFetchUsers,
    formatTime,
    getStatusLabel,
    getStatusClass,
    normalizeRiskLevelText,
    getRiskClass,
    updateAge,
    updateUserProfile,
    requestCurrentRegion,
    handleProvinceChange,
    handleCityChange,
    handleDistrictChange,
    deleteAccount,
    upgradeAccount,
    inviteCode,
    openImage,
    generateSimulationPack,
    startSimulationSession,
    submitSimulationAnswer,
    resetSimulation,
    fetchSimulationPacks,
    fetchSimulationSessions,
    resumeOngoingSimulationSession,
    deleteSimulationSession,
    getUserDisplayName,
    getUserEmailText,
    getUserPhoneText,
    getUserAvatarText,
    chatPosition,
    startDrag,
    isSidebarCollapsed,
    toggleSidebar,
    adminAnalyticsSection,
    setAdminAnalyticsSection,
    taskCenterSection,
    isTaskCenterTab,
    openTaskCenter,
    setTaskCenterSection,
    familyCenterSection,
    setFamilyCenterSection,
    focusAdminTargetGroup,
    focusAdminProfile,
    ...alertsModule,
    ...familyModule,
    ...caseLibraryModule,
    ...chartsModule,
    ...geoRiskMapModule,
    ...chatModule
  };
}
