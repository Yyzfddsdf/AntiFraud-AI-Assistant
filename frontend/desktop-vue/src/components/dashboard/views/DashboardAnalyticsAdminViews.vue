<template src="../../../../DashboardAnalyticsAdminViews.template.html"></template>

<script>
import DashboardSectionShell from '../DashboardSectionShell.vue';

export default {
  name: 'DashboardAnalyticsAdminViews',
  components: {
    DashboardSectionShell
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
