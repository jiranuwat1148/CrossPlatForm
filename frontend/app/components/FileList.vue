<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  title: string
  path: string
  files: string[]
  busy: boolean
  selected: string[]
}>()

const emit = defineEmits<{
  selectionChange: [files: string[]]
}>()

const searchQuery = ref('')

const filteredFiles = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase()
  if (!query) return props.files

  return props.files.filter((name) => name.toLocaleLowerCase().includes(query))
})

const allFilteredSelected = computed(() =>
  filteredFiles.value.length > 0 &&
  filteredFiles.value.every((name) => props.selected.includes(name)),
)

function changeAllSelection(event: Event) {
  const checkbox = event.target as HTMLInputElement
  const selected = new Set(props.selected)

  for (const name of filteredFiles.value) {
    if (checkbox.checked) selected.add(name)
    else selected.delete(name)
  }

  emit('selectionChange', Array.from(selected))
}

function changeSelection(name: string, event: Event) {
  const checkbox = event.target as HTMLInputElement
  const selected = new Set(props.selected)

  if (checkbox.checked) selected.add(name)
  else selected.delete(name)

  emit('selectionChange', Array.from(selected))
}

watch(() => props.files, (files) => {
  const selected = props.selected.filter((name) => files.includes(name))
  if (selected.length !== props.selected.length) emit('selectionChange', selected)
})
</script>

<template>
  <section class="file-panel" :aria-label="title" :aria-busy="busy">
    <div v-if="path" class="file-search">
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="11" cy="11" r="7" />
        <path d="m16 16 4 4" />
      </svg>
      <input
        v-model="searchQuery"
        type="search"
        placeholder="ค้นหาไฟล์..."
        aria-label="ค้นหาไฟล์"
      >
    </div>

    <header class="panel-header">
      <div>
        <h2>{{ title }}</h2>
        <p v-if="path" class="path">{{ path }}</p>
      </div>
    </header>

    <p v-if="!path" class="empty">กรุณาตั้งค่าโฟลเดอร์ก่อน</p>
    <p v-else-if="!files.length" class="empty">ไม่มีไฟล์ในโฟลเดอร์นี้</p>
    <p v-else-if="!filteredFiles.length" class="empty">ไม่พบไฟล์ที่ค้นหา</p>
    <table v-else>
      <thead>
        <tr>
          <th scope="col">
            <input
              type="checkbox"
              aria-label="เลือกทั้งหมด"
              :checked="allFilteredSelected"
              :disabled="busy || !filteredFiles.length"
              @change="changeAllSelection"
            >
            เลือกทั้งหมด
          </th>
          <th scope="col">ชื่อไฟล์</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="name in filteredFiles" :key="name">
          <td>
            <input
              type="checkbox"
              :checked="selected.includes(name)"
              :disabled="busy"
              :aria-label="`เลือก ${name}`"
              @change="changeSelection(name, $event)"
            >
          </td>
          <td>{{ name }}</td>
        </tr>
      </tbody>
    </table>

    <footer v-if="path" class="panel-footer">{{ files.length }} ไฟล์</footer>
  </section>
</template>
