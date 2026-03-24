<template src="./DashboardAccountFamilyViews.template.html"></template>

<script>
export default {
  name: 'DashboardAccountFamilyViews',
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
