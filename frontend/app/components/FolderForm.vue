<script setup lang="ts">
const props = defineProps<{
  label: string
  path: string
  placeholder: string
  disabled: boolean
}>()
defineEmits<{ submit: [path: string] }>()
const draft = ref(props.path)
watch(() => props.path, (value) => { draft.value = value })
</script>

<template>
  <form class="folder-form" @submit.prevent="$emit('submit', draft)">
    <label>
      <span>{{ label }}</span>
      <input v-model="draft" :placeholder="placeholder" :disabled="disabled"
        spellcheck="false" autocomplete="off" />
    </label>
    <button class="button" :aria-label="`ตั้ง${label}`"
      :disabled="disabled || !draft.trim()">ตั้งค่า</button>
  </form>
</template>
