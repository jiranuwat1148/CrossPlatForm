<script setup lang="ts">
defineProps<{ title: string; path: string; files: string[]; busy: boolean }>()
defineEmits<{ reload: [] }>()
</script>

<template>
  <section class="file-panel" :aria-label="title" :aria-busy="busy">
    <header class="panel-header">
      <div>
        <h2>{{ title }}</h2>
        <p v-if="path" class="path">{{ path }}</p>
      </div>
      <button class="button" :disabled="busy || !path" @click="$emit('reload')">
        {{ busy ? 'กำลังโหลด...' : 'โหลดรายการใหม่' }}
      </button>
    </header>

    <p v-if="!path" class="empty">กรุณาตั้งค่าโฟลเดอร์ก่อน</p>
    <p v-else-if="!files.length" class="empty">ไม่มีไฟล์ในโฟลเดอร์นี้</p>
    <table v-else>
      <thead><tr><th scope="col">ชื่อไฟล์</th></tr></thead>
      <tbody><tr v-for="name in files" :key="name"><td>{{ name }}</td></tr></tbody>
    </table>

    <footer v-if="path" class="panel-footer">{{ files.length }} ไฟล์</footer>
  </section>
</template>
