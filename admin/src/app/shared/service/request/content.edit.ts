import { Paragraph } from '../../domain/paragraph';

export interface ContentEditRequest {
	slug: string;
	title: string;
	subtitle: string;
	paragraphs: Paragraph[];
	tags: string[];
	technologies: string[];
	references: { [index: string]: string };
}
