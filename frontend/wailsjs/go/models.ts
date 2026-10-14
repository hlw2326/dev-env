export namespace wails {
	
	export class ConfigItemDTO {
	    key: string;
	    label: string;
	    value: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigItemDTO(source);
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
	    items: ConfigItemDTO[];
	    file: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], ConfigItemDTO);
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
	
	export class CreateSiteInput {
	    kind: string;
	    name: string;
	    host: string;
	    root: string;
	    run: string;
	    command: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateSiteInput(source);
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
	export class SaveConfigInput {
	    id: number;
	    items: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new SaveConfigInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.items = source["items"];
	    }
	}
	export class SaveConfigResult {
	    ok: boolean;
	    message: string;
	    restart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SaveConfigResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.restart = source["restart"];
	    }
	}
	export class SaveSettingsInput {
	    theme: string;
	    root: string;
	    port: number;
	    hosts: boolean;
	    autostart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SaveSettingsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.root = source["root"];
	        this.port = source["port"];
	        this.hosts = source["hosts"];
	        this.autostart = source["autostart"];
	    }
	}
	export class SettingsDTO {
	    theme: string;
	    root: string;
	    port: number;
	    hosts: boolean;
	    autostart: boolean;
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
	        this.autostart = source["autostart"];
	        this.listening = source["listening"];
	        this.error = source["error"];
	    }
	}
	export class SiteDTO {
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
	    port: number;
	    autostart: boolean;
	    state: string;
	    url: string;
	    need: string;
	    created: string;
	    updated: string;
	
	    static createFrom(source: any = {}) {
	        return new SiteDTO(source);
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
	        this.port = source["port"];
	        this.autostart = source["autostart"];
	        this.state = source["state"];
	        this.url = source["url"];
	        this.need = source["need"];
	        this.created = source["created"];
	        this.updated = source["updated"];
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
	    }
	}
	export class UpdateSiteInput {
	    id: number;
	    hosts: string[];
	    root: string;
	    run: string;
	    software?: number;
	    command: string;
	    autostart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSiteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.hosts = source["hosts"];
	        this.root = source["root"];
	        this.run = source["run"];
	        this.software = source["software"];
	        this.command = source["command"];
	        this.autostart = source["autostart"];
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

