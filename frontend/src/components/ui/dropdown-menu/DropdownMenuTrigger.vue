<script setup>
import { DropdownMenuTrigger } from 'reka-ui'
const props = defineProps({ asChild: { type: Boolean, default: false } })
</script>
<template>
  <DropdownMenuTrigger :as-child="props.asChild"><slot /></DropdownMenuTrigger>
</template>
