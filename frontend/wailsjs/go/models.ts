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
	export class ConfCheckDTO {
	    valid: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfCheckDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	    }
	}
	export class CreateSiteInput {
	    name: string;
	    kind: string;
	    domains: string[];
	    root: string;
	    runtimeId?: number;
	    serviceId?: number;
	    autoService: boolean;
	    git: boolean;
	    remote: string;
	    branch: string;
	    port: number;
	    tls: boolean;
	    issuer: string;
	    cert: string;
	    key: string;
	    upstream?: number;
	    entry: string;
	    cmd: string;
	    autoStart: boolean;
	    startNow: boolean;
	    custom: string;
	    env?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new CreateSiteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.domains = source["domains"];
	        this.root = source["root"];
	        this.runtimeId = source["runtimeId"];
	        this.serviceId = source["serviceId"];
	        this.autoService = source["autoService"];
	        this.git = source["git"];
	        this.remote = source["remote"];
	        this.branch = source["branch"];
	        this.port = source["port"];
	        this.tls = source["tls"];
	        this.issuer = source["issuer"];
	        this.cert = source["cert"];
	        this.key = source["key"];
	        this.upstream = source["upstream"];
	        this.entry = source["entry"];
	        this.cmd = source["cmd"];
	        this.autoStart = source["autoStart"];
	        this.startNow = source["startNow"];
	        this.custom = source["custom"];
	        this.env = source["env"];
	    }
	}
	export class DeleteSiteInput {
	    id: number;
	    purge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeleteSiteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.purge = source["purge"];
	    }
	}
	export class GitCommitInput {
	    siteId: number;
	    message: string;
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new GitCommitInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.siteId = source["siteId"];
	        this.message = source["message"];
	        this.files = source["files"];
	    }
	}
	export class GitStatusDTO {
	    branch: string;
	    commit: string;
	    message: string;
	    clean: boolean;
	    ahead: number;
	    behind: number;
	    modified: string[];
	    untracked: string[];
	    checkedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new GitStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.branch = source["branch"];
	        this.commit = source["commit"];
	        this.message = source["message"];
	        this.clean = source["clean"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.modified = source["modified"];
	        this.untracked = source["untracked"];
	        this.checkedAt = source["checkedAt"];
	    }
	}
	export class InstallResultDTO {
	    runtimeId: number;
	    serviceId?: number;
	    serviceCreated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InstallResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runtimeId = source["runtimeId"];
	        this.serviceId = source["serviceId"];
	        this.serviceCreated = source["serviceCreated"];
	    }
	}
	export class InstallRuntimeInput {
	    runtimeType: string;
	    version: string;
	    variant: string;
	    architecture: string;
	    downloadUrl: string;
	    installPath: string;
	    mirror: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallRuntimeInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runtimeType = source["runtimeType"];
	        this.version = source["version"];
	        this.variant = source["variant"];
	        this.architecture = source["architecture"];
	        this.downloadUrl = source["downloadUrl"];
	        this.installPath = source["installPath"];
	        this.mirror = source["mirror"];
	    }
	}
	export class LogResultDTO {
	    lines: string[];
	
	    static createFrom(source: any = {}) {
	        return new LogResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lines = source["lines"];
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
	export class RegisterRuntimeInput {
	    runtimeType: string;
	    version: string;
	    variant: string;
	    architecture: string;
	    installPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterRuntimeInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runtimeType = source["runtimeType"];
	        this.version = source["version"];
	        this.variant = source["variant"];
	        this.architecture = source["architecture"];
	        this.installPath = source["installPath"];
	    }
	}
	export class RuntimeDTO {
	    id: number;
	    runtimeType: string;
	    version: string;
	    variant: string;
	    architecture: string;
	    installPath: string;
	    status: string;
	    isDefault: boolean;
	    siteCount: number;
	    serviceCount: number;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.runtimeType = source["runtimeType"];
	        this.version = source["version"];
	        this.variant = source["variant"];
	        this.architecture = source["architecture"];
	        this.installPath = source["installPath"];
	        this.status = source["status"];
	        this.isDefault = source["isDefault"];
	        this.siteCount = source["siteCount"];
	        this.serviceCount = source["serviceCount"];
	    }
	}
	export class RuntimeOptionDTO {
	    id: number;
	    version: string;
	    variant: string;
	    architecture: string;
	    installPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeOptionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.variant = source["variant"];
	        this.architecture = source["architecture"];
	        this.installPath = source["installPath"];
	    }
	}
	export class SaveConfInput {
	    siteId: number;
	    content: string;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveConfInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.siteId = source["siteId"];
	        this.content = source["content"];
	        this.summary = source["summary"];
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
	    runtimeId: number;
	    runtimeVersion: string;
	    runtimePath: string;
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
	        this.runtimeId = source["runtimeId"];
	        this.runtimeVersion = source["runtimeVersion"];
	        this.runtimePath = source["runtimePath"];
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
	export class ServiceStatusInfoDTO {
	    serviceId: number;
	    status: string;
	    pid?: number;
	    startedAt?: string;
	    errorCode?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceStatusInfoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.status = source["status"];
	        this.pid = source["pid"];
	        this.startedAt = source["startedAt"];
	        this.errorCode = source["errorCode"];
	    }
	}
	export class SiteDTO {
	    id: number;
	    name: string;
	    kind: string;
	    domains: string[];
	    root: string;
	    runtimeId?: number;
	    serviceId?: number;
	    git: boolean;
	    remote: string;
	    branch: string;
	    commit: string;
	    pulledAt: string;
	    conf: string;
	    port: number;
	    tls: boolean;
	    cert: string;
	    key: string;
	    issuer: string;
	    proxy: boolean;
	    upstream?: number;
	    meta: site.Meta;
	    status: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SiteDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.domains = source["domains"];
	        this.root = source["root"];
	        this.runtimeId = source["runtimeId"];
	        this.serviceId = source["serviceId"];
	        this.git = source["git"];
	        this.remote = source["remote"];
	        this.branch = source["branch"];
	        this.commit = source["commit"];
	        this.pulledAt = source["pulledAt"];
	        this.conf = source["conf"];
	        this.port = source["port"];
	        this.tls = source["tls"];
	        this.cert = source["cert"];
	        this.key = source["key"];
	        this.issuer = source["issuer"];
	        this.proxy = source["proxy"];
	        this.upstream = source["upstream"];
	        this.meta = this.convertValues(source["meta"], site.Meta);
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SiteLogInput {
	    siteId: number;
	    kind: string;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new SiteLogInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.siteId = source["siteId"];
	        this.kind = source["kind"];
	        this.limit = source["limit"];
	    }
	}
	export class SwitchRuntimeInput {
	    serviceId: number;
	    runtimeId: number;
	
	    static createFrom(source: any = {}) {
	        return new SwitchRuntimeInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.runtimeId = source["runtimeId"];
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
	export class UpdateSiteInput {
	    id: number;
	    name?: string;
	    domains?: string[];
	    root?: string;
	    runtimeId?: number;
	    serviceId?: number;
	    port?: number;
	    tls?: boolean;
	    issuer?: string;
	    cert?: string;
	    key?: string;
	    upstream?: number;
	    entry?: string;
	    cmd?: string;
	    custom?: string;
	    autoStart?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSiteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.domains = source["domains"];
	        this.root = source["root"];
	        this.runtimeId = source["runtimeId"];
	        this.serviceId = source["serviceId"];
	        this.port = source["port"];
	        this.tls = source["tls"];
	        this.issuer = source["issuer"];
	        this.cert = source["cert"];
	        this.key = source["key"];
	        this.upstream = source["upstream"];
	        this.entry = source["entry"];
	        this.cmd = source["cmd"];
	        this.custom = source["custom"];
	        this.autoStart = source["autoStart"];
	    }
	}

}

export namespace site {
	
	export class Snapshot {
	    time: string;
	    summary: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	    }
	}
	export class Meta {
	    entry: string;
	    cmd: string;
	    autoStart: boolean;
	    env: Record<string, string>;
	    custom: string;
	    accessLog: string;
	    errorLog: string;
	    history: Snapshot[];
	
	    static createFrom(source: any = {}) {
	        return new Meta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entry = source["entry"];
	        this.cmd = source["cmd"];
	        this.autoStart = source["autoStart"];
	        this.env = source["env"];
	        this.custom = source["custom"];
	        this.accessLog = source["accessLog"];
	        this.errorLog = source["errorLog"];
	        this.history = this.convertValues(source["history"], Snapshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

