import { ref, computed } from 'vue'
import { useProjectsStore } from '@/store/projects'
import { useToast } from 'primevue/usetoast'
import { createProject, deleteProject } from '@/api/projects'

export function useProjects() {
  const store = useProjectsStore()
  const toast = useToast()

  const creating = ref(false)
  const deleting = ref(false)

  const projects = computed(() => store.list)
  const loading = computed(() => store.loading)

  async function loadProjects() {
    await store.fetchProjects()
  }

  async function addProject(data: { name: string; path: string; phpVersion?: string; webServer?: string }) {
    creating.value = true
    try {
      const res = await createProject(data)
      await loadProjects()
      toast.add({
        severity: 'success',
        summary: '创建成功',
        detail: `项目 ${data.name} 已添加`,
        life: 3000
      })
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
      toast.add({
        severity: 'info',
        summary: '删除成功',
        detail: '项目记录已移除',
        life: 3000
      })
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
