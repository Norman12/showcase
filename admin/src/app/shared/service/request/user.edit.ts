import { Media } from '../../domain/media';

export interface UserEditRequest {
	name: string;
	title: string;
	about: string;
	image: [Media];
	logo: [Media];
	references: {[index: string]: string};
	networks: {[index: string]: string};
	experiences: {[index: string]: string};
	interests: string[];
	contact: {
		country: string;
		city: string;
		street: string;
		email: string;
		phone: string;
	};
}