import {
  Component,
  OnInit,
  OnDestroy
} from '@angular/core';

import { Subscription } from 'rxjs/Rx';

import { ActivatedRoute } from '@angular/router';

import { ApiService } from '../shared/service/api.service';

import { MenuAddRequest } from '../shared/service/request/menu.add';
import { MenuRemoveRequest } from '../shared/service/request/menu.remove';

import { Menu } from '../shared/domain/menu';

@Component({
  selector: 'menu',
  styleUrls: ['./menu.component.scss'],
  templateUrl: './menu.component.html'
})
export class MenuComponent implements OnInit, OnDestroy {

  public model: Menu = {
    added: [],
    routes: []
  }

  private subs: Subscription[] = [];

  constructor(
    private route: ActivatedRoute,
    private api: ApiService,
  ) { }

  public ngOnInit() {
    this.model = this.route.snapshot.data['menu'];
  }

  public ngOnDestroy() {
    for (let s of this.subs) {
      s.unsubscribe();
    }
  }

  public addRoute(route: string) {
    this.subs.push(
      this.api.addToMenu({slug : route})
        .subscribe(result => { if(result) this.reload() })
    );
  }

  public removeRoute(route: string) {
    this.subs.push(
      this.api.removeFromMenu({slug : route})
        .subscribe(result => { if(result) this.reload() })
    );
  }

  private reload(){
    this.subs.push(
      this.api.getMenu()
        .subscribe(menu => this.model = menu)
    );
  }
}
