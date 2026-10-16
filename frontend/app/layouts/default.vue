<script setup lang="ts">
const { source, dest, busy, error, setFolder } = useFolders()
</script>

<template>
  <div class="program">
    <header class="program-header">
      <h1>Flash Drive Backup</h1>
      <span class="demo-label">ข้อมูลตัวอย่าง</span>
    </header>

    <main>
      <section class="folder-settings" aria-label="ตั้งค่าโฟลเดอร์">
        <FolderForm label="ต้นทาง" :path="source" placeholder="E:\Photos"
          :disabled="busy" @submit="setFolder('source', $event)" />
        <FolderForm label="ปลายทาง" :path="dest" placeholder="D:\Backup"
          :disabled="busy" @submit="setFolder('dest', $event)" />
      </section>

      <p v-if="error" class="error-message" role="alert">{{ error }}</p>

      <nav class="tabs" aria-label="หน้ารายการไฟล์">
        <NuxtLink to="/" exact-active-class="active">ต้นทาง</NuxtLink>
        <NuxtLink to="/destination" active-class="active">ปลายทาง</NuxtLink>
      </nav>

      <slot />
      <p class="demo-note">ข้อมูลตัวอย่าง · ยังไม่อ่านไฟล์จากเครื่อง</p>
    </main>
  </div>
</template>

<!-- CSS ของหน้าบ้านอยู่ในไฟล์ Vue จึงส่งเฉพาะ app/ ได้ -->
<style>
:root {
  font-family: Tahoma, "Segoe UI", sans-serif;
  color: #293b32;
  background: #f5f6f5;
  font-size: 14px;
  font-synthesis: none;
}
* { box-sizing: border-box; }
body { margin: 0; }
button, input { font: inherit; }
button, a, input { -webkit-tap-highlight-color: transparent; }
button:focus-visible, a:focus-visible, input:focus-visible {
  outline: 2px solid #397153; outline-offset: 3px;
}
.program { max-width: 1040px; padding: 0 28px 28px; margin: auto; }
.program-header {
  display: flex; align-items: center; justify-content: space-between;
  gap: 16px; min-height: 84px; border-bottom: 1px solid #dce3de;
}
h1 { font-size: 21px; font-weight: 600; margin: 0; }
.demo-label { font-size: 12px; color: #66776c; white-space: nowrap; }
.folder-settings { display: grid; gap: 14px; margin-top: 24px; }
.folder-form { display: flex; align-items: end; gap: 12px; }
.folder-form label { display: flex; flex: 1; align-items: center; gap: 14px; min-width: 0; }
.folder-form label span { width: 60px; flex-shrink: 0; color: #56665c; }
.folder-form input {
  width: 100%; min-width: 0; height: 40px; padding: 8px 12px;
  background: white; border: 1px solid #cfd9d1; border-radius: 5px;
  color: #293b32; font-family: "Segoe UI", Tahoma, sans-serif;
}
.folder-form input::placeholder { color: #9ba59f; }
.button {
  border: 1px solid #cfd9d1; border-radius: 5px; background: white;
  min-height: 40px; padding: 8px 14px; color: #355c45; cursor: pointer;
  white-space: nowrap;
}
.button:hover:not(:disabled) { background: #edf3ee; border-color: #91ae9a; }
button:disabled, input:disabled { opacity: .5; cursor: not-allowed; }
.tabs { display: flex; gap: 28px; margin: 26px 0 18px; border-bottom: 1px solid #dce3de; }
.tabs a { padding: 12px 2px; text-decoration: none; color: #6f7e74; border-bottom: 2px solid transparent; }
.tabs a.active { color: #326044; border-bottom-color: #397153; }
.file-panel { background: white; border: 1px solid #dce3de; border-radius: 6px; overflow: hidden; }
.panel-header { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 18px; }
h2 { margin: 0; font-size: 15px; font-weight: 600; }
.path { margin: 7px 0 0; color: #6b7a70; font: 13px "Segoe UI", Tahoma, sans-serif; overflow-wrap: anywhere; }
.empty { margin: 0; padding: 85px 18px; text-align: center; color: #7b877f; border-top: 1px solid #edf0ed; }
table { width: 100%; border-collapse: collapse; text-align: left; table-layout: fixed; }
th { padding: 12px 18px; background: #f7f9f7; color: #69796e; font-weight: 400; font-size: 12px; border-block: 1px solid #e6ebe7; }
td { padding: 16px 18px; border-bottom: 1px solid #edf0ed; font-family: "Segoe UI", Tahoma, sans-serif; overflow-wrap: anywhere; }
tbody tr:last-child td { border-bottom: 0; }
.panel-footer { border-top: 1px solid #e6ebe7; padding: 14px 18px; color: #718075; font-size: 12px; }
.demo-note { font-size: 12px; color: #7d897f; line-height: 1.8; margin: 15px 0 0; }
.error-message { color: #934b3d; border: 1px solid #e7cec6; border-radius: 5px; background: #fff4ef; padding: 12px; overflow-wrap: anywhere; line-height: 1.7; }
@media (max-width: 520px) {
  .program { padding-inline: 16px; }
  .program-header { min-height: 74px; gap: 8px; }
  h1 { font-size: 18px; }
  .demo-label { font-size: 10px; }
  .folder-form { gap: 8px; }
  .folder-form label { gap: 8px; }
  .folder-form label span { width: 49px; font-size: 12px; }
  .folder-form input { font-size: 13px; padding-inline: 8px; }
  .button { padding-inline: 10px; font-size: 12px; }
  .panel-header { padding: 14px; gap: 12px; align-items: start; }
  th, td, .panel-footer { padding-inline: 14px; }
}

</style>
