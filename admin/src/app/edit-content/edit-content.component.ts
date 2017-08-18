import {
  Component,
  OnInit,
  OnDestroy
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { ActivatedRoute } from '@angular/router';

import { ApiService } from '../shared/service/api.service';
import { Emitter } from '../shared/service/emitter.service';

import { ContentEditRequest } from '../shared/service/request/content.edit';

import { WaitingEvent, SuccessEvent, FailureEvent } from '../shared/view/result-button';
import { ClearEvent } from '../shared/view/media-input';

@Component({
  selector: 'edit-content',
  styleUrls: [],
  templateUrl: './edit-content.component.html'
})
export class EditContentComponent implements OnInit, OnDestroy {

  public model: ContentEditRequest = {
    slug: '',
    title: '',
    subtitle: '',
    paragraphs: [],
    tags: [],
    technologies: [],
    references: {}
  };

  public buttonChannel: string = "button-edit-content";

  private mediaChannel: string = 'media-input';

  private subs: Subscription[] = [];

  constructor(
    private route: ActivatedRoute,
    private api: ApiService
  ) { }

  public ngOnInit() {
    this.model = this.route.snapshot.data['content'];
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public save() {
    Emitter.get(this.buttonChannel).emit(new WaitingEvent());

    this.subs.push(this.api.editContent(this.model)
      .subscribe(result => {
        if (result) {
          Emitter.get(this.buttonChannel).emit(new SuccessEvent());
          Emitter.get(this.mediaChannel).emit(new ClearEvent());

          this.reload();
        } else {
          Emitter.get(this.buttonChannel).emit(new FailureEvent());
        }
      })
    );
  }

  private reload() {
    this.subs.push(
      this.api.getContent(this.model.slug)
        .subscribe(model => this.model = model)
    )
  }
}
