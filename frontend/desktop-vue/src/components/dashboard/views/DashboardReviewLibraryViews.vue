<template src="./DashboardReviewLibraryViews.template.html"></template>

<script>
import CustomSelect from '../../common/CustomSelect.vue';

export default {
  name: 'DashboardReviewLibraryViews',
  components: {
    CustomSelect
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
