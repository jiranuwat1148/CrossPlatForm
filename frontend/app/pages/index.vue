<script setup lang="ts">
const { source, sourceFiles, busy, reloadFiles } = useFolders()
useHead({ title: 'ต้นทาง | Flash Drive Backup' })
</script>

<template>
  <FileList title="ไฟล์ต้นทาง" :path="source" :files="sourceFiles" :busy="busy"
    @reload="reloadFiles('source')" />
</template>
