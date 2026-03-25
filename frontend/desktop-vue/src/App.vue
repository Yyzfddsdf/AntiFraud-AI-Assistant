<template>
  <div class="h-full flex flex-col transition-colors duration-300">
    <AppOverlays :app="app" />
    <DesktopAuthView v-if="authReady && !isAuthenticated" :app="app" />
    <GeoRiskMapPage v-else-if="activeTab === 'geo_risk_map_full'" :app="app" />
    <DashboardWorkspace v-else :app="app" />
  </div>
</template>

<script>
import { useDesktopApp } from './app/useDesktopApp';
import AppOverlays from './components/shell/AppOverlays.vue';
import DesktopAuthView from './components/auth/DesktopAuthView.vue';
import GeoRiskMapPage from './components/geo/GeoRiskMapPage.vue';
import DashboardWorkspace from './components/dashboard/DashboardWorkspace.vue';

export default {
  name: 'DesktopApp',
  components: {
    AppOverlays,
    DesktopAuthView,
    GeoRiskMapPage,
    DashboardWorkspace
  },
  setup() {
    const app = useDesktopApp();
    return {
      app,
      ...app
    };
  }
};
</script>
