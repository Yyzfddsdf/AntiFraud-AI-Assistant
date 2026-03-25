<template>
  <div class="flex-1 flex flex-col h-full overflow-hidden bg-slate-100">
    <DashboardTopNav :app="app" />

    <main class="flex-1 overflow-y-auto custom-scrollbar">
      <div class="max-w-[1600px] mx-auto p-4 min-h-full">
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
