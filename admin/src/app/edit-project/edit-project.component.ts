import {
  Component,
  OnInit,
  OnDestroy
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { ActivatedRoute } from '@angular/router';

import { AppState } from '../app.service';

import { ApiService } from '../shared/service/api.service';
import { Emitter } from '../shared/service/emitter.service';

import { ProjectEditRequest } from '../shared/service/request/project.edit';

import { WaitingEvent, SuccessEvent, FailureEvent } from '../shared/view/result-button';
import { ClearEvent } from '../shared/view/media-input';

@Component({
  selector: 'edit-project',
  styleUrls: [],
  templateUrl: './edit-project.component.html'
})
export class EditProjectComponent implements OnInit, OnDestroy {

  public model: ProjectEditRequest = {
    slug: '',
    title: '',
    subtitle: '',
    about: '',
    image: [{
      name: '',
      caption: '',
      resource: '',
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
      resource: '',
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

  public config: any = {};

  public buttonChannel: string = "button-edit-project";

  private mediaChannel: string = 'media-input';

  private subs: Subscription[] = [];

  constructor(
    private route: ActivatedRoute,
    private api: ApiService,
    private state: AppState
  ) {
  }

  public ngOnInit() {
    this.model = this.route.snapshot.data['project'];
    this.config = this.state.get("toolbar");
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public save() {
    Emitter.get(this.buttonChannel).emit(new WaitingEvent());

    this.subs.push(this.api.editProject(this.model)
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
      this.api.getProject(this.model.slug)
        .subscribe(model => this.model = model)
    )
  }
}
