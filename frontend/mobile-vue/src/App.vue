<template src="./App.template.html"></template>

<script>
import { useMobileApp } from './app/useMobileApp';

export default {
  name: 'MobileApp',
  setup() {
    return useMobileApp();
  }
};
</script>
