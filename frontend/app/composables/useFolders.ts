import type { FolderTarget } from '~/plugins/folders'
import type { main, model } from '~~/wailsjs/go/models'

// useState ทำให้ข้อมูลคงอยู่เมื่อสลับหน้า และแยกข้อมูลแต่ละผู้ใช้เมื่อรัน SSR
export function useFolders() {
  // plugin เป็นจุดเรียกข้อมูล ส่วน composable เก็บข้อมูลที่หน้าจอใช้ร่วมกัน
  const { $folders: folders } = useNuxtApp()
  const source = useState('folders:source', () => '')
  const dest = useState('folders:dest', () => '')
  const sourceFiles = useState<string[]>('folders:sourceFiles', () => [])
  const destFiles = useState<string[]>('folders:destFiles', () => [])
  const selectedSourceFiles = useState<string[]>('folders:selectedSourceFiles', () => [])
  const selectedDestFiles = useState<string[]>('folders:selectedDestFiles', () => [])
  const busy = useState('folders:busy', () => false)
  const error = useState('folders:error', () => '')
  const copiedCount = useState<number | null>('folders:copiedCount', () => null)
  const movedCount = useState<number | null>('folders:movedCount', () => null)
  const deletedCount = useState<number | null>('folders:deletedCount', () => null)
  const historyRecords = useState<model.BackupRecord[]>('folders:history', () => [])
  const historyDestination = useState('folders:historyDestination', () => '')
  const historyRequest = useState('folders:historyRequest', () => 0)
  const historyLoading = useState('folders:historyLoading', () => false)
  const tagCloud = useState<model.TagCount[]>('folders:tagCloud', () => [])
  const analyzeResults = useState<main.ImageTags[]>('folders:analyzeResults', () => [])

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
        selectedSourceFiles.value = []
      } else {
        dest.value = path
        destFiles.value = []
        selectedDestFiles.value = []
      }

      const files = await folders.listFiles(target, path)
      if (target === 'source') sourceFiles.value = files
      else destFiles.value = files
    })
  }

  function setSelectedFiles(target: FolderTarget, files: string[]) {
    if (target === 'source') selectedSourceFiles.value = files
    else selectedDestFiles.value = files
  }

  async function reloadFiles(target: FolderTarget) {
    await run(async () => {
      setSelectedFiles(target, [])
      const path = target === 'source' ? source.value : dest.value
      const names = await folders.listFiles(target, path)
      if (target === 'source') sourceFiles.value = names
      else destFiles.value = names
    })
  }

  async function refreshFiles(targets: FolderTarget[]) {
    const results = await Promise.allSettled(targets.map(async (target) => {
      const path = target === 'source' ? source.value : dest.value
      const names = await folders.listFiles(target, path)
      if (target === 'source') sourceFiles.value = names
      else destFiles.value = names
    }))

    const failed = results.find(result => result.status === 'rejected')
    if (failed?.status === 'rejected') throw failed.reason
  }

  async function runFileAction(
    fileNames: string[],
    selectionTarget: FolderTarget,
    refreshTargets: FolderTarget[],
    action: (names: string[]) => Promise<void>,
  ) {
    await run(async () => {
      if (!fileNames.length) throw new Error('กรุณาเลือกไฟล์ก่อน')

      let actionError: unknown
      let actionFailed = false
      try {
        await action([...fileNames])
      } catch (cause) {
        actionFailed = true
        actionError = cause
      }

      setSelectedFiles(selectionTarget, [])

      let refreshError: unknown
      let refreshFailed = false
      try {
        await refreshFiles(refreshTargets)
      } catch (cause) {
        refreshFailed = true
        refreshError = cause
      }

      if (actionFailed) throw actionError
      if (refreshFailed) throw refreshError
    })
  }

  async function copySelected() {
    copiedCount.value = null
    movedCount.value = null
    deletedCount.value = null
    await runFileAction(selectedSourceFiles.value, 'source', ['source', 'dest'], async (fileNames) => {
      copiedCount.value = await folders.copySelectedFiles(source.value, dest.value, fileNames)
    })
  }

  async function moveSelected() {
    copiedCount.value = null
    movedCount.value = null
    deletedCount.value = null
    await runFileAction(selectedSourceFiles.value, 'source', ['source', 'dest'], async (fileNames) => {
      movedCount.value = await folders.moveSelectedFiles(source.value, dest.value, fileNames)
    })
  }

  async function deleteSelected() {
    copiedCount.value = null
    movedCount.value = null
    deletedCount.value = null
    await runFileAction(selectedDestFiles.value, 'dest', ['dest'], async (fileNames) => {
      deletedCount.value = await folders.deleteSelectedFiles(dest.value, fileNames)
    })
  }

  async function reloadTagCloud() {
    try {
      tagCloud.value = (await folders.getTagCloud()) ?? []
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'โหลด Tag Cloud ไม่สำเร็จ'
    }
  }

  async function analyzeSource() {
    await run(async () => {
      if (!source.value) throw new Error('กรุณาตั้งค่าโฟลเดอร์ต้นทางก่อน')

      analyzeResults.value = await folders.analyzeFolder(source.value)
      tagCloud.value = (await folders.getTagCloud()) ?? []
    })
  }

  async function reloadHistory() {
    const requestNumber = ++historyRequest.value
    if (!dest.value) {
      historyRecords.value = []
      historyDestination.value = ''
      error.value = ''
      historyLoading.value = false
      return
    }

    const requestedDestination = dest.value
    if (historyDestination.value !== requestedDestination) historyRecords.value = []
    historyLoading.value = true
    error.value = ''
    try {
      const records = await folders.getHistory(requestedDestination)
      if (requestNumber === historyRequest.value) {
        historyRecords.value = records
        historyDestination.value = requestedDestination
      }
    } catch (cause) {
      if (requestNumber === historyRequest.value) {
        historyRecords.value = []
        error.value = cause instanceof Error ? cause.message : 'อ่านประวัติไม่สำเร็จ'
      }
    } finally {
      if (requestNumber === historyRequest.value) historyLoading.value = false
    }
  }

  return {
    source,
    dest,
    sourceFiles,
    destFiles,
    selectedSourceFiles,
    selectedDestFiles,
    busy,
    error,
    copiedCount,
    movedCount,
    deletedCount,
    historyRecords,
    historyLoading,
    tagCloud,
    analyzeResults,
    setFolder,
    setSelectedFiles,
    reloadFiles,
    copySelected,
    moveSelected,
    deleteSelected,
    reloadTagCloud,
    analyzeSource,
    reloadHistory,
  }
}
