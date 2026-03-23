(function (global) {
    const createDesktopTabChangeHandler = (deps) => {
        return (newTab) => {
            deps.syncRouteFromActiveTab();

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
            if (newTab === 'risk_trend') deps.fetchRiskTrend();
            if (newTab === 'admin_stats') deps.fetchAdminStats();
        };
    };

    global.SentinelTabEffects = {
        createDesktopTabChangeHandler
    };
})(window);
