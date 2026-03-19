(function (global) {
    const createMobileTabChangeHandler = (deps) => {
        return (newTab) => {
            deps.syncRouteFromActiveTab();
            deps.closeDropdown();

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
            if (newTab === 'history') deps.fetchHistory();
            if (newTab === 'tasks') {
                deps.fetchTasks();
                deps.fetchRiskTrend();
            }
            if (newTab === 'risk_trend') deps.fetchRiskTrend();
        };
    };

    global.SentinelTabEffects = {
        createMobileTabChangeHandler
    };
})(window);
