import type { FlowExportObject } from '@vue-flow/core'

export type SaveResult =
  | { nodes: FlowExportObject['nodes']; edges: FlowExportObject['edges']; order: OrderedFilter[] }
  | { error: string }

export function save(stuff: FlowExportObject): SaveResult {
  const result = genOrder(stuff)
  if (typeof result === 'string') return { error: result }
  return { nodes: stuff.nodes, edges: stuff.edges, order: result }
}

export type OrderedFilter = {
  id: string
  data_int?: number
  data_array?: string[]
  skipFuture?: boolean
}

function genOrder(stuff: FlowExportObject): OrderedFilter[] | string {
  const nextNode = new Map(stuff.edges.map((e) => [e.source, e.target]))
  const hasIncoming = new Set(stuff.edges.map((e) => e.target))
  const nodeById = new Map(stuff.nodes.map((n) => [n.id, n]))

  const startNode = stuff.nodes.find((n) => !hasIncoming.has(n.id))
  if (!startNode) return 'No start node found'

  const order: OrderedFilter[] = []
  let current: string | undefined = startNode.id
  while (current) {
    const node = nodeById.get(current)
    if (node?.data?.id) {
      const item: OrderedFilter = { id: node.data.id }

      if (node.data.skipFuture === true) item.skipFuture = true

      switch (node.type) {
        case 'number':
          if (typeof node.data.num === 'number') item.data_int = node.data.num
          break
        case 'codec':
          if (Array.isArray(node.data.mediaCodecs)) item.data_array = node.data.mediaCodecs
          break
      }

      order.push(item)
    }
    current = nextNode.get(current)
  }

  if (order[0]?.id !== 'start')
    return 'Flow must begin with the Start node and end with the End node'
  if (order[order.length - 1]?.id !== 'end')
    return 'Flow must begin with the Start node and end with the End node'
  if (!order.some((item) => item.id === 'transcode'))
    return 'Flow must include the Transcode node'

  return order.slice(1, -1)
}
