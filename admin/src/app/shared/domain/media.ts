export class Media {
  constructor(
    public resource: string,

    public name: string,
    public caption: string,
    public uploaded: boolean,
    public removed: boolean,
    public file: {
      name: string,
      type: string,
      data: string
    }
  ) {
  }
}
