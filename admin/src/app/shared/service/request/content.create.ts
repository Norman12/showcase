import { Paragraph } from '../../domain/paragraph';

export interface ContentCreateRequest {
	title: string;
	subtitle: string;
	paragraphs: Paragraph[];
	tags: string[];
	technologies: string[];
	references: { [index: string]: string };
}
