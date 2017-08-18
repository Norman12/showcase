import { Component, OnInit, OnDestroy, Input } from '@angular/core';
import { Media } from '../../domain/media';

import { Subscription } from 'rxjs/Rx';

import { Generator } from '../../../app.service';
import { Emitter } from '../../service/emitter.service';

@Component({
  selector: 'media-input',
  template: `
	<div class="row">
		<div class="col-sm-12">
			<h2>{{title}}</h2>
		</div>
	</div>
  <div *ngIf="isEmpty()" class="empty">
    None added
  </div>
	<div class="media" *ngFor="let item of data; let i = index">
    <form>
			<div class="row upload">
				<div class="col-sm-12 col-md-8">
					<input type="file" (change)="uploadFile($event, item)" [disabled]="multiple && item.uploaded">
				</div>
				<div *ngIf="multiple" class="col-sm-12 col-md-4">
					<button type="button" class="btn btn-danger btn-sm pull-right" (click)="remove(item)">Remove</button>
				</div>
        <div *ngIf="!multiple" class="col-sm-12 col-md-4">
          <button type="button" [ngClass]="{'btn-default': !item.removed, 'btn-warning': item.removed}" class="btn btn-sm pull-right" (click)="markRemoved(item)">
            <span *ngIf="!item.removed; else removed">
              Clear
            </span>
            <ng-template #removed>Cleared</ng-template>
          </button>
        </div>
			</div>
			<div class="row">
				<div class="col-sm-12 col-md-6">
				  <div class="form-group">
				    <label for="{{'name-' + i}}">Media name</label>
				    <input type="text" class="form-control" name="{{'name-' + i}}" [(ngModel)] = "item.name">
				  </div>
				</div>
				<div class="col-sm-12 col-md-6">
				  <div class="form-group">
				    <label for="{{'caption-' + i}}">Media caption</label>
				    <input type="text" class="form-control" name="{{'caption-' + i}}" [(ngModel)] = "item.caption">
				  </div>
				</div>
			</div>
      <div *ngIf="multiple" class="row">
        <div class="col-sm-12">
          <hr />
        </div>
      </div>
    </form>
	</div>
	<div class="row" *ngIf="multiple">
		<div class="col-sm-12">
			<button type="button" class="btn btn-default pull-right" (click)="addEmpty()">
        <span *ngIf="isEmpty(); else not">
            Add new media
        </span>
        <ng-template #not>Add another media</ng-template>
      </button>
		</div>
	</div>
  `,
  styleUrls: ['./media-input.component.scss'],
})
export class MediaInputComponent implements OnInit, OnDestroy {

  @Input()
  public data: Media[];

  @Input()
  public multiple: boolean;

  @Input()
  public title: string;

  @Input()
  public prefix: string = 'media-';

  private channel: string = 'media-input';

  private sub: Subscription;

  constructor(
    private generator: Generator
  ) {
    if (FileReader.prototype.readAsBinaryString === undefined) {
      FileReader.prototype.readAsBinaryString = function(fileData) {
        let binary = "";
        let pt = this;
        let reader = new FileReader();
        reader.onload = function(e) {
          let bytes = new Uint8Array(reader.result);
          let length = bytes.byteLength;
          for (let i = 0; i < length; i++) {
            binary += String.fromCharCode(bytes[i]);
          }

          pt.content = binary;
          pt.onload();
        }
        reader.readAsArrayBuffer(fileData);
      }
    }
  }

  public ngOnInit() {
    if (this.data == null)
      this.data = [];

    this.sub = Emitter.get(this.channel)
      .filter(event => event instanceof ClearEvent)
      .subscribe(event => {
        this.clearUpload()
      });
  }

  public ngOnDestroy() {
    if (this.sub != null)
      this.sub.unsubscribe();
  }

  public addEmpty() {
    this.data.push(new Media(this.generator.generateRandomString(this.prefix, 16), '', '', false, false, { name: '', type: '', data: null }));
  }

  public remove(media: Media) {
    let index = this.data.indexOf(media);
    if (index > -1) {
      this.data.splice(index, 1);
    }
  }

  public uploadFile(event, media: Media) {
    let fileList: FileList = event.target.files;
    if (fileList.length > 0) {
      let file: File = fileList[0];

      let reader = new FileReader();
      reader.onload = function(event: any) {
        media.file = {
          name: file.name,
          type: file.type,
          data: btoa(event.target.result)
        };

        media.name = file.name;
        media.uploaded = true;
        media.removed = false;
      }

      reader.readAsBinaryString(file);
    }
  }

  public markRemoved(media: Media) {
    media.removed = true;
    media.uploaded = false;
    media.name = '';
    media.file = { name: '', type: '', data: '' };
  }

  public isEmpty() {
    return this.data.length === 0;
  }

  private clearUpload() {
    for (let media of this.data) {
      media.file = { name: '', type: '', data: '' };
    }
  }
}

export class ClearEvent {
}