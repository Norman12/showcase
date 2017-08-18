import { Component, OnInit, Input } from '@angular/core';
import { Theme } from '../../domain/theme';

@Component({
  selector: 'theme-input',
  template: `
    <div class="row">
		<div class="col-sm-12">
			<h2>Themes</h2>
		</div>
	</div>
  	<div class="row themes">
  		<div *ngFor="let item of data" (click)="select(item)" [class.selected]="isSelected(item)" class="col-sm-12 col-md-4">
  			<div class="theme">
  			  	<div class="theme__image" [ngStyle]="{'background-image': 'url(' + item.image + ')'}">
	  			</div>
	  			<div class="theme__content">
		  			<div class="content_title">
		   				{{item.name}}
		  			</div>
		  			<div class="content_author">
		  				{{item.author}}
		  			</div>
	  			</div>
  			</div>
  		</div>
  	</div>
  `,
  styleUrls: ['./theme-input.component.scss'],
})
export class ThemeInputComponent implements OnInit {

  @Input()
  public data: Theme[];

  @Input()
  public selected: string;

  constructor(
  ) {
  }

  public ngOnInit() {

  }

  public select(theme: Theme) {
    this.selected = theme.path;
  }

  public isSelected(theme: Theme): boolean {
    return this.selected === theme.path
  }
}