import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { Paragraph } from '../../domain/paragraph';

import { Generator } from '../../../app.service';

@Component({
  selector: 'paragraph-input',
  template: `
  <div class="row">
		<div class="col-sm-12">
			<h2>Paragraphs</h2>
		</div>
	</div>
	<div *ngIf="isEmpty()" class="empty">
		None added
	</div>

  <div *ngFor="let item of data">
    <form>
      <div class="row">
        <div class="col-sm-12">
          <div class="form-group" style="margin-bottom: 0;">
            <label>Title</label>
          </div>
        </div>
      </div>
      <div class="row" style="margin-bottom: 15px;">
        <div class="col-xs-8 col-sm-10">
          <input class="form-control" type="text" name="title" [(ngModel)]="item.title">
        </div>
        <div class="col-xs-4 col-sm-2">
          <button type="button" class="btn btn-danger btn-block" (click)="remove(item)">Remove</button>
        </div>
      </div>
      <div class="row">
        <div class="col-sm-12">
          <div class="form-group">
            <label for="content">Content</label>
            <ckeditor [(ngModel)]="item.content" name="content"></ckeditor>
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-sm-12">
          <media-input title="Media" [(data)]="item.media"></media-input>
        </div>
      </div>
      <div class="row">
        <div class="col-sm-12">
          <hr />
        </div>
      </div>
    </form>
  </div>

  <div class="row">
    <div class="col-sm-12">
      <button type="button" class="btn btn-default pull-right" (click)="add()">
        <span *ngIf="isEmpty(); else not">
            Add new paragraph
        </span>
        <ng-template #not>Add another paragraph</ng-template>
      </button>
    </div>
  </div>
  `,
  styleUrls: ['./paragraph-input.component.scss']
})
export class ParagraphInputComponent implements OnInit {

  @Input()
  public data: Paragraph[];

  public counter: number = 0;

  private prefixParagraph: string = 'paragraph-';
  private prefixMedia: string = 'media-';

  constructor(
    private generator: Generator
  ) {
  }

  public ngOnInit() {
    if (this.data == null) {
      this.data = [];
    }
  }

  public add() {
    this.data.push(new Paragraph(this.generator.generateRandomString(this.prefixParagraph, 16), '', '', [{
      name: '',
      caption: '',
      resource: this.generator.generateRandomString(this.prefixMedia, 16),
      uploaded: false,
      removed: false,
      file: {
        name: '',
        type: '',
        data: null
      }
    }]));

    console.log(this.data);
  }

  public remove(item: Paragraph) {
    let index = this.data.indexOf(item);
    if (index !== -1) {
      this.data.splice(index, 1);
    }
  }

  public isEmpty() {
    return this.data.length === 0;
  }

}