export namespace main {
	
	export class Proposal {
	    id: string;
	    title: string;
	    subtitle: string;
	    prompt: string;
	    // Go type: time
	    created_at: any;
	    md_url?: string;
	    html_url?: string;
	    pdf_url?: string;
	    size_html: number;
	    size_pdf: number;
	
	    static createFrom(source: any = {}) {
	        return new Proposal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.subtitle = source["subtitle"];
	        this.prompt = source["prompt"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.md_url = source["md_url"];
	        this.html_url = source["html_url"];
	        this.pdf_url = source["pdf_url"];
	        this.size_html = source["size_html"];
	        this.size_pdf = source["size_pdf"];
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
	export class TextGenerationRequest {
	    title: string;
	    subtitle: string;
	    prompt: string;
	    model?: string;
	
	    static createFrom(source: any = {}) {
	        return new TextGenerationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.subtitle = source["subtitle"];
	        this.prompt = source["prompt"];
	        this.model = source["model"];
	    }
	}
	export class TextGenerationResponse {
	    id: string;
	    // Go type: time
	    created_at: any;
	    md_url?: string;
	
	    static createFrom(source: any = {}) {
	        return new TextGenerationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.md_url = source["md_url"];
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

