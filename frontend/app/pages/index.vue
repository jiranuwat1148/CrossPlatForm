<script setup lang="ts">
const {
  source,
  dest,
  sourceFiles,
  selectedSourceFiles,
  busy,
  copiedCount,
  movedCount,
  tagCloud,
  setSelectedFiles,
  copySelected,
  moveSelected,
  analyzeSource,
  reloadTagCloud,
} = useFolders()

useHead({ title: 'ต้นทาง | Flash Drive Backup' })

onMounted(() => {
  void reloadTagCloud()
})
</script>

<template>
  <FileList
    title="ไฟล์ต้นทาง"
    :path="source"
    :files="sourceFiles"
    :selected="selectedSourceFiles"
    :busy="busy"
    @selection-change="setSelectedFiles('source', $event)"
  />

  <div class="actions">
    <button class="button" :disabled="busy || !source" @click="analyzeSource">
      {{ busy ? 'กำลังทำงาน...' : 'วิเคราะห์และติดแท็ก' }}
    </button>
    <button
      class="button"
      :disabled="busy || !source || !dest || !selectedSourceFiles.length"
      @click="moveSelected"
    >
      {{ busy ? 'กำลังทำงาน...' : `ย้ายไฟล์ที่เลือก (${selectedSourceFiles.length})` }}
    </button>
    <button
      class="button"
      :disabled="busy || !source || !dest || !selectedSourceFiles.length"
      @click="copySelected"
    >
      {{ busy ? 'กำลังทำงาน...' : `สำรองไฟล์ที่เลือก (${selectedSourceFiles.length})` }}
    </button>
  </div>

  <p v-if="movedCount !== null">ย้ายสำเร็จ {{ movedCount }} ไฟล์</p>
  <p v-if="copiedCount !== null">สำรองสำเร็จ {{ copiedCount }} ไฟล์</p>

  <section class="tag-cloud">
    <h2>แท็กจากรูปภาพ</h2>
    <p v-if="!tagCloud.length">ยังไม่มีแท็ก — กด “วิเคราะห์และติดแท็ก” ก่อน</p>
    <div v-else class="tag-cloud-list">
      <span v-for="tag in tagCloud" :key="tag.name" class="tag-chip">
        {{ tag.name }} <small>{{ tag.count }}</small>
      </span>
    </div>
  </section>
</template>
