import {
  Component,
  OnInit,
  OnDestroy,
  ViewChildren
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { Router } from '@angular/router';

import { ApiService } from '../shared/service/api.service';
import { Emitter } from '../shared/service/emitter.service';

import { ProjectCreateRequest } from '../shared/service/request/project.create';

import { MediaInputComponent } from '../shared/view/media-input';
import { WaitingEvent, SuccessEvent, FailureEvent } from '../shared/view/result-button';
import { ClearEvent } from '../shared/view/media-input';

@Component({
  selector: 'add-project',
  styleUrls: [],
  templateUrl: './add-project.component.html'
})
export class AddProjectComponent implements OnInit, OnDestroy {

  public model: ProjectCreateRequest = {
    title: '',
    subtitle: '',
    about: '',
    image: [{
      name: '',
      caption: '',
      resource: 'image',
      uploaded: false,
      removed: false,
      file: {
        name: '',
        type: '',
        data: null
      }
    }],
    logo: [{
      name: '',
      caption: '',
      resource: 'logo',
      uploaded: false,
      removed: false,
      file: {
        name: '',
        type: '',
        data: null
      }
    }],
    media: [],
    tags: [],
    technologies: [],
    references: {},
    client: {
      name: '',
      about: '',
      image: [{
        name: '',
        caption: '',
        resource: 'client',
        uploaded: false,
        removed: false,
        file: {
          name: '',
          type: '',
          data: null
        }
      }]
    }
  };

  public buttonChannel: string = "button-add-project";

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
      this.subs.push(this.api.createProject(this.model)
        .subscribe(result => {
          if (result) {
            Emitter.get(this.buttonChannel).emit(new SuccessEvent());
            Emitter.get(this.mediaChannel).emit(new ClearEvent());

            let $this = this;
            setTimeout(() => {
              $this.router.navigate(['/']);
            }, 1000);
          } else {
            Emitter.get(this.buttonChannel).emit(new FailureEvent());
          }
        })
      );
    }, 100);

  }

}
