import type { FolderTarget } from '~/plugins/folders'

// useState ทำให้ข้อมูลคงอยู่เมื่อสลับหน้า และแยกข้อมูลแต่ละผู้ใช้เมื่อรัน SSR
export function useFolders() {
  // plugin เป็นจุดเรียกข้อมูล ส่วน composable เก็บข้อมูลที่หน้าจอใช้ร่วมกัน
  const { $folders: folders } = useNuxtApp()
  const source = useState('folders:source', () => '')
  const dest = useState('folders:dest', () => '')
  const sourceFiles = useState<string[]>('folders:sourceFiles', () => [])
  const destFiles = useState<string[]>('folders:destFiles', () => [])
  const busy = useState('folders:busy', () => false)
  const error = useState('folders:error', () => '')

  async function run(task: () => Promise<void>) {
    if (busy.value) return
    busy.value = true
    error.value = ''
    try {
      await task()
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'ทำงานไม่สำเร็จ'
    } finally {
      busy.value = false
    }
  }

async function setFolder(target: FolderTarget, input: string) {
  await run(async () => {
    const path = await folders.setFolder(target, input)

    if (target === 'source') {
      source.value = path
      sourceFiles.value = []
    } else {
      dest.value = path
      destFiles.value = []
    }
  })
}


  async function reloadFiles(target: FolderTarget) {
    await run(async () => {
      const path = target === 'source' ? source.value : dest.value
      const names = await folders.listFiles(target, path)
      if (target === 'source') sourceFiles.value = names
      else destFiles.value = names
    })
  }

  return { source, dest, sourceFiles, destFiles, busy, error, setFolder, reloadFiles }
}
