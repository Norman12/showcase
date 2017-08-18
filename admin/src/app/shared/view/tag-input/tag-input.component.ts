import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { ModalDialogComponent } from '../modal-dialog';

@Component({
  selector: 'tag-input',
  template: `
  	<div class="row">
		<div class="col-sm-12">
			<h2>{{title}}</h2>
		</div>
	</div>
	<div *ngIf="isEmpty()" class="empty">
		None added
	</div>
	<div class="md-chips">
		<div *ngFor="let item of data" class="md-chip">
			<span>{{item}}</span>
			<button type="button" class="md-chip-remove" (click)="remove(item)">
	    	</button>
		</div>
	</div>
	<div class="row">
		<div class="col-xs-8 col-sm-10">
			<input class="form-control" type="text" name="tag" [(ngModel)]="input" placeholder="Type here">
		</div>
		<div class="col-xs-4 col-sm-2">
			<button type="button" class="btn btn-default btn-block" (click)="add()">Add</button>
		</div>
	</div>
  `,
  styleUrls: ['./tag-input.component.scss']
})
export class TagInputComponent implements OnInit {

  @Input()
  public title: string;

  @Input()
  public data: string[];

  public input: string;

  constructor(
  ) {
  }

  public ngOnInit() {
    if (this.data == null) {
      this.data = [];
    }
  }

  public add() {
    if (this.input === '') {
      return;
    }

    this.data.push(this.input);
    this.input = '';
  }

  public remove(item: string) {
    let index = this.data.indexOf(item);
    if (index !== -1) {
      this.data.splice(index, 1);
    }
  }

  public isEmpty() {
    return this.data.length === 0;
  }

}