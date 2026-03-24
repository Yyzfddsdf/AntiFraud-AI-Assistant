<template>
  <div class="flex-1 flex flex-col md:flex-row h-full overflow-hidden bg-slate-50 relative">
    <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
      <div class="absolute inset-0 tech-bg opacity-40"></div>
      <div class="absolute -left-20 top-0 w-[500px] h-[500px] bg-brand-400/20 rounded-full mix-blend-multiply filter blur-3xl opacity-50"></div>
      <div class="absolute left-60 top-40 w-[400px] h-[400px] bg-purple-400/20 rounded-full mix-blend-multiply filter blur-3xl opacity-50"></div>
      <div class="absolute bottom-0 left-0 w-[600px] h-[400px] bg-indigo-300/20 rounded-full mix-blend-multiply filter blur-3xl opacity-40"></div>
    </div>

    <DashboardSidebar :app="app" />

    <main class="flex-1 overflow-y-auto relative tech-bg">
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
