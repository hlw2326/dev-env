export namespace wails {
	
	export class ConfigItem {
	    key: string;
	    label: string;
	    value: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.value = source["value"];
	        this.description = source["description"];
	    }
	}
	export class ConfigDTO {
	    items: ConfigItem[];
	    file: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], ConfigItem);
	        this.file = source["file"];
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
	export class ConfigInput {
	    id: number;
	    items: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ConfigInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.items = source["items"];
	    }
	}
	
	export class ConfigResult {
	    ok: boolean;
	    message: string;
	    restart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConfigResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.restart = source["restart"];
	    }
	}
	export class DeploymentConfigDTO {
	    projectId: number;
	    provider: string;
	    source: string;
	    exclude: string;
	    host: string;
	    port: number;
	    username: string;
	    privateKey: string;
	    fingerprint: string;
	    remote: string;
	    appId?: string;
	    appIds: string[];
	    developerToolsCLI: string;
	    version: string;
	    description: string;
	    robot: number;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.provider = source["provider"];
	        this.source = source["source"];
	        this.exclude = source["exclude"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.privateKey = source["privateKey"];
	        this.fingerprint = source["fingerprint"];
	        this.remote = source["remote"];
	        this.appId = source["appId"];
	        this.appIds = source["appIds"];
	        this.developerToolsCLI = source["developerToolsCLI"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.robot = source["robot"];
	    }
	}
	export class DeploymentFileDTO {
	    path: string;
	    size: number;
	    digest: string;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentFileDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.digest = source["digest"];
	    }
	}
	export class DeploymentInput {
	    projectId: number;
	    token: string;
	    password: string;
	    passphrase: string;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.token = source["token"];
	        this.password = source["password"];
	        this.passphrase = source["passphrase"];
	    }
	}
	export class DeploymentPlanDTO {
	    token: string;
	    projectId: number;
	    provider: string;
	    source: string;
	    target: string;
	    files: DeploymentFileDTO[];
	    bytes: number;
	    excluded: number;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentPlanDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.projectId = source["projectId"];
	        this.provider = source["provider"];
	        this.source = source["source"];
	        this.target = source["target"];
	        this.files = this.convertValues(source["files"], DeploymentFileDTO);
	        this.bytes = source["bytes"];
	        this.excluded = source["excluded"];
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
	export class DeploymentTaskDTO {
	    id: string;
	    projectId: number;
	    provider: string;
	    state: string;
	    message: string;
	    target: string;
	    completed: number;
	    total: number;
	    created: string;
	    finished: string;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentTaskDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.provider = source["provider"];
	        this.state = source["state"];
	        this.message = source["message"];
	        this.target = source["target"];
	        this.completed = source["completed"];
	        this.total = source["total"];
	        this.created = source["created"];
	        this.finished = source["finished"];
	    }
	}
	export class ExtensionDTO {
	    name: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExtensionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	    }
	}
	export class InstallInput {
	    kind: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.version = source["version"];
	    }
	}
	export class ItemDTO {
	    kind: string;
	    name: string;
	    description: string;
	    category: string;
	    count: number;
	    recommended: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.count = source["count"];
	        this.recommended = source["recommended"];
	    }
	}
	export class PluginActionDTO {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new PluginActionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class ProjectCreateInput {
	    kind: string;
	    name: string;
	    host: string;
	    root: string;
	    run: string;
	    command: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectCreateInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.root = source["root"];
	        this.run = source["run"];
	        this.command = source["command"];
	    }
	}
	export class ProjectDTO {
	    id: number;
	    name: string;
	    kind: string;
	    hosts: string[];
	    root: string;
	    run: string;
	    path: string;
	    software?: number;
	    version: string;
	    command: string;
	    nginx: string;
	    environmentFile: string;
	    environmentFiles: string[];
	    port: number;
	    autostart: boolean;
	    state: string;
	    url: string;
	    need: string;
	    created: string;
	    updated: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.hosts = source["hosts"];
	        this.root = source["root"];
	        this.run = source["run"];
	        this.path = source["path"];
	        this.software = source["software"];
	        this.version = source["version"];
	        this.command = source["command"];
	        this.nginx = source["nginx"];
	        this.environmentFile = source["environmentFile"];
	        this.environmentFiles = source["environmentFiles"];
	        this.port = source["port"];
	        this.autostart = source["autostart"];
	        this.state = source["state"];
	        this.url = source["url"];
	        this.need = source["need"];
	        this.created = source["created"];
	        this.updated = source["updated"];
	    }
	}
	export class ProjectGitCommitInput {
	    projectId: number;
	    message: string;
	    bumpVersion: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectGitCommitInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.message = source["message"];
	        this.bumpVersion = source["bumpVersion"];
	    }
	}
	export class ProjectGitFileDTO {
	    path: string;
	    status: string;
	    staged: boolean;
	    unstaged: boolean;
	    untracked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectGitFileDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.status = source["status"];
	        this.staged = source["staged"];
	        this.unstaged = source["unstaged"];
	        this.untracked = source["untracked"];
	    }
	}
	export class ProjectGitRemoteDTO {
	    name: string;
	    url: string;
	    pushUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectGitRemoteDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.pushUrl = source["pushUrl"];
	    }
	}
	export class ProjectGitRemoteInput {
	    projectId: number;
	    name: string;
	    url: string;
	    pushUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectGitRemoteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.pushUrl = source["pushUrl"];
	    }
	}
	export class ProjectGitStatusDTO {
	    remotes: ProjectGitRemoteDTO[];
	    upstream: string;
	    ahead: number;
	    behind: number;
	    projectId: number;
	    projectName: string;
	    available: boolean;
	    repository: boolean;
	    nested: boolean;
	    branch: string;
	    lastCommit: string;
	    commitMessage: string;
	    changed: number;
	    staged: number;
	    unstaged: number;
	    untracked: number;
	    files: ProjectGitFileDTO[];
	    version: string;
	    nextVersion: string;
	    versionFile: string;
	    versionCanBump: boolean;
	    versionError: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectGitStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.remotes = this.convertValues(source["remotes"], ProjectGitRemoteDTO);
	        this.upstream = source["upstream"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.available = source["available"];
	        this.repository = source["repository"];
	        this.nested = source["nested"];
	        this.branch = source["branch"];
	        this.lastCommit = source["lastCommit"];
	        this.commitMessage = source["commitMessage"];
	        this.changed = source["changed"];
	        this.staged = source["staged"];
	        this.unstaged = source["unstaged"];
	        this.untracked = source["untracked"];
	        this.files = this.convertValues(source["files"], ProjectGitFileDTO);
	        this.version = source["version"];
	        this.nextVersion = source["nextVersion"];
	        this.versionFile = source["versionFile"];
	        this.versionCanBump = source["versionCanBump"];
	        this.versionError = source["versionError"];
	        this.error = source["error"];
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
	export class ProjectPluginActionInput {
	    projectId: number;
	    pluginId: string;
	    actionId: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectPluginActionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.pluginId = source["pluginId"];
	        this.actionId = source["actionId"];
	    }
	}
	export class ProjectPluginDTO {
	    id: string;
	    name: string;
	    description: string;
	    available: boolean;
	    reason: string;
	    version: string;
	    actions: PluginActionDTO[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectPluginDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.available = source["available"];
	        this.reason = source["reason"];
	        this.version = source["version"];
	        this.actions = this.convertValues(source["actions"], PluginActionDTO);
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
	export class ProjectTemplateDTO {
	    kind: string;
	    label: string;
	    icon: string;
	    description: string;
	    runtimeKind: string;
	    commandMode: string;
	    defaultCommand: string;
	    commandLabel: string;
	    commandPlaceholder: string;
	    capabilities: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectTemplateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	        this.runtimeKind = source["runtimeKind"];
	        this.commandMode = source["commandMode"];
	        this.defaultCommand = source["defaultCommand"];
	        this.commandLabel = source["commandLabel"];
	        this.commandPlaceholder = source["commandPlaceholder"];
	        this.capabilities = source["capabilities"];
	    }
	}
	export class ProjectUpdateInput {
	    id: number;
	    hosts: string[];
	    root: string;
	    run: string;
	    software?: number;
	    command: string;
	    nginx: string;
	    environmentFile: string;
	    autostart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectUpdateInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hosts = source["hosts"];
	        this.root = source["root"];
	        this.run = source["run"];
	        this.software = source["software"];
	        this.command = source["command"];
	        this.nginx = source["nginx"];
	        this.environmentFile = source["environmentFile"];
	        this.autostart = source["autostart"];
	    }
	}
	export class Result {
	    ok: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	    }
	}
	export class SettingsDTO {
	    theme: string;
	    root: string;
	    port: number;
	    hosts: boolean;
	    developerToolsCLI: string;
	    listening: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.root = source["root"];
	        this.port = source["port"];
	        this.hosts = source["hosts"];
	        this.developerToolsCLI = source["developerToolsCLI"];
	        this.listening = source["listening"];
	        this.error = source["error"];
	    }
	}
	export class SettingsInput {
	    theme: string;
	    root: string;
	    port: number;
	    hosts: boolean;
	    developerToolsCLI: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.root = source["root"];
	        this.port = source["port"];
	        this.hosts = source["hosts"];
	        this.developerToolsCLI = source["developerToolsCLI"];
	    }
	}
	export class SoftwareDTO {
	    id: number;
	    kind: string;
	    name: string;
	    version: string;
	    path: string;
	    state: string;
	    progress: number;
	    error: string;
	    main: boolean;
	    port: number;
	    autostart: boolean;
	    running: boolean;
	    pid: number;
	    daemon: boolean;
	    service: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SoftwareDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.state = source["state"];
	        this.progress = source["progress"];
	        this.error = source["error"];
	        this.main = source["main"];
	        this.port = source["port"];
	        this.autostart = source["autostart"];
	        this.running = source["running"];
	        this.pid = source["pid"];
	        this.daemon = source["daemon"];
	        this.service = source["service"];
	    }
	}
	export class VersionDTO {
	    version: string;
	    recommended: boolean;
	    installed: boolean;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.recommended = source["recommended"];
	        this.installed = source["installed"];
	        this.note = source["note"];
	    }
	}

}

