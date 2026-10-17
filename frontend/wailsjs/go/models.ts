export namespace main {
	
	export class ImageTags {
	    file: string;
	    path: string;
	    tags: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageTags(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.path = source["path"];
	        this.tags = source["tags"];
	        this.error = source["error"];
	    }
	}
	export class MoveResult {
	    moved: number;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new MoveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.moved = source["moved"];
	        this.durationMs = source["durationMs"];
	    }
	}

}

export namespace model {
	
	export class BackupRecord {
	    ID: number;
	    SourcePath: string;
	    DestinationPath: string;
	    FileName: string;
	    Status: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new BackupRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SourcePath = source["SourcePath"];
	        this.DestinationPath = source["DestinationPath"];
	        this.FileName = source["FileName"];
	        this.Status = source["Status"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
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
	export class TagCount {
	    name: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new TagCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}

}

