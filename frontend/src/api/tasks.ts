import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { Task, OperationResult } from '../types'

export async function listTasks(): Promise<Task[]> {
  const items = await AppAPI.ListTasks()
  return (items || []).map((t) => ({
    id: t.id,
    taskType: t.taskType,
    status: t.status as Task['status'],
    progress: t.progress,
    message: t.message,
    errorCode: t.errorCode ?? null,
    errorMessage: t.errorMessage ?? null,
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
    finishedAt: t.finishedAt ?? null,
  }))
}

export async function getTask(taskId: number): Promise<Task> {
  const t = await AppAPI.GetTask(taskId)
  return {
    id: t.id,
    taskType: t.taskType,
    status: t.status as Task['status'],
    progress: t.progress,
    message: t.message,
    errorCode: t.errorCode ?? null,
    errorMessage: t.errorMessage ?? null,
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
    finishedAt: t.finishedAt ?? null,
  }
}

export async function cancelTask(taskId: number): Promise<OperationResult> {
  return await AppAPI.CancelTask(taskId)
}
