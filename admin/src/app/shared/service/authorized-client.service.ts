import { Injectable } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { HttpClient, HttpHeaders, HttpEvent, HttpInterceptor, HttpHandler, HttpRequest, HttpResponse, HttpErrorResponse, HttpEventType } from '@angular/common/http';

//import 'rxjs/add/operator/do';

import { Observable } from 'rxjs/Rx';

import { AppState } from '../../app.service';

import { Emitter } from './emitter.service';

import { ApiResponse } from './response/api';

import { Error } from '../domain/error';

@Injectable()
export class AuthorizedClient {

  constructor(
    private http: HttpClient,
    private state: AppState
  ) { }

  public get<T>(url): Observable<ApiResponse<T>> {
    return this.http.get<ApiResponse<T>>(url, {
      headers: this.createAuthorizationHeader(),
    })
  }

  public post<T>(url, data) {
    return this.http.post<ApiResponse<T>>(url, data, {
      headers: this.createAuthorizationHeader(),
    })
  }

  public put<T>(url, data) {
    return this.http.put<ApiResponse<T>>(url, data, {
      headers: this.createAuthorizationHeader(),
    })
  }

  public delete<T>(url) {
    return this.http.delete<ApiResponse<T>>(url, {
      headers: this.createAuthorizationHeader(),
    })
  }

  private createAuthorizationHeader(): HttpHeaders {
    if (this.state.get('token')) {
      return new HttpHeaders().set('Authorization', this.state.get('token'));
    } else {
      return new HttpHeaders();
    }
  }
}

@Injectable()
export class AuthorizationInterceptor implements HttpInterceptor {

  private errorChannel: string = 'error';

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next
      .handle(req)
      .catch(err => {
        return Observable.of(err)
      })
      .do(event => {
        if (event instanceof HttpResponse || event instanceof HttpErrorResponse) {
          switch (event.status) {
            case 401:
              Emitter.get(this.errorChannel).emit(new Error('You entered incorrect email or password.', event.status));
              break;
            case 403:
              Emitter.get(this.errorChannel).emit(new Error('You are not logged in.', event.status));
              break;
            case 500:
              Emitter.get(this.errorChannel).emit(new Error('Critical error occured.', event.status));
              break;
          }
        }

      });
  }

}