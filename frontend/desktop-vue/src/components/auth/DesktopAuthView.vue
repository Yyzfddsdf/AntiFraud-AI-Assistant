<template src="./DesktopAuthView.template.html"></template>

<script>
export default {
  name: 'DesktopAuthView',
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
