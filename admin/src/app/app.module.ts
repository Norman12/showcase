import { BrowserModule } from '@angular/platform-browser';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';
import {
  NgModule,
  ApplicationRef,
  isDevMode
} from '@angular/core';
import {
  removeNgStyles,
  createNewHosts,
  createInputTransfer
} from '@angularclass/hmr';
import {
  RouterModule,
  PreloadAllModules
} from '@angular/router';

import { SharedServicesModule } from './shared/service/shared-services.module';
import { SharedViewsModule } from './shared/view/shared-views.module';

import { CKEditorModule } from 'ng2-ckeditor';

/*
 * Platform and Environment providers/directives/pipes
 */
import { ENV_PROVIDERS } from './environment';
import { ROUTES } from './app.routes';

import { AppComponent } from './app.component';
import { APP_RESOLVER_PROVIDERS } from './app.resolver';
import { APP_ACTIVATE_PROVIDERS } from './app.activate';
import { AppState, InternalStateType, Formatter, Permanent, Generator } from './app.service';

import { PagesComponent } from './pages';
import { AddContentComponent } from './add-content';
import { EditContentComponent } from './edit-content';

import { ProjectsComponent } from './projects';
import { AddProjectComponent } from './add-project';
import { EditProjectComponent } from './edit-project';

import { SettingsComponent } from './settings';
import { AboutComponent } from './about';
import { MenuComponent } from './menu';

import { NoContentComponent } from './no-content';
import { LoginComponent } from './login';

import '../styles/styles.scss';

// Application wide providers
const APP_PROVIDERS = [
  ...APP_RESOLVER_PROVIDERS,
  ...APP_ACTIVATE_PROVIDERS,
  AppState,
  Formatter,
  Permanent,
  Generator
];

type StoreType = {
  state: InternalStateType,
  restoreInputValues: () => void,
  disposeOldHosts: () => void
};

/**
 * `AppModule` is the main entry point into Angular2's bootstraping process
 */
@NgModule({
  bootstrap: [AppComponent],
  declarations: [
    AppComponent,
    AboutComponent,
    MenuComponent,
    ProjectsComponent,
    PagesComponent,
    AddContentComponent,
    EditContentComponent,
    SettingsComponent,
    AddProjectComponent,
    EditProjectComponent,
    NoContentComponent,
    LoginComponent
  ],
  /**
   * Import Angular's modules.
   */
  imports: [
    BrowserModule,
    FormsModule,
    HttpModule,
    SharedServicesModule,
    SharedViewsModule,
    CKEditorModule,
    RouterModule.forRoot(ROUTES, { useHash: true, preloadingStrategy: PreloadAllModules })
  ],
  /**
   * Expose our Services and Providers into Angular's dependency injection.
   */
  providers: [
    ENV_PROVIDERS,
    APP_PROVIDERS
  ]
})
export class AppModule {

  constructor(
    public appRef: ApplicationRef,
    public appState: AppState,
    public permanent: Permanent,
  ) {
    appState.set("token", permanent.get("token"));

    if (isDevMode()) {
      appState.set("api.root", "http://localhost:8080/admin/");
    } else {
      appState.set("api.root", "/admin/");
    }

    appState.set("toolbar", {
      toolbarGroups: [
        { name: 'clipboard', groups: ['clipboard', 'undo'] },
        { name: 'editing', groups: ['find', 'selection', 'spellchecker', 'editing'] },
        { name: 'links', groups: ['links'] },
        { name: 'insert', groups: ['insert'] },
        { name: 'forms', groups: ['forms'] },
        { name: 'tools', groups: ['tools'] },
        { name: 'document', groups: ['mode', 'document', 'doctools'] },
        { name: 'others', groups: ['others'] },
        '/',
        { name: 'basicstyles', groups: ['basicstyles', 'cleanup'] },
        { name: 'paragraph', groups: ['list', 'indent', 'blocks', 'align', 'bidi', 'paragraph'] },
        { name: 'styles', groups: ['styles'] },
        { name: 'colors', groups: ['colors'] },
        { name: 'about', groups: ['about'] }
      ],

      removeButtons: 'Underline,Subscript,Superscript,About'
    });

  }

  public hmrOnInit(store: StoreType) {
    if (!store || !store.state) {
      return;
    }
    console.log('HMR store', JSON.stringify(store, null, 2));
    /**
     * Set state
     */
    this.appState._state = store.state;
    /**
     * Set input values
     */
    if ('restoreInputValues' in store) {
      let restoreInputValues = store.restoreInputValues;
      setTimeout(restoreInputValues);
    }

    this.appRef.tick();
    delete store.state;
    delete store.restoreInputValues;
  }

  public hmrOnDestroy(store: StoreType) {
    const cmpLocation = this.appRef.components.map((cmp) => cmp.location.nativeElement);
    /**
     * Save state
     */
    const state = this.appState._state;
    store.state = state;
    /**
     * Recreate root elements
     */
    store.disposeOldHosts = createNewHosts(cmpLocation);
    /**
     * Save input values
     */
    store.restoreInputValues = createInputTransfer();
    /**
     * Remove styles
     */
    removeNgStyles();
  }

  public hmrAfterDestroy(store: StoreType) {
    /**
     * Display new elements
     */
    store.disposeOldHosts();
    delete store.disposeOldHosts;
  }

}
