<script setup lang="ts">
import { watch } from 'vue'

const { dest, historyRecords, historyLoading, reloadHistory } = useFolders()
useHead({ title: 'ประวัติ | Flash Drive Backup' })

watch(dest, reloadHistory, { immediate: true })

function statusText(status: string) {
  if (status === 'moved') return 'ย้ายแล้ว'
  if (status === 'copied') return 'สำรองแล้ว'
  return status
}

function formatDate(value: unknown) {
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleString()
}
</script>

<template>
  <section class="file-panel" :aria-busy="historyLoading">
    <header class="panel-header">
      <div>
        <h2>ประวัติการทำงาน</h2>
        <p v-if="dest" class="path">ปลายทาง: {{ dest }}</p>
      </div>
    </header>

    <p v-if="!dest" class="empty">กรุณาตั้งค่าโฟลเดอร์ปลายทางก่อน</p>
    <p v-else-if="!historyLoading && !historyRecords.length" class="empty">ยังไม่มีประวัติในฐานข้อมูล</p>
    <p v-else-if="historyLoading && !historyRecords.length" class="empty">กำลังโหลดประวัติ...</p>

    <table v-else>
      <thead>
        <tr>
          <th scope="col">ชื่อไฟล์</th>
          <th scope="col">ต้นทาง</th>
          <th scope="col">สถานะ</th>
          <th scope="col">เวลาที่บันทึก</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="record in historyRecords" :key="record.ID">
          <td>{{ record.FileName }}</td>
          <td>{{ record.SourcePath }}</td>
          <td>{{ statusText(record.Status) }}</td>
          <td>{{ formatDate(record.CreatedAt) }}</td>
        </tr>
      </tbody>
    </table>

    <footer v-if="dest" class="panel-footer">{{ historyRecords.length }} รายการ</footer>
  </section>
</template>
