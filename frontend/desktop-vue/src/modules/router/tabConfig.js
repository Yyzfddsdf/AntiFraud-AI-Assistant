import { createTabRouter } from './tabRouter';

const adminTabs = new Set(['chat', 'admin_stats', 'geo_risk_map_full', 'users', 'case_review', 'case_library', 'profile']);

export const createDesktopTabRouter = () => createTabRouter({
  appTabs: ['chat', 'admin_stats', 'geo_risk_map_full', 'users', 'case_review', 'case_library', 'profile'],
  defaultAppTab: 'admin_stats',
  isTabAllowed: (tab, context) => {
    return adminTabs.has(tab) && String(context?.userRole || '').trim() === 'admin';
  }
});
