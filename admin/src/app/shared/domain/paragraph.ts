import { Media } from './media';

export class Paragraph {
	constructor(
		public resource: string,

		public title: string,
		public content: string,
		public media: [Media],
	) {
	}
}
