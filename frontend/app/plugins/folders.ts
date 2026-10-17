export type FolderTarget = 'source' | 'dest'
import {
  AnalyzeFolder,
  CopyFiles,
  CopySelectedFiles,
  DeleteSelectedFiles,
  GetTagCloud,
  GetHistory,
  ListFiles,
  MoveSelectedFiles,
  SetFolder,
} from '~~/wailsjs/go/main/App'

// Nuxt โหลด plugin ให้อัตโนมัติ และให้เรียกผ่าน useNuxtApp().$folders
export default defineNuxtPlugin(() => {
  async function setFolder(target: FolderTarget, input: string): Promise<string> {
    return SetFolder(target, input)
  }

  async function listFiles(_target: FolderTarget, path: string): Promise<string[]> {
    return ListFiles(path)
  }

  async function copyFiles(sourcePath: string, destPath: string): Promise<number> {
    return CopyFiles(sourcePath, destPath)
  }

  async function copySelectedFiles(sourcePath: string, destPath: string, fileNames: string[]) {
    return CopySelectedFiles(sourcePath, destPath, fileNames)
  }

  async function moveSelectedFiles(sourcePath: string, destPath: string, fileNames: string[]) {
    return MoveSelectedFiles(sourcePath, destPath, fileNames)
  }

  async function deleteSelectedFiles(destPath: string, fileNames: string[]) {
    return DeleteSelectedFiles(destPath, fileNames)
  }

  async function getHistory(destPath: string) {
    return GetHistory(destPath)
  }

  async function getTagCloud() {
    return GetTagCloud()
  }

  async function analyzeFolder(path: string) {
    return AnalyzeFolder(path)
  }

  return {
    provide: {
      folders: {
        setFolder,
        listFiles,
        copyFiles,
        copySelectedFiles,
        moveSelectedFiles,
        deleteSelectedFiles,
        getHistory,
        getTagCloud,
        analyzeFolder,
      },
    },
  }
})
