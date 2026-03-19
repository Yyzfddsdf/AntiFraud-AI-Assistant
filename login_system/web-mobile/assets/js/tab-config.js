(function (global) {
    const createMobileTabRouter = () => {
        if (!global.SentinelTabRouter?.createTabRouter) {
            return null;
        }

        return global.SentinelTabRouter.createTabRouter({
            appTabs: ['tasks', 'history', 'risk_trend', 'chat', 'alerts', 'family', 'family_invite', 'profile', 'profile_privacy', 'submit'],
            defaultAppTab: 'tasks'
        });
    };

    global.SentinelTabConfig = {
        createMobileTabRouter
    };
})(window);
