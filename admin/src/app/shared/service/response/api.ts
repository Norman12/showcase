export class ApiResponse<T> {
	constructor(public error: string,
		public content: T) {
	}
}