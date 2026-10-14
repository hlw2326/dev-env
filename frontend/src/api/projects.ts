import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { Project, CreateProjectInput, UpdateProjectInput, ProjectPathValidation, OperationResult } from '../types'

export async function listProjects(): Promise<Project[]> {
  const items = await AppAPI.ListProjects()
  return (items || []).map((p) => ({
    id: p.id,
    name: p.name,
    path: p.path,
    phpVersionId: p.phpVersionId ?? null,
    status: p.status as Project['status'],
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
  }))
}

export async function getProject(id: number): Promise<Project> {
  const p = await AppAPI.GetProject(id)
  return {
    id: p.id,
    name: p.name,
    path: p.path,
    phpVersionId: p.phpVersionId ?? null,
    status: p.status as Project['status'],
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
  }
}

export async function createProject(input: CreateProjectInput): Promise<Project> {
  const p = await AppAPI.CreateProject({
    name: input.name,
    path: input.path,
    phpVersionId: input.phpVersionId ?? undefined,
  })
  return {
    id: p.id,
    name: p.name,
    path: p.path,
    phpVersionId: p.phpVersionId ?? null,
    status: p.status as Project['status'],
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
  }
}

export async function importProject(path: string): Promise<Project> {
  const p = await AppAPI.ImportProject(path)
  return {
    id: p.id,
    name: p.name,
    path: p.path,
    phpVersionId: p.phpVersionId ?? null,
    status: p.status as Project['status'],
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
  }
}

export async function updateProject(input: UpdateProjectInput): Promise<Project> {
  const p = await AppAPI.UpdateProject({
    id: input.id,
    name: input.name,
    phpVersionId: input.phpVersionId ?? undefined,
  })
  return {
    id: p.id,
    name: p.name,
    path: p.path,
    phpVersionId: p.phpVersionId ?? null,
    status: p.status as Project['status'],
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
  }
}

export async function deleteProject(id: number): Promise<OperationResult> {
  return await AppAPI.DeleteProject(id)
}

export async function validateProjectPath(path: string): Promise<ProjectPathValidation> {
  return await AppAPI.ValidateProjectPath(path)
}
