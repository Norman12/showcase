import { Injectable, EventEmitter } from '@angular/core';

@Injectable()
export class Emitter {

    private static store: { [ID: string]: EventEmitter<any> } = {};

    static get(ID: string): EventEmitter<any> {
        if (!this.store[ID]) {
            this.store[ID] = new EventEmitter();
        }
        return this.store[ID];
    }
}