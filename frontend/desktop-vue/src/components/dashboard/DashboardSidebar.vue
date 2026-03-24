<template src="./DashboardSidebar.template.html"></template>

<script>
export default {
  name: 'DashboardSidebar',
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
