import { defineStore } from 'pinia'
import type { Project } from '@/types'
import type { ProjectsState } from './types.d'
import { listProjects } from '@/api/projects'

export const useProjectsStore = defineStore('projects', {
  state: (): ProjectsState => ({
    list: [],
    loading: false,
    selectedId: null
  }),

  getters: {
    selectedProject: (state): Project | undefined =>
      state.list.find(p => p.id === state.selectedId)
  },

  actions: {
    async fetchProjects() {
      this.loading = true
      try {
        this.list = await listProjects()
      } finally {
        this.loading = false
      }
    },

    selectProject(id: number) {
      this.selectedId = id
    }
  },

  persist: {
    key: 'devenv-projects',
    paths: ['selectedId']
  }
})

export * from './types.d'
