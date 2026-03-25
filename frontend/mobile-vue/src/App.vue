<template src="./App.template.html"></template>

<script>
import { computed, reactive } from 'vue';
import { useMobileApp } from './app/useMobileApp';
import MobileAuthScreen from './components/auth/MobileAuthScreen.vue';
import MobileLoadingOverlay from './components/common/MobileLoadingOverlay.vue';
import MobileToastStack from './components/common/MobileToastStack.vue';
import MobileAppHeader from './components/layout/MobileAppHeader.vue';
import MobileBottomNav from './components/layout/MobileBottomNav.vue';
import MobileAlertDetailModal from './components/modals/MobileAlertDetailModal.vue';
import MobileFamilyAlertModal from './components/modals/MobileFamilyAlertModal.vue';
import MobileFamilyManageModal from './components/modals/MobileFamilyManageModal.vue';
import MobileTaskDetailModal from './components/modals/MobileTaskDetailModal.vue';
import MobileDashboardScreens from './components/screens/MobileDashboardScreens.vue';
import MobileProfileScreens from './components/screens/MobileProfileScreens.vue';
import MobileSocialScreens from './components/screens/MobileSocialScreens.vue';
import MobileWorkflowScreens from './components/screens/MobileWorkflowScreens.vue';

export default {
  name: 'MobileApp',
  components: {
    MobileAuthScreen,
    MobileLoadingOverlay,
    MobileToastStack,
    MobileAppHeader,
    MobileBottomNav,
    MobileAlertDetailModal,
    MobileFamilyAlertModal,
    MobileFamilyManageModal,
    MobileTaskDetailModal,
    MobileDashboardScreens,
    MobileProfileScreens,
    MobileSocialScreens,
    MobileWorkflowScreens
  },
  setup() {
    const app = useMobileApp();

    const loadingOverlayState = reactive({
      authReady: app.authReady
    });

    const toastState = reactive({
      toasts: app.toasts
    });

    const authState = reactive({
      authMode: app.authMode,
      loginMethod: app.loginMethod,
      form: app.form,
      loading: app.loading,
      captchaImage: app.captchaImage,
      shouldShowSMSCodeSection: app.shouldShowSMSCodeSection,
      requiresGraphCaptcha: app.requiresGraphCaptcha,
      authSubmitLabel: app.authSubmitLabel,
      smsCodeButtonText: app.smsCodeButtonText,
      canSendSMSCode: app.canSendSMSCode,
      handleAuth: app.handleAuth,
      fetchCaptcha: app.fetchCaptcha,
      sendSMSCode: app.sendSMSCode
    });

    const headerState = reactive({
      activeTab: app.activeTab,
      alertUnreadCount: app.alertUnreadCount
    });

    const bottomNavState = reactive({
      activeTab: app.activeTab,
      alertUnreadCount: app.alertUnreadCount
    });

    const dashboardScreensState = reactive({
      activeTab: app.activeTab,
      currentBannerIndex: app.currentBannerIndex,
      todayRiskStatsSummary: app.todayRiskStatsSummary,
      riskData: app.riskData,
      getRiskTrendHeadline: app.getRiskTrendHeadline,
      formatRiskTrendDescriptor: app.formatRiskTrendDescriptor,
      tasks: app.tasks,
      getStatusClass: app.getStatusClass,
      getStatusLabel: app.getStatusLabel,
      formatTime: app.formatTime,
      viewTaskDetail: app.viewTaskDetail,
      history: app.history,
      getRiskClass: app.getRiskClass,
      normalizeRiskLevelText: app.normalizeRiskLevelText,
      viewHistoryDetail: app.viewHistoryDetail,
      deleteHistoryCase: app.deleteHistoryCase,
      riskStatsSummary: app.riskStatsSummary,
      regionCaseStats: app.regionCaseStats,
      getRegionSignalClass: app.getRegionSignalClass,
      getRegionStatsSignal: app.getRegionStatsSignal,
      getRegionTopScamHint: app.getRegionTopScamHint,
      getRecentRiskTrendRows: app.getRecentRiskTrendRows,
      formatChartLabel: app.formatChartLabel
    });

    const socialScreensState = reactive({
      activeTab: app.activeTab,
      chatMessages: app.chatMessages,
      isChatting: app.isChatting,
      clearChatHistory: app.clearChatHistory,
      openImage: app.openImage,
      chatImages: app.chatImages,
      removeChatImage: app.removeChatImage,
      sendChatMessage: app.sendChatMessage,
      triggerChatImagePicker: app.triggerChatImagePicker,
      handleChatImageSelect: app.handleChatImageSelect,
      chatInput: app.chatInput,
      recentRiskAlerts: app.recentRiskAlerts,
      alertUnreadCount: app.alertUnreadCount,
      openAlertCaseDetail: app.openAlertCaseDetail,
      formatTime: app.formatTime,
      getAlertSeverityTheme: app.getAlertSeverityTheme,
      familyLoading: app.familyLoading,
      familyHasGroup: app.familyHasGroup,
      familyReceivedInvitations: app.familyReceivedInvitations,
      familyReceivedLoading: app.familyReceivedLoading,
      familyAcceptingInvitations: app.familyAcceptingInvitations,
      acceptFamilyInvitation: app.acceptFamilyInvitation,
      familyCreateForm: app.familyCreateForm,
      createFamily: app.createFamily,
      familyAcceptForm: app.familyAcceptForm,
      familyOverview: app.familyOverview,
      familyMembers: app.familyMembers,
      familyUnreadCount: app.familyUnreadCount,
      familyGuardianLinks: app.familyGuardianLinks,
      familyNotifications: app.familyNotifications
    });

    const profileScreensState = reactive({
      activeTab: app.activeTab,
      user: app.user,
      logout: app.logout,
      openProfilePrivacyPage: app.openProfilePrivacyPage,
      closeProfilePrivacyPage: app.closeProfilePrivacyPage,
      getUserPhoneText: app.getUserPhoneText,
      getUserEmailText: app.getUserEmailText,
      ageEditorVisible: app.ageEditorVisible,
      ageForm: app.ageForm,
      profileForm: app.profileForm,
      occupationOptions: app.occupationOptions,
      provinceOptions: app.provinceOptions,
      cityOptions: app.cityOptions,
      districtOptions: app.districtOptions,
      profileSaving: app.profileSaving,
      locationResolving: app.locationResolving,
      openDropdownKey: app.openDropdownKey,
      toggleAgeEditor: app.toggleAgeEditor,
      cancelAgeEditor: app.cancelAgeEditor,
      toggleDropdown: app.toggleDropdown,
      selectDropdownValue: app.selectDropdownValue,
      selectProvinceValue: app.selectProvinceValue,
      selectCityValue: app.selectCityValue,
      selectDistrictValue: app.selectDistrictValue,
      requestCurrentRegion: app.requestCurrentRegion,
      updateUserProfile: app.updateUserProfile,
      deleteAccount: app.deleteAccount
    });

    const workflowScreensState = reactive({
      activeTab: app.activeTab,
      analyzeForm: app.analyzeForm,
      handleFileSelect: app.handleFileSelect,
      submitAnalysis: app.submitAnalysis,
      analyzing: app.analyzing,
      simulationViewMode: app.simulationViewMode,
      simulationForm: app.simulationForm,
      simulationGenerating: app.simulationGenerating,
      simulationSubmitting: app.simulationSubmitting,
      simulationPackList: app.simulationPackList,
      simulationSessionList: app.simulationSessionList,
      generateSimulationPack: app.generateSimulationPack,
      fetchSimulationPacks: app.fetchSimulationPacks,
      fetchSimulationSessions: app.fetchSimulationSessions,
      startSimulationSession: app.startSimulationSession,
      deleteSimulationSession: app.deleteSimulationSession,
      simulationPack: app.simulationPack,
      simulationCurrentStep: app.simulationCurrentStep,
      simulationCurrentScore: app.simulationCurrentScore,
      simulationStatus: app.simulationStatus,
      simulationAnswers: app.simulationAnswers,
      simulationResult: app.simulationResult,
      closeSimulationExamView: app.closeSimulationExamView,
      submitSimulationAnswer: app.submitSimulationAnswer
    });

    const familyManageModalState = reactive({
      visible: computed(() => app.activeTab.value === 'family_invite'),
      close: () => {
        app.activeTab.value = 'family';
      },
      familyInviteForm: app.familyInviteForm,
      familyRoleSelectOptions: app.familyRoleSelectOptions,
      familyGuardianForm: app.familyGuardianForm,
      familyGuardianSelectOptions: app.familyGuardianSelectOptions,
      familyProtectedSelectOptions: app.familyProtectedSelectOptions,
      openDropdownKey: app.openDropdownKey,
      toggleDropdown: app.toggleDropdown,
      getSelectedOptionLabel: app.getSelectedOptionLabel,
      getSelectedOptionHint: app.getSelectedOptionHint,
      selectDropdownValue: app.selectDropdownValue,
      createFamilyInvitation: app.createFamilyInvitation,
      createGuardianLink: app.createGuardianLink,
      familyMembers: app.familyMembers,
      familyOverview: app.familyOverview,
      familyDeletingMembers: app.familyDeletingMembers,
      deleteFamilyMember: app.deleteFamilyMember,
      familyInvitations: app.familyInvitations,
      formatTime: app.formatTime
    });

    const taskDetailModalState = reactive({
      selectedTask: app.selectedTask,
      close: () => {
        app.selectedTask.value = null;
      },
      getStatusClass: app.getStatusClass,
      getStatusLabel: app.getStatusLabel,
      formatTime: app.formatTime,
      openDropdownKey: app.openDropdownKey,
      toggleDropdown: app.toggleDropdown,
      closeDropdown: app.closeDropdown,
      exportData: app.exportData,
      parseRiskSummary: app.parseRiskSummary,
      parseReport: app.parseReport,
      extractAttackSteps: app.extractAttackSteps,
      extractScamKeywordSentences: app.extractScamKeywordSentences,
      getRiskClass: app.getRiskClass,
      normalizeRiskLevelText: app.normalizeRiskLevelText
    });

    const alertDetailModalState = reactive({
      activeAlertEvent: app.activeAlertEvent,
      acknowledgeActiveAlert: app.acknowledgeActiveAlert,
      getAlertSeverityTheme: app.getAlertSeverityTheme,
      formatTime: app.formatTime,
      openAlertHistory: app.openAlertHistory
    });

    const familyAlertModalState = reactive({
      familyAlertModalVisible: app.familyAlertModalVisible,
      activeFamilyNotification: app.activeFamilyNotification,
      acknowledgeFamilyAlert: app.acknowledgeFamilyAlert,
      openFamilyNotificationCenter: app.openFamilyNotificationCenter,
      formatTime: app.formatTime
    });

    return {
      ...app,
      loadingOverlayState,
      toastState,
      authState,
      headerState,
      bottomNavState,
      dashboardScreensState,
      socialScreensState,
      profileScreensState,
      workflowScreensState,
      familyManageModalState,
      taskDetailModalState,
      alertDetailModalState,
      familyAlertModalState
    };
  }
};
</script>
