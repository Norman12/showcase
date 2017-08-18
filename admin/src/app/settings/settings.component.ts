import {
  Component,
  OnInit,
  OnDestroy,
  ViewChild
} from '@angular/core';

import { Observable, Subscription } from 'rxjs/Rx';

import { ActivatedRoute } from '@angular/router';

import { Theme } from '../shared/domain/theme';

import { ApiService } from '../shared/service/api.service';
import { Emitter } from '../shared/service/emitter.service';

import { MetaEditRequest } from '../shared/service/request/meta.edit';
import { CredentialsEditRequest } from '../shared/service/request/credentials.edit';

import { ModalDialogComponent } from '../shared/view/modal-dialog';
import { WaitingEvent, SuccessEvent, FailureEvent } from '../shared/view/result-button';

@Component({
  selector: 'settings',
  styleUrls: ['./settings.component.scss'],
  templateUrl: './settings.component.html'
})
export class SettingsComponent implements OnInit, OnDestroy {

  public active: number = 0;

  public models: {
    meta: MetaEditRequest;
    credentials: CredentialsEditRequest;
    theme: { selected: string, themes: Theme[] };
  } = {
    meta: {
      title: '',
      site: '',
      tags: {},
      og_tags: {}
    },
    credentials: {
      email: '',
      password: '',
      password_repeat: '',
    },
    theme: { selected: '', themes: [] }
  };

  public buttonChannelMeta: string = "button-settings-meta";
  public buttonChannelCredentials: string = "button-settings-credentials";
  public buttonChannelTheme: string = "button-settings-theme";

  @ViewChild('modalPasswordMatch')
  modalPasswordMatch: ModalDialogComponent;

  private subs: Subscription[] = [];

  private siteChannel: string = 'site';

  constructor(
    private api: ApiService,
    private route: ActivatedRoute
  ) { }

  public ngOnInit() {
    let settings = this.route.snapshot.data['settings'];
    this.models = {
      meta: settings[0],
      credentials: settings[1],
      theme: settings[2]
    }
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public saveMeta() {
    Emitter.get(this.buttonChannelMeta).emit(new WaitingEvent());

    this.subs.push(
      this.api.editMeta(this.models.meta)
        .subscribe(result => {
          if (result) {
            Emitter.get(this.buttonChannelMeta).emit(new SuccessEvent());
            Emitter.get(this.siteChannel).emit(0);
          } else {
            Emitter.get(this.buttonChannelMeta).emit(new FailureEvent());
          }
        })
    );
  }

  public saveCredentials() {
    if ((this.models.credentials.password === '' || this.models.credentials.password_repeat === '') || this.models.credentials.password !== this.models.credentials.password_repeat) {
      this.modalPasswordMatch.show();
      return;
    }

    Emitter.get(this.buttonChannelCredentials).emit(new WaitingEvent());

    this.subs.push(
      this.api.editCredentials(this.models.credentials)
        .subscribe(result => {
          if (result) {
            Emitter.get(this.buttonChannelCredentials).emit(new SuccessEvent());
          } else {
            Emitter.get(this.buttonChannelCredentials).emit(new FailureEvent());
          }
        })
    );
  }

  public saveTheme() {
    Emitter.get(this.buttonChannelTheme).emit(new WaitingEvent());

    this.subs.push(
      this.api.editTheme({ path: this.models.theme.selected })
        .subscribe(result => {
          if (result) {
            Emitter.get(this.buttonChannelTheme).emit(new SuccessEvent());
          } else {
            Emitter.get(this.buttonChannelTheme).emit(new FailureEvent());
          }
        })
    );
  }

  public isActive(i: number) {
    return this.active === i;
  }

  public setActive(i: number) {
    return this.active = i;
  }
}
