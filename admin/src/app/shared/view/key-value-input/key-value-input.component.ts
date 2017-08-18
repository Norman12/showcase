import { Component, OnInit, Input, ViewChild } from '@angular/core';
import { ModalDialogComponent } from '../modal-dialog';

@Component({
	selector: 'key-value-input',
	template: `
  	<div class="row">
		<div class="col-sm-12">
			<h2>{{title}}</h2>
		</div>
	</div>
	<div *ngIf="isEmpty()" class="empty">
		None added
	</div>
	<div *ngFor="let item of data | keys" class="row tag">
		<div class="col-sm-4 key-row">
			<input class="form-control key" type="text" [value]="item.key" disabled>
		</div>
		<div class="col-sm-8">
			<div class="flex-row">
				<input  class="form-control" type="text" [value]="item.val" disabled>
				<div class="btn-group actions" role="group" aria-label="...">
				  <button type="button" class="btn btn-default" (click)="openEdit(item)" >Edit</button>
				  <button type="button" class="btn btn-danger" (click)="remove(item)" ><span style="font-weight: bold;">×</span></button>
				</div>
			</div>
		</div>
	</div>
	<div class="row">
		<div class="col-sm-12">
			<button type="button" class="btn btn-default pull-right" (click)="openAddNew()">Add</button>
		</div>
	</div>

	<modal-dialog #modalAddNew>
	  <div class="app-modal-header">
	    Add New
	  </div>
	  <div class="app-modal-body">
	    <form>
		  <div class="form-group">
		    <label for="key">Key</label>
		    <input type="text" class="form-control" id="key" name="key" [(ngModel)] = "newTag.key">
		  </div>
		  <div class="form-group">
		    <label for="value">Value</label>
		    <input type="text" class="form-control" id="value" name="value" [(ngModel)] = "newTag.val">
		  </div>
		</form>
	  </div>
	  <div class="app-modal-footer">
	    <button type="button" class="btn btn-default" (click)="modalAddNew.hide()">Close</button>
	    <button type="button" class="btn btn-primary" (click)="addNew()">Save</button>
	  </div>
	</modal-dialog>

	<modal-dialog #modalEdit>
	  <div class="app-modal-header">
	    Edit
	  </div>
	  <div class="app-modal-body">
	    <form>
		  <div class="form-group">
		    <label for="key">Key</label>
		    <input type="text" class="form-control" id="key" name="key" [(ngModel)] = "editTag.key" disabled>
		  </div>
		  <div class="form-group">
		    <label for="value">Value</label>
		    <input type="text" class="form-control" id="value" name="value" [(ngModel)] = "editTag.val">
		  </div>
		</form>
	  </div>
	  <div class="app-modal-footer">
	    <button type="button" class="btn btn-default" (click)="modalEdit.hide()">Close</button>
	    <button type="button" class="btn btn-primary" (click)="edit()">Save</button>
	  </div>
	</modal-dialog>
  `,
	styleUrls: ['./key-value-input.component.scss']
})
export class KeyValueInputComponent implements OnInit {

	public newTag: { val: string, key: string } = { val: '', key: '' };
	public editTag: { val: string, key: string } = { val: '', key: '' };

	@Input()
	public title: string;

	@Input()
	public data: { [index: string]: string };

	@ViewChild('modalAddNew')
	modalAddNew: ModalDialogComponent;

	@ViewChild('modalEdit')
	modalEdit: ModalDialogComponent;

	constructor(
	) {
	}

	public ngOnInit() {
		if (this.data == null) {
			this.data = {};
		}
	}

	public openAddNew() {
		this.modalAddNew.show();
	}

	public openEdit(item) {
		this.editTag = { val: item.val, key: item.key };

		this.modalEdit.show();
	}

	public addNew() {
		if (this.newTag.key === '' || this.newTag.val === '') {
			return;
		}

		if (this.data.hasOwnProperty(this.newTag.key)) {
			return;
		}

		this.data[this.newTag.key] = this.newTag.val;

		this.newTag = { val: '', key: '' };

		this.modalAddNew.hide();
	}

	public edit() {

		this.data[this.editTag.key] = this.editTag.val;

		this.editTag = { val: '', key: '' };

		this.modalEdit.hide();
	}

	public remove(item) {
		delete this.data[item.key]
	}

	public isEmpty() {
		for (var prop in this.data) {
			if (this.data.hasOwnProperty(prop))
				return false;
		}
		return true;
	}

}