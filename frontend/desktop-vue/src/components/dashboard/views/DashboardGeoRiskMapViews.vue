<template src="./DashboardGeoRiskMapViews.template.html"></template>

<script>
export default {
  name: 'DashboardGeoRiskMapViews',
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
