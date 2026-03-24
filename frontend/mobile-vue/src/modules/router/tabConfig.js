import { createTabRouter } from './tabRouter';

export function createMobileTabRouter() {
  return createTabRouter({
    appTabs: ['tasks', 'history', 'risk_trend', 'chat', 'alerts', 'family', 'family_invite', 'profile', 'profile_privacy', 'submit'],
    defaultAppTab: 'tasks'
  });
}
