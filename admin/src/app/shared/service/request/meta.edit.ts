export interface MetaEditRequest {
	title: string;
	site: string;
	tags: {[index: string]: string};
	og_tags: {[index: string]: string};
}