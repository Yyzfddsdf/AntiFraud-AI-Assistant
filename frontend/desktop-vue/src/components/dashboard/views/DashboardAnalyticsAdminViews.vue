<template src="../../../../DashboardAnalyticsAdminViews.template.html"></template>

<script>
export default {
  name: 'DashboardAnalyticsAdminViews',
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
