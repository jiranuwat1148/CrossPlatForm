<script setup lang="ts">
const { dest, destFiles, busy, reloadFiles } = useFolders()
useHead({ title: 'ปลายทาง | Flash Drive Backup' })
</script>

<template>
  <FileList title="ไฟล์ปลายทาง" :path="dest" :files="destFiles" :busy="busy"
    @reload="reloadFiles('dest')" />
</template>
