(function (global) {
    const adminTabs = new Set(['admin_stats', 'users', 'case_review', 'case_library']);

    const createDesktopTabRouter = () => {
        if (!global.SentinelTabRouter?.createTabRouter) {
            return null;
        }

        return global.SentinelTabRouter.createTabRouter({
            appTabs: ['submit', 'tasks', 'risk_trend', 'history', 'family', 'profile', 'admin_stats', 'users', 'case_review', 'case_library'],
            defaultAppTab: 'tasks',
            isTabAllowed: (tab, context) => {
                if (!adminTabs.has(tab)) return true;
                return String(context?.userRole || '').trim() === 'admin';
            }
        });
    };

    global.SentinelTabConfig = {
        createDesktopTabRouter
    };
})(window);
