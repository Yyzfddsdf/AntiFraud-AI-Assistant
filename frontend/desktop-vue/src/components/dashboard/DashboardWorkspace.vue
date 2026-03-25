<template>
  <div class="flex-1 flex flex-col md:flex-row h-full overflow-hidden bg-slate-50 relative">
    <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
      <div class="absolute inset-0 tech-bg"></div>
      <!-- High-end, concise geometric overlay instead of messy colorful blurs -->
      <div class="absolute top-0 inset-x-0 h-[500px] bg-gradient-to-b from-brand-50/80 to-transparent"></div>
    </div>

    <DashboardSidebar :app="app" />

    <main class="flex-1 overflow-y-auto relative z-30 custom-scrollbar">
      <div class="max-w-7xl mx-auto p-6 md:p-10 min-h-full">
        <DashboardCoreViews :app="app" />
        <DashboardAnalyticsAdminViews :app="app" />
        <DashboardAccountFamilyViews :app="app" />
        <DashboardReviewLibraryViews :app="app" />
      </div>
    </main>
  </div>
</template>

<script>
import DashboardSidebar from './DashboardSidebar.vue';
import DashboardCoreViews from './views/DashboardCoreViews.vue';
import DashboardAnalyticsAdminViews from './views/DashboardAnalyticsAdminViews.vue';
import DashboardAccountFamilyViews from './views/DashboardAccountFamilyViews.vue';
import DashboardReviewLibraryViews from './views/DashboardReviewLibraryViews.vue';

export default {
  name: 'DashboardWorkspace',
  components: {
    DashboardSidebar,
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
