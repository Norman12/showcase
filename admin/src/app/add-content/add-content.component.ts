import {
  Component,
  OnInit,
  OnDestroy
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { Router } from '@angular/router';

import { ApiService } from '../shared/service/api.service';
import { Emitter } from '../shared/service/emitter.service';

import { ContentCreateRequest } from '../shared/service/request/content.create';

import { WaitingEvent, SuccessEvent, FailureEvent } from '../shared/view/result-button';
import { ClearEvent } from '../shared/view/media-input';

@Component({
  selector: 'add-content',
  styleUrls: [],
  templateUrl: './add-content.component.html'
})
export class AddContentComponent implements OnInit, OnDestroy {

  public model: ContentCreateRequest = {
    title: '',
    subtitle: '',
    paragraphs: [],
    tags: [],
    technologies: [],
    references: {}
  };

  public buttonChannel: string = "button-add-content";

  private mediaChannel: string = 'media-input';

  private subs: Subscription[] = [];

  constructor(
    private router: Router,
    private api: ApiService
  ) { }

  public ngOnInit() {

  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public save() {
    Emitter.get(this.buttonChannel).emit(new WaitingEvent());

    setTimeout(() => {
      this.subs.push(this.api.createContent(this.model)
        .subscribe(result => {
          if (result) {
            Emitter.get(this.buttonChannel).emit(new SuccessEvent());
            Emitter.get(this.mediaChannel).emit(new ClearEvent());

            let $this = this;
            setTimeout(function() {
              $this.router.navigate(['/pages']);
            }, 1000);
          } else {
            Emitter.get(this.buttonChannel).emit(new FailureEvent());
          }
        })
      );
    }, 100);

  }
}
