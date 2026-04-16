export const createDesktopTabChangeHandler = (deps) => {
  return (newTab) => {
    deps.syncRouteFromActiveTab();

    if (newTab === 'chat') deps.fetchChatHistory();
    if (newTab === 'case_review') {
      deps.fetchPendingReviews();
    }
    if (newTab === 'case_library') {
      deps.fetchCaseLibrary();
      deps.fetchCaseOptionLists();
    }
    if (newTab === 'family') {
      deps.fetchFamilyOverview().then(() => {
        if (deps.familyHasGroup()) {
          deps.connectFamilyNotificationWebSocket();
        } else {
          deps.fetchReceivedFamilyInvitations();
        }
      });
    }
    if (newTab === 'users') deps.fetchUsers();
    if (newTab === 'history') deps.fetchHistory();
    if (newTab === 'tasks') deps.fetchTasks();
    if (newTab === 'risk_trend') {
      deps.fetchRiskTrend();
      deps.fetchCurrentRegionCaseStats();
    }
    if (newTab === 'admin_stats') deps.fetchAdminStats();
    if (newTab === 'geo_risk_map') deps.fetchGeoRiskMap();
    if (newTab === 'geo_risk_map_full') deps.fetchGeoRiskMap();
    if (newTab === 'simulation_quiz') {
      deps.resetSimulation();
      deps.fetchSimulationPacks();
      deps.fetchSimulationSessions();
      deps.resumeOngoingSimulationSession();
    }
  };
};
