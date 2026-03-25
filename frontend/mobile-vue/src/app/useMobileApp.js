import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { createMobileTabRouter } from '../modules/router/tabConfig';
import { createMobileSessionHandlers } from '../modules/session/mobileSession';
import { createMobileTabChangeHandler } from '../modules/tabs/mobileTabEffects';
import { useTasksModule } from '../modules/tasks/useTasksModule';
import { useAlertsModule } from '../modules/alerts/useAlertsModule';
import { useFamilyModule } from '../modules/family/useFamilyModule';
import { useChartsModule } from '../modules/charts/useChartsModule';
import { useChatModule } from '../modules/chat/useChatModule';

export function useMobileApp() {
  const isAuthenticated = ref(false);
  const authReady = ref(false);
  const token = ref(localStorage.getItem('token') || '');
  const user = ref({});
  const authMode = ref('login');
  const loginMethod = ref('password');
  const activeTab = ref('tasks');
  const tabRouter = createMobileTabRouter();
  let stopTabRouter = null;

  const loading = ref(false);
  const ageEditorVisible = ref(false);
  const captchaImage = ref('');
  const captchaId = ref('');
  const toasts = ref([]);
  const openDropdownKey = ref('');
  const currentBannerIndex = ref(0);
  const simulationGenerating = ref(false);
  const simulationSubmitting = ref(false);
  const simulationForm = reactive({
    caseType: '冒充客服',
    targetPersona: '普通居民',
    difficulty: 'medium',
    locale: 'zh-CN'
  });
  const simulationPack = ref(null);
  const simulationPackId = ref('');
  const simulationCurrentStep = ref(null);
  const simulationCurrentScore = ref(60);
  const simulationStatus = ref('idle');
  const simulationViewMode = ref('overview');
  const simulationAnswers = ref([]);
  const simulationResult = ref(null);
  const simulationPackList = ref([]);
  const simulationSessionList = ref([]);

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
  let bannerCarouselTimer = null;

  const scrollContainerToTop = (selector) => {
    const element = document.querySelector(selector);
    if (element instanceof HTMLElement) {
      element.scrollTo({ top: 0, behavior: 'auto' });
    }
  };

  const scrollMainToTop = () => {
    requestAnimationFrame(() => {
      [
        'main',
        '[data-mobile-scroll="submit-page"]',
        '[data-mobile-scroll="simulation-overview"]',
        '[data-mobile-scroll="simulation-exam"]'
      ].forEach(scrollContainerToTop);
    });
  };

  const showToast = (message, type = 'success') => {
    const id = Date.now();
    toasts.value.push({ id, message, type });
    setTimeout(() => {
      toasts.value = toasts.value.filter((item) => item.id !== id);
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

  const syncProfileForm = (profile) => {
    const normalizedAge = Number(profile && profile.age);
    ageForm.age = Number.isFinite(normalizedAge) && normalizedAge > 0 ? normalizedAge : 28;
    profileForm.occupation = String(profile && profile.occupation ? profile.occupation : '').trim();
    profileForm.provinceCode = String(profile?.province_code || '').trim();
    profileForm.provinceName = String(profile?.province_name || '').trim();
    profileForm.cityCode = String(profile?.city_code || '').trim();
    profileForm.cityName = String(profile?.city_name || '').trim();
    profileForm.districtCode = String(profile?.district_code || '').trim();
    profileForm.districtName = String(profile?.district_name || '').trim();
    profileForm.locationSource = String(profile?.location_source || '').trim() || 'manual';
    profileForm.recentTagsText = Array.isArray(profile && profile.recent_tags)
      ? profile.recent_tags.filter((item) => typeof item === 'string' && item.trim()).join('\n')
      : '';
  };

  let logout = () => {};

  const request = async (path, method = 'GET', body = null, options = {}) => {
    const { silent = false } = options || {};
    const headers = {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    };
    if (token.value) {
      headers.Authorization = `Bearer ${token.value}`;
    }

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
        throw new Error(data.error || data.message || 'Request failed');
      }
      return data;
    } catch (error) {
      if (!silent) {
        showToast(error.message, 'error');
      }
      return null;
    }
  };

  const resetSimulation = () => {
    simulationPack.value = null;
    simulationPackId.value = '';
    simulationCurrentStep.value = null;
    simulationCurrentScore.value = 60;
    simulationStatus.value = 'idle';
    simulationViewMode.value = 'overview';
    simulationAnswers.value = [];
    simulationResult.value = null;
  };

  const openSimulationExamView = () => {
    simulationViewMode.value = 'exam';
    requestAnimationFrame(() => {
      scrollContainerToTop('[data-mobile-scroll="simulation-exam"]');
    });
  };

  const closeSimulationExamView = () => {
    simulationViewMode.value = 'overview';
    requestAnimationFrame(() => {
      scrollContainerToTop('[data-mobile-scroll="simulation-overview"]');
    });
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
    if (!window.confirm('确定删除该报告吗？')) return;
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
      showToast('请先生成题包', 'error');
      return;
    }
    simulationSubmitting.value = true;
    try {
      const res = await request('/scam/simulation/sessions/answer', 'POST', { pack_id: packID });
      if (!res) {
        const statusRes = await request(`/scam/simulation/packs/${encodeURIComponent(packID)}/ongoing`, 'GET', null, { silent: true });
        if (statusRes && String(statusRes.status || '').trim() === 'in_progress') {
          simulationPackId.value = String(statusRes.pack_id || packID).trim();
          simulationPack.value = statusRes.pack || simulationPack.value;
          simulationCurrentStep.value = statusRes.next_step || null;
          simulationCurrentScore.value = Number(statusRes.current_score) || simulationCurrentScore.value;
          simulationStatus.value = 'in_progress';
          openSimulationExamView();
          await fetchSimulationSessions();
        }
        return;
      }
      simulationPackId.value = packID;
      simulationStatus.value = String(res.status || 'in_progress').trim();
      simulationCurrentScore.value = Number(res.current_score) || 60;
      simulationPack.value = res.pack || simulationPack.value;
      const steps = Array.isArray(simulationPack.value?.steps) ? simulationPack.value.steps : [];
      simulationCurrentStep.value = steps.length ? steps[0] : null;
      openSimulationExamView();
      simulationAnswers.value = [];
      simulationResult.value = null;
      await fetchSimulationSessions();
      showToast('答题已开始');
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

      simulationStatus.value = String(res.status || simulationStatus.value).trim();
      simulationCurrentScore.value = Number(res.current_score) || simulationCurrentScore.value;
      simulationCurrentStep.value = res.next_step || null;
      simulationResult.value = res.result || simulationResult.value;
      openSimulationExamView();

      await fetchSimulationSessions();
      const currentSession = simulationSessionList.value.find((item) => String(item.pack_id || '').trim() === simulationPackId.value);
      if (currentSession && Array.isArray(currentSession.answers)) {
        simulationAnswers.value = currentSession.answers;
        simulationResult.value = currentSession.result || simulationResult.value;
        simulationPack.value = currentSession.pack || simulationPack.value;
      }
      if (simulationStatus.value === 'completed') {
        await fetchSimulationSessions();
        await fetchSimulationPacks();
        showToast('模拟答题完成');
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
    openSimulationExamView();
    await fetchSimulationSessions();
    const ongoingSession = simulationSessionList.value.find((item) => String(item.pack_id || '').trim() === simulationPackId.value);
    if (ongoingSession && Array.isArray(ongoingSession.answers)) {
      simulationAnswers.value = ongoingSession.answers;
      simulationResult.value = ongoingSession.result || simulationResult.value;
      simulationPack.value = ongoingSession.pack || simulationPack.value;
    }
    return true;
  };

  const fetchOccupationOptions = async () => {
    const res = await request('/user/profile/options/occupations', 'GET', null, { silent: true });
    if (res && Array.isArray(res.occupations)) {
      occupationOptions.value = res.occupations.filter((item) => typeof item === 'string' && item.trim());
    }
  };

  const toOptionMap = (items) => new Map((Array.isArray(items) ? items : []).map((item) => [String(item.code), item]));

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

  const selectProvinceValue = async (code) => {
    profileForm.provinceCode = String(code || '').trim();
    await handleProvinceChange();
    closeDropdown();
  };

  const selectCityValue = async (code) => {
    profileForm.cityCode = String(code || '').trim();
    await handleCityChange();
    closeDropdown();
  };

  const selectDistrictValue = (code) => {
    profileForm.districtCode = String(code || '').trim();
    handleDistrictChange();
    closeDropdown();
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
        .map((item) => String(item?.name || '').trim())
        .filter((name) => name && name !== geoData.countryName && name !== geoData.principalSubdivision)
        .reverse();
      const districtName = String(districtCandidates[0] || geoData.locality || geoData.city || '').trim();
      const resolvePayload = {
        province_name: String(geoData?.principalSubdivision || '').trim(),
        city_name: String(geoData?.city || '').trim(),
        district_name: districtName,
        district_candidates: districtCandidates
      };
      const resolveRes = await request('/regions/resolve', 'POST', resolvePayload);
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
  const authSubmitLabel = computed(() => {
    if (authMode.value === 'register') return '创建账户';
    return loginMethod.value === 'sms' ? '短信登录' : '立即登录';
  });
  const smsCodeButtonLabel = computed(() => authMode.value === 'register' ? '发送注册短信码' : '发送登录短信码');
  const smsCodeButtonText = computed(() => smsCodeCooldown.value > 0 ? `${smsCodeCooldown.value}s后重试` : smsCodeButtonLabel.value);
  const canSendSMSCode = computed(() => !smsCodeSending.value && smsCodeCooldown.value === 0);

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

  const fileToBase64 = (file) => new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result);
    reader.onerror = (error) => reject(error);
  });

  const tasksModule = useTasksModule({
    isAuthenticated,
    activeTab,
    request,
    replaceListIfChanged,
    showToast,
    fileToBase64
  });

  const alertsModule = useAlertsModule({
    token,
    isAuthenticated,
    history: tasksModule.history,
    fetchHistory: (...args) => tasksModule.fetchHistory(...args),
    viewTaskDetail: (...args) => tasksModule.viewTaskDetail(...args),
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
    history: tasksModule.history,
    request,
    stableJSONStringify,
    showToast
  });

  const chatModule = useChatModule({
    token,
    isAuthenticated,
    request,
    showToast,
    fileToBase64,
    renderMarkdown: tasksModule.renderMarkdown
  });

  let startPolling = () => {};
  let stopPolling = () => {};

  const sessionHandlers = createMobileSessionHandlers({
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
    resetChat: chatModule.resetChatState,
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
      ageEditorVisible.value = false;
      showToast(res.message || '用户画像更新成功');
    }
  };

  const updateAge = updateUserProfile;

  const deleteAccount = async () => {
    if (!confirm('确定要删除账户吗？此操作不可逆！')) return;
    const res = await request('/user', 'DELETE');
    if (res) {
      showToast('账户已删除');
      logout();
    }
  };

  const getUserDisplayName = (userInfo) => {
    const candidate = userInfo && typeof userInfo === 'object' ? userInfo : {};
    return String(candidate.username || '').trim()
      || String(candidate.email || '').trim()
      || String(candidate.phone || '').trim()
      || '未设置';
  };

  const getUserEmailText = (userInfo) => {
    const candidate = userInfo && typeof userInfo === 'object' ? userInfo : {};
    return String(candidate.email || '').trim();
  };

  const getUserPhoneText = (userInfo) => {
    const candidate = userInfo && typeof userInfo === 'object' ? userInfo : {};
    return String(candidate.phone || '').trim();
  };

  const getUserAvatarText = (userInfo) => getUserDisplayName(userInfo).slice(0, 2).toUpperCase();

  const toggleAgeEditor = () => {
    ageEditorVisible.value = !ageEditorVisible.value;
  };

  const cancelAgeEditor = () => {
    ageEditorVisible.value = false;
    syncProfileForm(user.value || {});
  };

  const openProfilePrivacyPage = () => {
    ageEditorVisible.value = false;
    activeTab.value = 'profile_privacy';
  };

  const closeProfilePrivacyPage = () => {
    cancelAgeEditor();
    activeTab.value = 'profile';
  };

  const normalizeOptionValue = (value) => String(value ?? '').trim();
  const getSelectedOption = (options, value) => (options || []).find((option) => normalizeOptionValue(option.value) === normalizeOptionValue(value)) || null;
  const getSelectedOptionLabel = (options, value, placeholder = '请选择') => {
    const matched = getSelectedOption(options, value);
    return matched ? matched.label : placeholder;
  };
  const getSelectedOptionHint = (options, value, fallback = '') => {
    const matched = getSelectedOption(options, value);
    return matched ? String(matched.hint || '').trim() : fallback;
  };

  const toggleDropdown = (dropdownKey) => {
    openDropdownKey.value = openDropdownKey.value === dropdownKey ? '' : dropdownKey;
  };

  const closeDropdown = () => {
    openDropdownKey.value = '';
  };

  const selectDropdownValue = (dropdownKey, target, field, value) => {
    if (target && typeof target === 'object' && field) {
      target[field] = value;
    }
    openDropdownKey.value = '';
  };

  const handleDocumentClick = (event) => {
    const target = event && event.target;
    if (!(target instanceof Element)) {
      closeDropdown();
      return;
    }
    if (!target.closest('[data-custom-dropdown]')) {
      closeDropdown();
    }
  };

  const handleActiveTabChange = createMobileTabChangeHandler({
    syncRouteFromActiveTab,
    closeDropdown,
    scrollMainToTop,
    fetchFamilyOverview: () => familyModule.fetchFamilyOverview(),
    familyHasGroup: () => familyModule.familyHasGroup.value,
    connectFamilyNotificationWebSocket: () => familyModule.connectFamilyNotificationWebSocket(),
    fetchReceivedFamilyInvitations: () => familyModule.fetchReceivedFamilyInvitations(),
    chatHistoryLoaded: () => chatModule.chatHistoryLoaded.value,
    fetchChatHistory: () => chatModule.fetchChatHistory(),
    scrollToBottom: () => chatModule.scrollToBottom(),
    fetchHistory: () => tasksModule.fetchHistory(),
    fetchTasks: () => tasksModule.fetchTasks(),
	    fetchRiskTrend: () => chartsModule.fetchRiskTrend(),
    fetchCurrentRegionCaseStats: () => chartsModule.fetchCurrentRegionCaseStats(),
    resetSimulation,
    fetchSimulationPacks,
    fetchSimulationSessions,
    resumeOngoingSimulationSession
  });

  watch(activeTab, handleActiveTabChange);
  watch([authMode, loginMethod], () => {
    if (requiresGraphCaptcha.value) {
      fetchCaptcha();
    } else {
      form.captchaCode = '';
    }
  });

  let pollInterval = null;
  startPolling = () => {
    tasksModule.fetchTasks({ silent: true });
    tasksModule.fetchHistory({ silent: true });
    familyModule.fetchFamilyOverview({ silent: true });
    chartsModule.fetchRiskTrend();
    chartsModule.fetchCurrentRegionCaseStats();
    alertsModule.connectAlertWebSocket();
    familyModule.connectFamilyNotificationWebSocket();

    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(() => {
      if (isAuthenticated.value && activeTab.value === 'tasks') {
        tasksModule.fetchTasks({ silent: true });
        chartsModule.fetchRiskTrend();
        chartsModule.fetchCurrentRegionCaseStats();
      }

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

  const handleWindowResize = () => {
    chartsModule.resizeCharts();
  };

  const startBannerCarousel = () => {
    bannerCarouselTimer = setInterval(() => {
      currentBannerIndex.value = (currentBannerIndex.value + 1) % 2;
    }, 5000);
  };

  const stopBannerCarousel = () => {
    if (bannerCarouselTimer) {
      clearInterval(bannerCarouselTimer);
      bannerCarouselTimer = null;
    }
  };

  onMounted(async () => {
    stopTabRouter = tabRouter.mount({
      getContext: getRouteContext,
      onResolve: applyResolvedRoute
    });
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
    window.addEventListener('resize', handleWindowResize);
    document.addEventListener('click', handleDocumentClick);
    startBannerCarousel();
  });

  onUnmounted(() => {
    stopPolling();
    chartsModule.disposeCharts();
    clearSMSCodeCooldownTimer();
    if (stopTabRouter) stopTabRouter();
    window.removeEventListener('resize', handleWindowResize);
    document.removeEventListener('click', handleDocumentClick);
    stopBannerCarousel();
  });

  return {
    isAuthenticated,
    authReady,
    token,
    user,
    authMode,
    loginMethod,
    activeTab,
    loading,
    ageEditorVisible,
    captchaImage,
    toasts,
    openDropdownKey,
    currentBannerIndex,
    simulationGenerating,
    simulationSubmitting,
    simulationForm,
    simulationPack,
    simulationCurrentStep,
    simulationCurrentScore,
    simulationStatus,
    simulationViewMode,
    simulationAnswers,
    simulationResult,
    simulationPackList,
    simulationSessionList,
    form,
    ageForm,
    profileForm,
    occupationOptions,
    provinceOptions,
    cityOptions,
    districtOptions,
    profileSaving,
    locationResolving,
    demoSMSCode,
    requiresGraphCaptcha,
    shouldShowSMSCodeSection,
    authSubmitLabel,
    smsCodeButtonText,
    canSendSMSCode,
    handleAuth,
    fetchCaptcha,
    sendSMSCode,
    logout,
    updateAge,
    updateUserProfile,
    requestCurrentRegion,
    handleProvinceChange,
    handleCityChange,
    handleDistrictChange,
    selectProvinceValue,
    selectCityValue,
    selectDistrictValue,
    deleteAccount,
    getUserDisplayName,
    getUserEmailText,
    getUserPhoneText,
    getUserAvatarText,
    toggleAgeEditor,
    cancelAgeEditor,
    openProfilePrivacyPage,
    closeProfilePrivacyPage,
    getSelectedOptionLabel,
    getSelectedOptionHint,
    toggleDropdown,
    closeDropdown,
    generateSimulationPack,
    startSimulationSession,
    submitSimulationAnswer,
    resetSimulation,
    openSimulationExamView,
    closeSimulationExamView,
    fetchSimulationPacks,
    fetchSimulationSessions,
    resumeOngoingSimulationSession,
    deleteSimulationSession,
    selectDropdownValue,
    ...tasksModule,
    ...alertsModule,
    ...familyModule,
    ...chartsModule,
    ...chatModule
  };
}
