import * as TaskAPI from '../../wailsjs/go/wails/TaskAPI'
import type { Task, OperationResult } from '../types'

export async function listTasks(): Promise<Task[]> {
  const items = await TaskAPI.List()
  return items.map((t) => ({
    id: t.id,
    taskType: t.taskType,
    status: t.status as Task['status'],
    progress: t.progress,
    message: t.message,
    errorCode: t.errorCode,
    errorMessage: t.errorMessage,
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
    finishedAt: t.finishedAt,
  }))
}

export async function getTask(taskId: number): Promise<Task> {
  const t = await TaskAPI.Get(taskId)
  return {
    id: t.id,
    taskType: t.taskType,
    status: t.status as Task['status'],
    progress: t.progress,
    message: t.message,
    errorCode: t.errorCode,
    errorMessage: t.errorMessage,
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
    finishedAt: t.finishedAt,
  }
}

export async function cancelTask(taskId: number): Promise<OperationResult> {
  return await TaskAPI.Cancel(taskId)
}
