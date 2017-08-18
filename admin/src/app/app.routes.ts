import { Routes } from '@angular/router';
import { HomeComponent } from './home';
import { AboutComponent } from './about';
import { MenuComponent } from './menu';
import { ProjectsComponent } from './projects';
import { PagesComponent } from './pages';
import { AddContentComponent } from './add-content';
import { EditContentComponent } from './edit-content';
import { SettingsComponent } from './settings';
import { AddProjectComponent } from './add-project';
import { EditProjectComponent } from './edit-project';
import { NoContentComponent } from './no-content';
import { LoginComponent } from './login';

import { ProjectsResolver, ContentsResolver, AboutResolver, SettingsResolver, ProjectResolver, ContentResolver, MenuResolver } from './app.resolver';
import { AuthenticateActivate } from './app.activate';

export const ROUTES: Routes = [
  { path: '', component: ProjectsComponent, resolve: { projects: ProjectsResolver }, canActivate: [AuthenticateActivate] },
  { path: 'about', component: AboutComponent, resolve: { about: AboutResolver }, canActivate: [AuthenticateActivate] },
  { path: 'menu', component: MenuComponent, resolve: { menu: MenuResolver }, canActivate: [AuthenticateActivate] },
  { path: 'projects', component: ProjectsComponent, resolve: { projects: ProjectsResolver }, canActivate: [AuthenticateActivate] },
  { path: 'add-project', component: AddProjectComponent, canActivate: [AuthenticateActivate] },
  { path: 'edit-project/:id', component: EditProjectComponent, resolve: { project: ProjectResolver }, canActivate: [AuthenticateActivate] },
  { path: 'pages', component: PagesComponent, resolve: { contents: ContentsResolver }, canActivate: [AuthenticateActivate] },
  { path: 'add-content', component: AddContentComponent, canActivate: [AuthenticateActivate] },
  { path: 'edit-content/:id', component: EditContentComponent, resolve: { content: ContentResolver }, canActivate: [AuthenticateActivate] },
  { path: 'settings', component: SettingsComponent, resolve: { settings: SettingsResolver }, canActivate: [AuthenticateActivate] },
  { path: 'login', component: LoginComponent },
  { path: '**', component: NoContentComponent, canActivate: [AuthenticateActivate] },
];
