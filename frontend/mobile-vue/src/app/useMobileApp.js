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
    recentTagsText: ''
  });
  const occupationOptions = ref([]);
  const profileSaving = ref(false);
  const demoSMSCode = '000000';
  const smsCodeCooldown = ref(0);
  const smsCodeSending = ref(false);
  let smsCodeCooldownTimer = null;
  let bannerCarouselTimer = null;

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

  const fetchOccupationOptions = async () => {
    const res = await request('/user/profile/options/occupations', 'GET', null, { silent: true });
    if (res && Array.isArray(res.occupations)) {
      occupationOptions.value = res.occupations.filter((item) => typeof item === 'string' && item.trim());
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
      occupation: profileForm.occupation
    });
    profileSaving.value = false;
    if (res && res.user) {
      user.value = res.user;
      syncProfileForm(res.user);
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
    fetchFamilyOverview: () => familyModule.fetchFamilyOverview(),
    familyHasGroup: () => familyModule.familyHasGroup.value,
    connectFamilyNotificationWebSocket: () => familyModule.connectFamilyNotificationWebSocket(),
    fetchReceivedFamilyInvitations: () => familyModule.fetchReceivedFamilyInvitations(),
    chatHistoryLoaded: () => chatModule.chatHistoryLoaded.value,
    fetchChatHistory: () => chatModule.fetchChatHistory(),
    scrollToBottom: () => chatModule.scrollToBottom(),
    fetchHistory: () => tasksModule.fetchHistory(),
    fetchTasks: () => tasksModule.fetchTasks(),
    fetchRiskTrend: () => chartsModule.fetchRiskTrend()
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
    alertsModule.connectAlertWebSocket();
    familyModule.connectFamilyNotificationWebSocket();

    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(() => {
      if (isAuthenticated.value && activeTab.value === 'tasks') {
        tasksModule.fetchTasks({ silent: true });
        chartsModule.fetchRiskTrend();
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
    fetchCaptcha();
    if (token.value) {
      await getUserInfo();
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
    form,
    ageForm,
    profileForm,
    occupationOptions,
    profileSaving,
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
    selectDropdownValue,
    ...tasksModule,
    ...alertsModule,
    ...familyModule,
    ...chartsModule,
    ...chatModule
  };
}
