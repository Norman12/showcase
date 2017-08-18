import { Component, OnInit, OnDestroy, Input, ViewChild } from '@angular/core';
import { Subscription } from 'rxjs/Rx';
import { ModalDialogComponent } from '../modal-dialog';
import { Emitter } from '../../service/emitter.service';

const enum State {
  Default,
  Waiting,
  Success,
  Failure,
  Timeout,
}

@Component({
  selector: 'result-button',
  template: `
  	<button type="button" [ngClass]="getClasses()">
      {{ getTitle() }}
    </button>
  `,
  styleUrls: ['./result-button.component.scss']
})
export class ResultButtonComponent implements OnInit, OnDestroy {

  @Input()
  public default: string;

  @Input()
  public waiting: string;

  @Input()
  public success: string;

  @Input()
  public failure: string;

  @Input()
  public timeout: string;

  @Input()
  public channel: string;

  private state: State = State.Default;

  private sub: Subscription;

  private timerRef: any;
  private defaultTimerRef: any;

  constructor(
  ) {
  }

  public ngOnInit() {
    this.sub = Emitter.get(this.channel)
      .filter(event => event instanceof ButtonEvent)
      .subscribe(event => {
        switch (event.constructor) {
          case WaitingEvent:
            this.clearDefaultTimerRef();
            this.state = State.Waiting;
            this.fireTimerCountdown(5000);
            break;
          case SuccessEvent:
            this.clearTimerRef();
            this.state = State.Success;
            this.fireDefaultCountdown(3000);
            break;
          case FailureEvent:
            this.clearTimerRef();
            this.state = State.Failure;
            this.fireDefaultCountdown(7000);
            break;
        }
      });
  }

  public ngOnDestroy() {
    this.clearTimerRef();
    this.clearDefaultTimerRef();

    if (this.sub != null)
      this.sub.unsubscribe();
  }

  public getClasses(): any {
    let classes = {
      'btn': true,
      'btn-lg': true,
      'pull-right': true
    }

    switch (this.state) {
      case State.Default:
        classes['btn-primary'] = true;
        break;
      case State.Waiting:
        classes['btn-primary'] = true;
        break;
      case State.Success:
        classes['btn-success'] = true;
        break;
      case State.Failure:
        classes['btn-danger'] = true;
        break;
      case State.Timeout:
        classes['btn-warning'] = true;
        break;
      default:
        classes['btn-primary'] = true;
    }

    return classes;
  }

  public getTitle(): string {
    switch (this.state) {
      case State.Default:
        return this.default;
      case State.Waiting:
        return this.waiting;
      case State.Success:
        return this.success;
      case State.Failure:
        return this.failure;
      case State.Timeout:
        return this.timeout;
      default:
        return this.default;
    }
  }

  private fireDefaultCountdown(millis: number) {
    this.clearDefaultTimerRef();

    let $this = this;

    this.defaultTimerRef = setTimeout(function() {
      $this.state = State.Default;
    }, millis);
  }

  private fireTimerCountdown(millis: number) {
    this.clearTimerRef();

    let $this = this;

    this.timerRef = setTimeout(function() {
      $this.state = State.Timeout;
      $this.fireDefaultCountdown(1000);
    }, millis);
  }

  private clearTimerRef() {
    if (this.timerRef != null)
      clearTimeout(this.timerRef);
  }

  private clearDefaultTimerRef() {
    if (this.defaultTimerRef != null)
      clearTimeout(this.defaultTimerRef);
  }
}

export class ButtonEvent {
}

export class WaitingEvent extends ButtonEvent {
}

export class SuccessEvent extends ButtonEvent {
}

export class FailureEvent extends ButtonEvent {
}