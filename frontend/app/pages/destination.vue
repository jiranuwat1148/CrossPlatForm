<script setup lang="ts">
const { dest, destFiles, selectedDestFiles, busy, setSelectedFiles, deleteSelected } = useFolders()
useHead({ title: 'ปลายทาง | Flash Drive Backup' })

async function confirmDeleteSelected() {
  if (!selectedDestFiles.value.length) return

  const confirmed = window.confirm(
    `ยืนยันลบไฟล์ที่เลือก ${selectedDestFiles.value.length} ไฟล์จากปลายทางหรือไม่?`,
  )

  if (confirmed) await deleteSelected()
}
</script>

<template>
  <FileList
    title="ไฟล์ปลายทาง"
    :path="dest"
    :files="destFiles"
    :selected="selectedDestFiles"
    :busy="busy"
    @selection-change="setSelectedFiles('dest', $event)"
  />

  <button
    class="button"
    :disabled="busy || !dest || !selectedDestFiles.length"
    @click="confirmDeleteSelected"
  >
    ลบไฟล์ที่เลือก ({{ selectedDestFiles.length }})
  </button>
</template>
