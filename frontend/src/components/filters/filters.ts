import type { FlowExportObject } from '@vue-flow/core'
import Tooltip from '../Tooltip.vue'

export const filters = [
  {
    type: 'number',
    tooltip: 'Skip if file is newer than this number of days',
    data: { id: 'fileAge', label: 'File Age', placeholder: 'Days', skipFuture: false },
  },
  {
    type: 'number',
    tooltip: 'Skip if file is smaller than this size in MB',
    data: { id: 'minimumFileSize', label: 'Minimum Size', placeholder: 'MB', skipFuture: true },
  },
  { type: 'noinput', tooltip: 'Skip if file has hardlinks', data: { id: 'hardlinks', label: 'Hardlinks', skipFuture: false } },
  { type: 'noinput', tooltip: 'Skip if the original file size is smaller than the transcoded file', data: { id: 'newFileSize', label: 'Original File Size', skipFuture: true } },
  {
    type: 'codec',
    tooltip: 'Skip if media codec is in the selected list',
    data: { id: 'mediaCodec', label: 'Media Codec', mediaCodecs: [], skipFuture: true },
  },
]

export type OrderedFilter = {
  id: string
  data_int?: number
  data_array?: string[]
  skipFuture?: boolean
}

export function createNodeId(): string {
  return `node-${Math.random().toString(36).slice(2, 10)}`
}

export function genOrder(stuff: FlowExportObject): OrderedFilter[] | string {
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
  if (!order.some((item) => item.id === 'transcode')) return 'Flow must include the Transcode node'

  return order.slice(1, -1)
}

export function genNodes(order: OrderedFilter[]): FlowExportObject['nodes'] {
  const nodes: FlowExportObject['nodes'] = [
    {
      id: '1',
      type: 'input',
      position: { x: 100, y: 0 },
      deletable: false,
      data: { label: 'Start', id: 'start' },
    },
  ]

  if (order.length === 0) {
    nodes.push({
      id: '2',
      type: 'default',
      position: { x: 100, y: 100 },
      deletable: false,
      style: { border: '2px solid #22c55e' },
      data: { label: 'Transcode', id: 'transcode' },
    })
  }

  let pos = 75

  for (const item of order) {
    if (item.id === 'transcode') {
      nodes.push({
        id: '2',
        type: 'default',
        position: { x: 100, y: pos },
        deletable: false,
        style: { border: '2px solid #22c55e' },
        data: { label: 'Transcode', id: 'transcode' },
      })

      pos += 75
    } else {
      const filter = filters.find((f) => f.data.id === item.id)
      if (filter) {
        nodes.push({
          id: createNodeId(),
          type: filter.type,
          position: { x: 100, y: pos },
          data: {
            ...filter.data,
            num: item.data_int,
            mediaCodecs: item.data_array,
            skipFuture: item.skipFuture,
          },
        })

        switch (filter.type) {
          case 'number':
            pos += 150

            break
          case 'codec':
            pos += 200
            break
          default:
            pos += 100
        }
      }
    }
  }

  nodes.push({
    id: '3',
    type: 'output',
    position: { x: 100, y: pos },
    deletable: false,
    data: { label: 'End', id: 'end' },
  })

  return nodes
}

export function genEdges(nodes: FlowExportObject['nodes']): FlowExportObject['edges'] {
  const edges: FlowExportObject['edges'] = []

  for (let i = 0; i < nodes.length - 1; i++) {
    const currentNode = nodes[i]
    const nextNode = nodes[i + 1]
    if (currentNode && nextNode) {
      edges.push({
        id: `${currentNode.id}-${nextNode.id}`,
        source: currentNode.id,
        target: nextNode.id,
        sourceHandle: null,
        targetHandle: null,
      })
    }
  }

  return edges
}
