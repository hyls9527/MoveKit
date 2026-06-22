export namespace main {
	
	export class CloudProvider {
	    name: string;
	    folder: string;
	
	    static createFrom(source: any = {}) {
	        return new CloudProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.folder = source["folder"];
	    }
	}
	export class ConflictInfo {
	    relativePath: string;
	    srcSize: number;
	    dstSize: number;
	    srcModTime: string;
	    dstModTime: string;
	
	    static createFrom(source: any = {}) {
	        return new ConflictInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relativePath = source["relativePath"];
	        this.srcSize = source["srcSize"];
	        this.dstSize = source["dstSize"];
	        this.srcModTime = source["srcModTime"];
	        this.dstModTime = source["dstModTime"];
	    }
	}
	export class DiskInfo {
	    drive: string;
	    label: string;
	    totalBytes: number;
	    freeBytes: number;
	    isFixedDrive: boolean;
	    isRemovable: boolean;
	    isNetwork: boolean;
	    fileSystem: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.drive = source["drive"];
	        this.label = source["label"];
	        this.totalBytes = source["totalBytes"];
	        this.freeBytes = source["freeBytes"];
	        this.isFixedDrive = source["isFixedDrive"];
	        this.isRemovable = source["isRemovable"];
	        this.isNetwork = source["isNetwork"];
	        this.fileSystem = source["fileSystem"];
	    }
	}
	export class DurationEstimate {
	    isCrossVolume: boolean;
	    targetIsFixedDrive: boolean;
	    estimatedSec: number;
	    estimatedText: string;
	
	    static createFrom(source: any = {}) {
	        return new DurationEstimate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isCrossVolume = source["isCrossVolume"];
	        this.targetIsFixedDrive = source["targetIsFixedDrive"];
	        this.estimatedSec = source["estimatedSec"];
	        this.estimatedText = source["estimatedText"];
	    }
	}
	export class FolderInfo {
	    name: string;
	    guid: string;
	    sourcePath: string;
	    targetPath: string;
	    sizeBytes: number;
	    fileCount: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new FolderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.guid = source["guid"];
	        this.sourcePath = source["sourcePath"];
	        this.targetPath = source["targetPath"];
	        this.sizeBytes = source["sizeBytes"];
	        this.fileCount = source["fileCount"];
	        this.status = source["status"];
	    }
	}
	export class IntegrityIssue {
	    type: string;
	    folder: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new IntegrityIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.folder = source["folder"];
	        this.message = source["message"];
	    }
	}
	export class OneDriveStatus {
	    isActive: boolean;
	    folders: string[];
	    hasGPOBlock: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OneDriveStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isActive = source["isActive"];
	        this.folders = source["folders"];
	        this.hasGPOBlock = source["hasGPOBlock"];
	    }
	}
	export class UpdateInfo {
	    hasUpdate: boolean;
	    newVersion: string;
	    downloadURL: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.newVersion = source["newVersion"];
	        this.downloadURL = source["downloadURL"];
	    }
	}

}

