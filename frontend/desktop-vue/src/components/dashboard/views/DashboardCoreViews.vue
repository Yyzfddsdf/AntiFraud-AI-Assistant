<template src="./DashboardCoreViews.template.html"></template>

<script>
export default {
  name: 'DashboardCoreViews',
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
