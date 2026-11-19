import { ref } from 'vue'

export const addSkip = ref<string[]>([])
export const removeSkip = ref<number[]>([])

interface Skip {
  id: number | null
  path: string
  description: string
}

interface SkipWithExists extends Skip {
  exists: boolean
}

export class FileNode {
  readonly id: number | null
  readonly name: string
  readonly path: string
  readonly exists: boolean
  readonly type: string
  children: FileNode[] | null

  constructor(
    name: string,
    type = 'file',
    id: number | null = null,
    exists: boolean,
    path: string,
  ) {
    this.id = id
    this.name = name
    this.exists = exists
    this.path = path

    this.type = type // "folder" or "file"
    this.children = type === 'folder' ? [] : null
  }

  add(node: FileNode) {
    if (this.type !== 'folder') {
      throw new Error('Cannot add children to a file.')
    }
    if (this.children === null) {
      throw new Error('Folder children are not initialized.')
    }
    this.children.push(node)
  }
}

export function combineStuff(skips: Skip[], paths: string[]): SkipWithExists[] {
  const pathsSet = new Set(paths)
  const combined = skips.map((skip) => ({
    ...skip,
    exists: pathsSet.has(skip.path),
  }))
  const skipPaths = new Set(skips.map((skip) => skip.path))

  for (const path of paths) {
    if (!skipPaths.has(path)) {
      combined.push({ id: null, path, description: '', exists: true })
      skipPaths.add(path)
    }
  }

  return combined
}

export function buildFileTree(skips: SkipWithExists[]) {
  const tree: FileNode[] = []

  for (const skip of skips) {
    const parts = skip.path.split('/').filter(Boolean)
    let current = tree
    for (const [index, part] of parts.entries()) {
      let node = current.find((n) => n.name === part)
      if (!node) {
        node = new FileNode(
          part,
          index === parts.length - 1 ? 'file' : 'folder',
          skip.id,
          skip.exists,
          skip.path,
        )
        current.push(node)
      }
      current = node.children ?? []
    }
  }

  return tree
}

export function checkboxChange(id: number | null, path: string, checked: boolean) {
  if (checked) {
    if (id === null) {
      addSkip.value.push(path)
    } else {
      const index = removeSkip.value.indexOf(id)
      if (index !== -1) {
        removeSkip.value.splice(index, 1)
      }
    }
  } else {
    if (id !== null) {
      removeSkip.value.push(id)
    } else {
      const index = addSkip.value.indexOf(path)
      if (index !== -1) {
        addSkip.value.splice(index, 1)
      }
    }
  }
}
