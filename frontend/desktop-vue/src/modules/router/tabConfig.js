import { createTabRouter } from './tabRouter';

const adminTabs = new Set(['admin_stats', 'geo_risk_map', 'geo_risk_map_full', 'users', 'case_review', 'case_library']);

export const createDesktopTabRouter = () => createTabRouter({
  appTabs: ['submit', 'tasks', 'risk_trend', 'history', 'family', 'profile', 'admin_stats', 'geo_risk_map', 'geo_risk_map_full', 'users', 'case_review', 'case_library'],
  defaultAppTab: 'tasks',
  isTabAllowed: (tab, context) => {
    if (!adminTabs.has(tab)) return true;
    return String(context?.userRole || '').trim() === 'admin';
  }
});
