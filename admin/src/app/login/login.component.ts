import {
  Component,
  OnInit,
  OnDestroy
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { Router, ActivatedRoute, Params } from '@angular/router';

import { AppState, Permanent } from '../app.service';

import { ApiService } from '../shared/service/api.service';

import { LoginRequest } from '../shared/service/request/login';

@Component({
  selector: 'login',
  styleUrls: [],
  templateUrl: './login.component.html'
})
export class LoginComponent implements OnInit, OnDestroy {

  public model: LoginRequest = {
    email: '',
    password: ''
  };

  private returnUrl: string = '/';

  private subs: Subscription[] = [];

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private api: ApiService,
    private state: AppState,
    private permanent: Permanent
  ) { }

  public ngOnInit() {
    this.subs.push(
      this.route.queryParams
        .filter((params: Params) => params['returnUrl'])
        .subscribe((params: Params) => {
          console.log('setted');
          this.returnUrl = params['returnUrl'];
        })
    );
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public login() {
    this.subs.push(
      this.api.login(this.model)
        .subscribe(token => {
          this.permanent.set('token', token);
          this.state.set('token', token);

          this.router.navigate([this.returnUrl]);
        })
    );
  }
}
