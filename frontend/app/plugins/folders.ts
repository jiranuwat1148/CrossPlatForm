export type FolderTarget = 'source' | 'dest'

// Nuxt โหลด plugin ให้อัตโนมัติ และให้เรียกผ่าน useNuxtApp().$folders
export default defineNuxtPlugin(() => {
  // ข้อมูลตัวอย่างสำหรับดูหน้าบ้าน ไม่มีการอ่านหรือแก้ไฟล์จริง
  const demoFolders = [
    { path: String.raw`E:\Photos`, files: ['mountain.jpg', 'notes.txt', 'ocean.png'] },
    { path: String.raw`D:\Backup`, files: ['report.pdf'] },
    { path: String.raw`D:\Empty`, files: [] },
    { path: '/home/demo/Pictures', files: ['mountain.jpg', 'notes.txt', 'ocean.png'] },
    { path: '/home/demo/Backup', files: ['report.pdf'] },
  ]

  function findFolder(input: string) {
    const path = input.trim().replace(/^"(.*)"$/, '$1')
    if (!path) throw new Error('กรุณาระบุ path')
    const windowsPath = /^[a-z]:[\\/]/i.test(path)
    const comparable = windowsPath ? path.replaceAll('/', '\\').toLowerCase() : path
    const folder = demoFolders.find((item) =>
      windowsPath ? item.path.toLowerCase() === comparable : item.path === comparable,
    )
    if (!folder) throw new Error(`เปิดโฟลเดอร์ไม่ได้: ${path}`)
    return folder
  }

  // จุดเชื่อม Go: เปลี่ยนเนื้อหาให้เรียก SetSource/SetDest ที่เพื่อนสร้าง
  async function setFolder(_target: FolderTarget, input: string): Promise<string> {
    return findFolder(input).path
  }

  // จุดเชื่อม Go: เปลี่ยนเนื้อหาให้เรียก ListFiles และคืนชื่อไฟล์เป็นอาร์เรย์
  async function listFiles(_target: FolderTarget, path: string): Promise<string[]> {
    if (!path) throw new Error('กรุณาตั้งค่าโฟลเดอร์ก่อน')
    return [...findFolder(path).files].sort()
  }

  return {
    provide: {
      folders: { setFolder, listFiles },
    },
  }
})
