import { Injectable } from '@angular/core';
import { Router, CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';

import { AppState } from './app.service'

@Injectable()
export class AuthenticateActivate implements CanActivate {

    constructor(
        private router: Router,
        private state: AppState
    ) { }

    canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean {
        if (this.state.get('token')) {
            return true;
        }

        this.router.navigate(['/login'], { queryParams: { returnUrl: state.url }});

        return false;
    }
}

export const APP_ACTIVATE_PROVIDERS = [
    AuthenticateActivate
];
