<template src="./DashboardCoreViews.template.html"></template>

<script>
import DesktopSimulationQuizView from './DesktopSimulationQuizView.vue';

export default {
  name: 'DashboardCoreViews',
  components: {
    DesktopSimulationQuizView
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
