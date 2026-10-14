import { ref, computed } from 'vue'
import { useProjectsStore } from '@/store/projects'
import { createProject, deleteProject } from '@/api/projects'
import type { CreateProjectParams } from './types'

export * from './types.d'

export function useProjects() {
  const store = useProjectsStore()

  const creating = ref(false)
  const deleting = ref(false)

  const projects = computed(() => store.list)
  const loading = computed(() => store.loading)

  async function loadProjects() {
    await store.fetchProjects()
  }

  async function addProject(data: CreateProjectParams) {
    creating.value = true
    try {
      const res = await createProject(data)
      await loadProjects()
      toast.success(`项目 ${data.name} 已添加`)
      return res
    } finally {
      creating.value = false
    }
  }

  async function removeProject(id: number) {
    deleting.value = true
    try {
      await deleteProject(id)
      await loadProjects()
      toast.info('项目记录已移除')
    } finally {
      deleting.value = false
    }
  }

  return {
    projects,
    loading,
    creating,
    deleting,
    loadProjects,
    addProject,
    removeProject
  }
}
