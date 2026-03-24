<template src="./AppOverlays.template.html"></template>

<script>
export default {
  name: 'AppOverlays',
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
