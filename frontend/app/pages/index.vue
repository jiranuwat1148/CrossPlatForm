<script setup lang="ts">
const { source, dest, sourceFiles, busy, copiedCount, reloadFiles, backup } = useFolders()
useHead({ title: 'ต้นทาง | Flash Drive Backup' })
</script>

<template>
  <FileList title="ไฟล์ต้นทาง" :path="source" :files="sourceFiles" :busy="busy" @reload="reloadFiles('source')" />
  <button class="button" :disabled="busy || !source || !dest" @click="backup">
    {{ busy ? 'กำลังทำงาน...' : 'เริ่มสำรองข้อมูล' }}
  </button>

  <p v-if="copiedCount !== null">
    สำรองเสร็จแล้ว คัดลอก {{ copiedCount }} ไฟล์
  </p>
</template>
