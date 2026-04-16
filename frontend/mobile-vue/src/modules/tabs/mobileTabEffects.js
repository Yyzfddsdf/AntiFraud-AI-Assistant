export function createMobileTabChangeHandler(deps) {
  return (newTab) => {
    deps.syncRouteFromActiveTab();
    deps.closeDropdown();
    deps.scrollMainToTop?.();

    if (newTab === 'family') {
      deps.fetchFamilyOverview().then(() => {
        if (deps.familyHasGroup()) {
          deps.connectFamilyNotificationWebSocket();
        } else {
          deps.fetchReceivedFamilyInvitations();
        }
      });
    }

    if (newTab === 'chat') {
      if (!deps.chatHistoryLoaded()) {
        deps.fetchChatHistory();
      } else {
        deps.scrollToBottom();
      }
    }

    if (newTab === 'history') {
      deps.fetchHistory();
    }

    if (newTab === 'tasks') {
      deps.fetchTasks();
      deps.fetchRiskTrend();
      deps.fetchCurrentRegionCaseStats();
    }

    if (newTab === 'risk_trend') {
      deps.fetchRiskTrend();
      deps.fetchCurrentRegionCaseStats();
    }

    if (newTab === 'simulation_quiz') {
      deps.resetSimulation();
      deps.fetchSimulationPacks();
      deps.fetchSimulationSessions();
      deps.resumeOngoingSimulationSession();
    }
  };
}
