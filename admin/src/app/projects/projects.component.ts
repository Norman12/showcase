import {
  Component,
  OnInit,
  OnDestroy,
  ViewChild
} from '@angular/core';

import { Observable, Subscription } from 'rxjs/Rx';

import { Router, ActivatedRoute } from '@angular/router';

import { Project } from '../shared/domain/project';

import { ModalDialogComponent } from '../shared/view/modal-dialog';

import { ApiService } from '../shared/service/api.service';
import { ProjectDeleteRequest } from '../shared/service/request/project.delete';

@Component({
  selector: 'projects',
  styleUrls: ['./projects.component.scss'],
  templateUrl: './projects.component.html'
})
export class ProjectsComponent implements OnInit, OnDestroy {

  public projects: Project[];
  public selected: Project = new Project("", "", "", "");

  @ViewChild('modalDelete')
  modalDelete: ModalDialogComponent;

  private subscriptions: Subscription[] = [];

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private api: ApiService,
  ) { }

  public ngOnInit() {
    this.projects = this.route.snapshot.data['projects'];
  }

  public openAddProject() {
    this.router.navigate(['/add-project']);
  }

  public openEdit(project: Project) {
    this.router.navigate(['/edit-project', project.slug]);
  }

  public openDelete(project: Project) {
    this.selected = project;
    this.modalDelete.show();
  }

  public delete() {
    this.subscriptions.push(
      this.api.deleteProject(this.selected.slug).subscribe(result => {
        if (result) {
          this.subscriptions.push(
            this.api.getProjects().subscribe(
              projects => this.projects = projects
            )
          )
          this.modalDelete.hide();
        } else {
          alert("Something went wrong");
        }
      })
    );
  }

  public ngOnDestroy() {
    for (let s of this.subscriptions) {
      s.unsubscribe();
    }
  }
}
