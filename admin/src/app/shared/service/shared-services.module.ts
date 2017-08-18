import {
  NgModule
} from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';

import { Emitter } from './emitter.service';
import { AuthorizedClient, AuthorizationInterceptor } from './authorized-client.service';
import { ApiService } from './api.service';

import { ApiResponse } from './api-response.class';
import { PostResponse } from './response/post-response.class';

@NgModule({
  imports: [
    HttpClientModule
  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthorizationInterceptor,
      multi: true,
    },
    Emitter,
    AuthorizedClient,
    ApiService
  ]
})
export class SharedServicesModule {
  constructor
    () { }
}
