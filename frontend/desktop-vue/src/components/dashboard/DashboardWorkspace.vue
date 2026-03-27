<template>
  <div class="flex-1 flex flex-col h-full overflow-hidden bg-slate-100">
    <DashboardTopNav :app="app" />

    <main :class="['flex-1 custom-scrollbar', activeTab === 'admin_stats' || activeTab === 'family' || activeTab === 'profile' || activeTab === 'simulation_quiz' || isTaskCenterTab(activeTab) ? 'overflow-hidden' : 'overflow-y-auto']">
      <div :class="[
        activeTab === 'admin_stats' || activeTab === 'family' || activeTab === 'simulation_quiz' || isTaskCenterTab(activeTab)
          ? 'w-full p-4 min-w-0'
          : activeTab === 'profile'
            ? 'w-full max-w-[1600px] mx-auto p-4 min-w-0'
            : 'max-w-[1600px] mx-auto p-4',
        activeTab === 'admin_stats' || activeTab === 'family' || activeTab === 'profile' || activeTab === 'simulation_quiz' || isTaskCenterTab(activeTab) ? 'h-full' : 'min-h-full'
      ]">
        <DashboardCoreViews :app="app" />
        <DashboardAnalyticsAdminViews :app="app" />
        <DashboardAccountFamilyViews :app="app" />
        <DashboardReviewLibraryViews :app="app" />
      </div>
    </main>
  </div>
</template>

<script>
import DashboardTopNav from './DashboardTopNav.vue';
import DashboardCoreViews from './views/DashboardCoreViews.vue';
import DashboardAnalyticsAdminViews from './views/DashboardAnalyticsAdminViews.vue';
import DashboardAccountFamilyViews from './views/DashboardAccountFamilyViews.vue';
import DashboardReviewLibraryViews from './views/DashboardReviewLibraryViews.vue';

export default {
  name: 'DashboardWorkspace',
  components: {
    DashboardTopNav,
    DashboardCoreViews,
    DashboardAnalyticsAdminViews,
    DashboardAccountFamilyViews,
    DashboardReviewLibraryViews
  },
  props: {
    app: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    return props.app;
  }
};
</script>
