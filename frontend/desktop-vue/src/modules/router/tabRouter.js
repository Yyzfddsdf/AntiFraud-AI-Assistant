const AUTH_ROUTE = 'auth';

const normalizeHashTab = (hash) => {
  const normalizedHash = String(hash || '').replace(/^#\/?/, '').trim();
  if (!normalizedHash) return '';
  return normalizedHash.split(/[?#]/)[0].trim().toLowerCase();
};

const buildHash = (routeTab) => (routeTab === AUTH_ROUTE ? '#/auth' : `#/${routeTab}`);

export const createTabRouter = ({ appTabs = [], defaultAppTab = 'tasks', isTabAllowed = () => true } = {}) => {
  const appTabSet = new Set(appTabs);
  const fallbackTab = appTabSet.has(defaultAppTab) ? defaultAppTab : (appTabs[0] || 'tasks');
  let pendingProtectedTab = fallbackTab;

  const normalizeAuthorizedTab = (candidate, context) => {
    if (!appTabSet.has(candidate)) return fallbackTab;
    try {
      return isTabAllowed(candidate, context) ? candidate : fallbackTab;
    } catch (error) {
      console.warn('tab route authorization failed', error);
      return fallbackTab;
    }
  };

  const writeHash = (routeTab, { replace = false } = {}) => {
    const nextHash = buildHash(routeTab);
    if (window.location.hash === nextHash) return;

    if (replace) {
      const nextURL = `${window.location.pathname}${window.location.search}${nextHash}`;
      window.history.replaceState(null, '', nextURL);
      return;
    }

    window.location.hash = nextHash;
  };

  const resolve = (context = {}) => {
    const requestedTab = normalizeHashTab(window.location.hash);
    const isAuthenticated = Boolean(context.isAuthenticated);

    if (!isAuthenticated) {
      if (appTabSet.has(requestedTab)) {
        pendingProtectedTab = requestedTab;
      }

      return {
        isAuthenticated: false,
        requestedTab,
        routeTab: AUTH_ROUTE,
        activeTab: fallbackTab,
        shouldReplace: requestedTab !== AUTH_ROUTE
      };
    }

    let resolvedTab = requestedTab;
    if (!resolvedTab || resolvedTab === AUTH_ROUTE) {
      resolvedTab = pendingProtectedTab || fallbackTab;
    }

    resolvedTab = normalizeAuthorizedTab(resolvedTab, context);
    pendingProtectedTab = resolvedTab;

    return {
      isAuthenticated: true,
      requestedTab,
      routeTab: resolvedTab,
      activeTab: resolvedTab,
      shouldReplace: requestedTab !== resolvedTab
    };
  };

  const reconcile = (context = {}, { replace = false } = {}) => {
    const resolved = resolve(context);
    if (replace || resolved.shouldReplace) {
      writeHash(resolved.routeTab, { replace: true });
    }
    return resolved;
  };

  const sync = (context = {}, activeTab, { replace = false } = {}) => {
    const nextRouteTab = context.isAuthenticated
      ? normalizeAuthorizedTab(String(activeTab || '').trim().toLowerCase(), context)
      : AUTH_ROUTE;

    if (nextRouteTab !== AUTH_ROUTE) {
      pendingProtectedTab = nextRouteTab;
    }

    writeHash(nextRouteTab, { replace });
    return nextRouteTab;
  };

  const mount = ({ getContext, onResolve }) => {
    if (typeof getContext !== 'function' || typeof onResolve !== 'function') {
      throw new Error('tab router mount requires getContext and onResolve callbacks');
    }

    const handleHashChange = () => {
      const resolved = reconcile(getContext());
      onResolve(resolved);
    };

    window.addEventListener('hashchange', handleHashChange);
    handleHashChange();

    return () => window.removeEventListener('hashchange', handleHashChange);
  };

  return {
    reconcile,
    sync,
    mount
  };
};
