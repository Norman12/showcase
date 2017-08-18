import { Resolve, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { ApiService } from './shared/service/api.service';

import { Project } from './shared/domain/project';
import { Content } from './shared/domain/content';
import { Theme } from './shared/domain/theme';
import { Menu } from './shared/domain/menu';
import { UserEditRequest } from './shared/service/request/user.edit';
import { MetaEditRequest } from './shared/service/request/meta.edit';
import { CredentialsEditRequest } from './shared/service/request/credentials.edit';
import { ProjectEditRequest } from './shared/service/request/project.edit';
import { ContentEditRequest } from './shared/service/request/content.edit';

@Injectable()
export class ProjectsResolver implements Resolve<Project[]> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<Project[]> {
    return this.api.getProjects();
  }
}

@Injectable()
export class ContentsResolver implements Resolve<Content[]> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<Content[]> {
    return this.api.getContents();
  }
}

@Injectable()
export class MenuResolver implements Resolve<Menu> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<Menu> {
    return this.api.getMenu();
  }
}

@Injectable()
export class AboutResolver implements Resolve<UserEditRequest> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<UserEditRequest> {
    return this.api.getUser();
  }
}

@Injectable()
export class SettingsResolver implements Resolve<[MetaEditRequest, CredentialsEditRequest, { selected: string, themes: Theme[] }]> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<[MetaEditRequest, CredentialsEditRequest, { selected: string, themes: Theme[] }]> {
    return Observable.zip(
      this.api.getMeta(),
      this.api.getCredentials(),
      this.api.getTheme(),

      function(s1, s2, s3): [MetaEditRequest, CredentialsEditRequest, { selected: string, themes: Theme[] }] {
        return [s1, s2, s3];
      }
    );
  }
}

@Injectable()
export class ProjectResolver implements Resolve<ProjectEditRequest> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<ProjectEditRequest> {
    return this.api.getProject(route.params['id']);
  }
}

@Injectable()
export class ContentResolver implements Resolve<ContentEditRequest> {

  constructor(
    private api: ApiService
  ) { }

  public resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<ContentEditRequest> {
    return this.api.getContent(route.params['id']);
  }
}

/**
 * An array of services to resolve routes with data.
 */
export const APP_RESOLVER_PROVIDERS = [
  ProjectsResolver,
  ContentsResolver,
  MenuResolver,
  AboutResolver,
  SettingsResolver,
  ProjectResolver,
  ContentResolver
];
