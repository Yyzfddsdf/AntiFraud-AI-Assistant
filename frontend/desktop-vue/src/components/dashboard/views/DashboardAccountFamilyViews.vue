<template src="./DashboardAccountFamilyViews.template.html"></template>

<script>
import CustomSelect from '../../common/CustomSelect.vue';

export default {
  name: 'DashboardAccountFamilyViews',
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
