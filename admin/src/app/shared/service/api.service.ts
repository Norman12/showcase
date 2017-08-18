import { Component, Injectable } from '@angular/core';
import { AuthorizedClient } from './authorized-client.service';
import { Emitter } from './emitter.service';

import { AppState, Formatter } from '../../app.service';

import { Observable } from 'rxjs/Rx';

import { ApiResponse } from './response/api';
import { Project } from '../domain/project';
import { Content } from '../domain/content';
import { Theme } from '../domain/theme';
import { Menu } from '../domain/menu';
import { Error } from '../domain/error';

import { ProjectCreateRequest } from './request/project.create';
import { ProjectEditRequest } from './request/project.edit';
import { ProjectDeleteRequest } from './request/project.delete';

import { ContentCreateRequest } from './request/content.create';
import { ContentEditRequest } from './request/content.edit';
import { ContentDeleteRequest } from './request/content.delete';

import { UserEditRequest } from './request/user.edit';
import { ThemeEditRequest } from './request/theme.edit';
import { MetaEditRequest } from './request/meta.edit';
import { CredentialsEditRequest } from './request/credentials.edit';

import { MenuAddRequest } from './request/menu.add';
import { MenuRemoveRequest } from './request/menu.remove';

import { LoginRequest } from './request/login';

@Injectable()
export class ApiService {

  private endpoints: { [index: string]: string } = {
    'projects.get': 'projects',
    'contents.get': 'contents',

    'project.get': 'project/{0}',
    'project.add': 'project/create',
    'project.edit': 'project/{0}/update',
    'project.delete': 'project/{0}/delete',

    'content.get': 'content/{0}',
    'content.add': 'content/create',
    'content.edit': 'content/{0}/update',
    'content.delete': 'content/{0}/delete',

    'user.get': 'user',
    'user.edit': 'user/update',

    'theme.get': 'theme',
    'theme.edit': 'theme/update',

    'meta.get': 'meta',
    'meta.edit': 'meta/update',

    'credentials.get': 'credentials',
    'credentials.edit': 'credentials/update',

    'menu.get': 'menu',
    'menu.add': 'menu/add',
    'menu.remove': 'menu/remove',

    'site.get': 'site',

    'login': 'login',
    'logout': 'logout',
  };

  private errorChannel: string = 'error';

  constructor(
    private http: AuthorizedClient,
    private state: AppState,
    private formatter: Formatter
  ) {
  }

  public getProjects(): Observable<Project[]> {
    return this.extract<Project[]>(
      this.http.get(this.state.get("api.root") + this.endpoints['projects.get'])
    );
  }

  public getContents(): Observable<Content[]> {
    return this.extract<Content[]>(
      this.http.get(this.state.get("api.root") + this.endpoints['contents.get'])
    );
  }

  public getUser(): Observable<UserEditRequest> {
    return this.extract<UserEditRequest>(
      this.http.get(this.state.get("api.root") + this.endpoints['user.get'])
    );
  }

  public editUser(request: UserEditRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.endpoints['user.edit'], request)
    )
  }

  public getTheme(): Observable<{ selected: string, themes: Theme[] }> {
    return this.extract<{ selected: string, themes: Theme[] }>(
      this.http.get(this.state.get("api.root") + this.endpoints['theme.get'])
    );
  }

  public editTheme(request: ThemeEditRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.endpoints['theme.edit'], request)
    )
  }

  public getMeta(): Observable<MetaEditRequest> {
    return this.extract<MetaEditRequest>(
      this.http.get(this.state.get("api.root") + this.endpoints['meta.get'])
    );
  }

  public editMeta(request: MetaEditRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.endpoints['meta.edit'], request)
    )
  }

  public getCredentials(): Observable<CredentialsEditRequest> {
    return this.extract<CredentialsEditRequest>(
      this.http.get(this.state.get("api.root") + this.endpoints['credentials.get'])
    );
  }

  public editCredentials(request: CredentialsEditRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.endpoints['credentials.edit'], request)
    )
  }

  public getProject(slug: string): Observable<ProjectEditRequest> {
    return this.extract<ProjectEditRequest>(
      this.http.get(this.state.get("api.root") + this.formatter.format(this.endpoints['project.get'], slug))
    );
  }

  public createProject(request: ProjectCreateRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.post(this.state.get("api.root") + this.endpoints['project.add'], request)
    )
  }

  public editProject(request: ProjectEditRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.formatter.format(this.endpoints['project.edit'], request.slug), request)
    )
  }

  public deleteProject(slug: string): Observable<boolean> {
    return this.extract<boolean>(
      this.http.delete(this.state.get("api.root") + this.formatter.format(this.endpoints['project.delete'], slug))
    )
  }

  public getContent(slug: string): Observable<ContentEditRequest> {
    return this.extract<ContentEditRequest>(
      this.http.get(this.state.get("api.root") + this.formatter.format(this.endpoints['content.get'], slug))
    );
  }

  public createContent(request: ContentCreateRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.post(this.state.get("api.root") + this.endpoints['content.add'], request)
    )
  }

  public editContent(request: ContentEditRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.formatter.format(this.endpoints['content.edit'], request.slug), request)
    )
  }

  public deleteContent(slug: string): Observable<boolean> {
    return this.extract<boolean>(
      this.http.delete(this.state.get("api.root") + this.formatter.format(this.endpoints['content.delete'], slug))
    )
  }

  public getMenu(): Observable<Menu> {
    return this.extract<Menu>(
      this.http.get(this.state.get("api.root") + this.endpoints['menu.get'])
    );
  }

  public addToMenu(request: MenuAddRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.endpoints['menu.add'], request)
    );
  }

  public removeFromMenu(request: MenuRemoveRequest): Observable<boolean> {
    return this.extract<boolean>(
      this.http.put(this.state.get("api.root") + this.endpoints['menu.remove'], request)
    );
  }

  public login(request: LoginRequest): Observable<string> {
    return this.extract<string>(
      this.http.post(this.state.get("api.root") + this.endpoints['login'], request)
    );
  }

  public getSite(): Observable<string> {
    return this.extract<string>(
      this.http.get(this.state.get("api.root") + this.endpoints['site.get'])
    );
  }

  private extract<T>(response: Observable<ApiResponse<T>>): Observable<T> {
    return response.flatMap(r => {
      if (r.error != null) {
        Emitter.get(this.errorChannel).emit(new Error(r.error, -1));
        return Observable.empty();
      } else {
        return Observable.of(r.content);
      }
    }
    )
  }

}