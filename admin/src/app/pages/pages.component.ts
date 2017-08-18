import {
  Component,
  OnInit,
  OnDestroy,
  ViewChild
} from '@angular/core';

import { Observable, Subscription } from 'rxjs/Rx';

import { Router, ActivatedRoute } from '@angular/router';

import { Content } from '../shared/domain/content';

import { ModalDialogComponent } from '../shared/view/modal-dialog';

import { ApiService } from '../shared/service/api.service';

@Component({
  selector: 'pages',
  styleUrls: ['./pages.component.scss'],
  templateUrl: './pages.component.html'
})
export class PagesComponent implements OnInit, OnDestroy {

  public contents: Content[];
  public selected: Content = new Content("", "", "");

  @ViewChild('modalDelete')
  modalDelete: ModalDialogComponent;

  private subscriptions: Subscription[] = [];

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private api: ApiService,
  ) { }

  public ngOnInit() {
    this.contents = this.route.snapshot.data['contents'];
  }

  public openAddContent() {
    this.router.navigate(['/add-content']);
  }

  public openEdit(content: Content) {
    this.router.navigate(['/edit-content', content.slug]);
  }

  public openDelete(content: Content) {
    this.selected = content;
    this.modalDelete.show();
  }

  public delete() {
    this.subscriptions.push(
      this.api.deleteContent(this.selected.slug).subscribe(result => {
        if (result) {
          this.subscriptions.push(
            this.api.getContents().subscribe(
              contents => this.contents = contents
            )
          )
          this.modalDelete.hide();
        } else {
          alert("Something went wrong");
        }
      })
    );
  }

  public ngOnDestroy() {
    for (let s of this.subscriptions) {
      s.unsubscribe();
    }
  }
}
