/**
 * Angular 2 decorators and services
 */
import {
  Component,
  OnInit,
  AfterViewInit,
  OnDestroy,
  ViewChild,
  ElementRef,
  Renderer,
  ViewEncapsulation
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { Router, NavigationEnd } from '@angular/router';

import { AppState } from './app.service';

import { ApiService } from './shared/service/api.service';
import { Emitter } from './shared/service/emitter.service';

import { ModalDialogComponent } from './shared/view/modal-dialog';

import { Error } from './shared/domain/error';

/**
 * App Component
 * Top Level Component
 */
@Component({
  selector: 'app',
  encapsulation: ViewEncapsulation.None,
  styleUrls: [
    './app.component.scss'
  ],
  template: `
      <nav class="navbar navbar-default navbar-fixed-top">
        <div class="container-fluid">

          <div class="navbar-header">
            <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1" aria-expanded="false" *ngIf="controlsVisible">
              <span class="sr-only">Toggle navigation</span>
              <span class="icon-bar"></span>
              <span class="icon-bar"></span>
              <span class="icon-bar"></span>
            </button>
            <a class="navbar-brand" [routerLink]=" ['./'] ">Showcase</a>
          </div>

          <div #collapse class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
            <ul class="nav navbar-nav">
              <li>
                <a [routerLink]=" ['./'] "
                  routerLinkActive="active" [routerLinkActiveOptions]= "{exact: true}">
                  Projects
                </a>
              </li>
              <li>
                <a [routerLink]=" ['./pages'] "
                  routerLinkActive="active" [routerLinkActiveOptions]= "{exact: true}">
                  Pages
                </a>
              </li>
              <li>
                <a [routerLink]=" ['./menu'] "
                  routerLinkActive="active" [routerLinkActiveOptions]= "{exact: true}">
                  Menu
                </a>
              </li>
              <li>
                <a [routerLink]=" ['./about'] "
                  routerLinkActive="active" [routerLinkActiveOptions]= "{exact: true}">
                  About Me
                </a>
              </li>
              <li>
                <a [routerLink]=" ['./settings'] "
                  routerLinkActive="active" [routerLinkActiveOptions]= "{exact: true}">
                  Settings
                </a>
              </li>
            </ul>

            <ul class="nav navbar-nav navbar-right">
              <button type="button" class="btn btn-success navbar-btn" style="margin: 8px 8px 8px 16px;" (click)="openSite()">View site</button>
              <button type="button" class="btn btn-default navbar-btn" style="margin: 8px 8px 8px 0px;" (click)="logout()">Logout</button>
            </ul>
          </div>

        </div>
      </nav>

    <main>
      <div class="container">
        <router-outlet></router-outlet>
      </div>
    </main>

    <footer>
      <div class="container">
        <p>Powered by Showcase</p>
      </div>
    </footer>

    <modal-dialog #modalError (closed)="onClosed($event)">
      <div class="app-modal-header">
        Error
      </div>
      <div class="app-modal-body">
        {{errorMessage}}
      </div>
      <div class="app-modal-footer">
        <button type="button" class="btn btn-primary" (click)="modalError.hide()">Try again</button>
      </div>
    </modal-dialog>
  `
})
export class AppComponent implements OnInit, OnDestroy, AfterViewInit {

  @ViewChild('collapse')
  collapse: ElementRef;

  @ViewChild('modalError')
  modalError: ModalDialogComponent;

  private site: string = '/';

  private siteChannel: string = 'site';
  private errorChannel: string = 'error';

  private errorShowing: boolean = false;
  private errorMessage: string = '';

  private subs: Subscription[] = [];

  constructor(
    private router: Router,
    private renderer: Renderer,
    private state: AppState,
    private api: ApiService
  ) {

  }

  public ngOnInit() {
    this.reloadSite();
    this.subs.push(
      Emitter.get(this.siteChannel)
        .subscribe(m => this.reloadSite())
    )
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public ngAfterViewInit() {
    this.subs.push(
      this.router.events
        .filter(event => event instanceof NavigationEnd)
        .subscribe(event => {
          this.renderer.setElementClass(this.collapse.nativeElement, 'in', false);
          if ((event as NavigationEnd).url.startsWith('/login')) {
            this.renderer.setElementStyle(this.collapse.nativeElement, 'visibility', 'hidden');
          } else {
            this.renderer.setElementStyle(this.collapse.nativeElement, 'visibility', 'visible');
          }
        }
        )
    );

    this.subs.push(
      Emitter.get(this.errorChannel)
        .filter(obj => obj instanceof Error)
        .subscribe((error: Error) => {
          if (!this.errorShowing) {
            this.errorMessage = error.content;
            this.modalError.show();
          }

          switch (error.code) {
            case 401:
            case 403:
              this.logout();
              break;
          }
        })
    );
  }

  public openSite() {
    window.open(this.site, "_blank");
  }

  public logout() {
    this.state.delete('token');
    this.router.navigate(['/login']);
  }

  public onClosed(event) {
    this.errorShowing = false;
  }

  private reloadSite() {
    this.subs.push(
      this.api.getSite()
        .subscribe(site => this.site = site)
    );
  }
}