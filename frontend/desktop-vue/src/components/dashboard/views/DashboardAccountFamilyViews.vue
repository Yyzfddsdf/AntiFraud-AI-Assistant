<template src="./DashboardAccountFamilyViews.template.html"></template>

<script>
import CustomSelect from '../../common/CustomSelect.vue';
import DashboardSectionShell from '../DashboardSectionShell.vue';

export default {
  name: 'DashboardAccountFamilyViews',
  components: {
    CustomSelect,
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
