<template src="./DashboardReviewLibraryViews.template.html"></template>

<script>
export default {
  name: 'DashboardReviewLibraryViews',
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
