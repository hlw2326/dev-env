import type { Project } from '@/types'

export interface ProjectsState {
  list: Project[]
  loading: boolean
  selectedId: number | null
}
