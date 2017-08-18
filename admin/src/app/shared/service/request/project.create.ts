import { Media } from '../../domain/media';

export interface ProjectCreateRequest {
	title: string;
	subtitle: string;
	about: string;
	image: [Media];
	logo: [Media];
	media: Media[];
	tags: string[];
	technologies: string[];
	references: {[index: string]: string};
	client: {
		name: string;
		about: string;
		image: [Media];
	};
}