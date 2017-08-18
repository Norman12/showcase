import {
  Component,
  OnInit,
  OnDestroy
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { ActivatedRoute } from '@angular/router';

import { ApiService } from '../shared/service/api.service';
import { Emitter } from '../shared/service/emitter.service';

import { UserEditRequest } from '../shared/service/request/user.edit';

import { WaitingEvent, SuccessEvent, FailureEvent } from '../shared/view/result-button';
import { ClearEvent } from '../shared/view/media-input';

@Component({
  selector: 'about',
  styleUrls: [],
  templateUrl: './about.component.html'
})
export class AboutComponent implements OnInit, OnDestroy {

  public model: UserEditRequest = {
    name: '',
    title: '',
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
    references: {},
    networks: {},
    experiences: {},
    interests: [],
    contact: {
      country: '',
      city: '',
      street: '',
      email: '',
      phone: '',
    }
  };

  public buttonChannel: string = 'button-about';

  private mediaChannel: string = 'media-input';

  private subs: Subscription[] = [];

  constructor(
    private route: ActivatedRoute,
    private api: ApiService
  ) { }

  public ngOnInit() {
    this.model = this.route.snapshot.data['about'];
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public save() {
    Emitter.get(this.buttonChannel).emit(new WaitingEvent());

    this.subs.push(this.api.editUser(this.model)
      .subscribe(result => {
        if (result) {
          Emitter.get(this.buttonChannel).emit(new SuccessEvent());
          Emitter.get(this.mediaChannel).emit(new ClearEvent());
        } else {
          Emitter.get(this.buttonChannel).emit(new FailureEvent());
        }
      })
    );
  }
}
