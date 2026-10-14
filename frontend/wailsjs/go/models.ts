export namespace dto {
	
	export class AppSettingsDTO {
	    language: string;
	    theme: string;
	    runtimeRootPath: string;
	    defaultProjectPath: string;
	    autoStartServices: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.runtimeRootPath = source["runtimeRootPath"];
	        this.defaultProjectPath = source["defaultProjectPath"];
	        this.autoStartServices = source["autoStartServices"];
	    }
	}
	export class CreateProjectInput {
	    name: string;
	    path: string;
	    phpVersionId?: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateProjectInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.phpVersionId = source["phpVersionId"];
	    }
	}
	export class CreateVirtualHostInput {
	    projectId: number;
	    domain: string;
	    documentRoot: string;
	    webServerType: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateVirtualHostInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.domain = source["domain"];
	        this.documentRoot = source["documentRoot"];
	        this.webServerType = source["webServerType"];
	        this.port = source["port"];
	    }
	}
	export class OperationResultDTO {
	    success: boolean;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new OperationResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class PHPExtensionDTO {
	    name: string;
	    enabled: boolean;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PHPExtensionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.available = source["available"];
	    }
	}
	export class PHPVersionDTO {
	    id: number;
	    version: string;
	    architecture: string;
	    installPath: string;
	    status: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PHPVersionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.architecture = source["architecture"];
	        this.installPath = source["installPath"];
	        this.status = source["status"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class ProjectDTO {
	    id: number;
	    name: string;
	    path: string;
	    phpVersionId?: number;
	    status: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.phpVersionId = source["phpVersionId"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ProjectPathValidationDTO {
	    valid: boolean;
	    normalizedPath: string;
	    exists: boolean;
	    writable: boolean;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectPathValidationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.normalizedPath = source["normalizedPath"];
	        this.exists = source["exists"];
	        this.writable = source["writable"];
	        this.errors = source["errors"];
	    }
	}
	export class RegisterPHPVersionInput {
	    version: string;
	    architecture: string;
	    installPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterPHPVersionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.architecture = source["architecture"];
	        this.installPath = source["installPath"];
	    }
	}
	export class ServiceActionInput {
	    serviceId: number;
	
	    static createFrom(source: any = {}) {
	        return new ServiceActionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	    }
	}
	export class ServiceInstanceDTO {
	    id: number;
	    name: string;
	    serviceType: string;
	    runtimeVersionId: number;
	    port?: number;
	    configPath?: string;
	    dataPath?: string;
	    enabled: boolean;
	    autoStart: boolean;
	    status: string;
	    pid?: number;
	
	    static createFrom(source: any = {}) {
	        return new ServiceInstanceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.serviceType = source["serviceType"];
	        this.runtimeVersionId = source["runtimeVersionId"];
	        this.port = source["port"];
	        this.configPath = source["configPath"];
	        this.dataPath = source["dataPath"];
	        this.enabled = source["enabled"];
	        this.autoStart = source["autoStart"];
	        this.status = source["status"];
	        this.pid = source["pid"];
	    }
	}
	export class ServiceLogInput {
	    serviceId: number;
	    level?: string;
	    limit?: number;
	
	    static createFrom(source: any = {}) {
	        return new ServiceLogInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.level = source["level"];
	        this.limit = source["limit"];
	    }
	}
	export class ServiceLogResultDTO {
	    lines: string[];
	
	    static createFrom(source: any = {}) {
	        return new ServiceLogResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lines = source["lines"];
	    }
	}
	export class ServiceStatusInfoDTO {
	    serviceId: number;
	    status: string;
	    pid?: number;
	    startedAt?: string;
	    lastErrorCode?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceStatusInfoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.status = source["status"];
	        this.pid = source["pid"];
	        this.startedAt = source["startedAt"];
	        this.lastErrorCode = source["lastErrorCode"];
	    }
	}
	export class TaskDTO {
	    id: number;
	    taskType: string;
	    status: string;
	    progress: number;
	    message: string;
	    errorCode?: string;
	    errorMessage?: string;
	    createdAt: string;
	    updatedAt: string;
	    finishedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.taskType = source["taskType"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.message = source["message"];
	        this.errorCode = source["errorCode"];
	        this.errorMessage = source["errorMessage"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.finishedAt = source["finishedAt"];
	    }
	}
	export class UpdateProjectInput {
	    id: number;
	    name?: string;
	    phpVersionId?: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateProjectInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phpVersionId = source["phpVersionId"];
	    }
	}
	export class ValidationResultDTO {
	    valid: boolean;
	    errors: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new ValidationResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	}
	export class VirtualHostDTO {
	    id: number;
	    projectId: number;
	    domain: string;
	    documentRoot: string;
	    webServerType: string;
	    port: number;
	    enabled: boolean;
	    configPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new VirtualHostDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.domain = source["domain"];
	        this.documentRoot = source["documentRoot"];
	        this.webServerType = source["webServerType"];
	        this.port = source["port"];
	        this.enabled = source["enabled"];
	        this.configPath = source["configPath"];
	    }
	}

}

