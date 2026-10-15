import * as ProjectAPI from '../../wailsjs/go/wails/ProjectAPI'

export const getDeploymentConfig = ProjectAPI.DeploymentConfig
export const saveDeploymentConfig = ProjectAPI.SaveDeployment
export const previewDeployment = ProjectAPI.PreviewDeployment
export const testDeployment = ProjectAPI.TestDeployment
export const startDeployment = ProjectAPI.StartDeployment
export const cancelDeployment = ProjectAPI.CancelDeployment
export const getDeploymentHistory = ProjectAPI.DeploymentHistory
