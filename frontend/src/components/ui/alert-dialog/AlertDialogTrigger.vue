<script setup>
import { AlertDialogTrigger } from 'reka-ui'
const props = defineProps({ asChild: { type: Boolean, default: false } })
</script>
<template>
  <AlertDialogTrigger :as-child="props.asChild"><slot /></AlertDialogTrigger>
</template>
