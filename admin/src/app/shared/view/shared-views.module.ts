import {
  NgModule
} from '@angular/core';

import { FormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';

import { MediaInputComponent } from './media-input';
import { TagInputComponent } from './tag-input';
import { KeyValueInputComponent } from './key-value-input';
import { ThemeInputComponent } from './theme-input';
import { ParagraphInputComponent } from './paragraph-input';
import { ModalDialogComponent } from './modal-dialog';
import { ResultButtonComponent } from './result-button';

import { KeysPipe } from '../pipe/keys';

import { CKEditorModule } from 'ng2-ckeditor';

@NgModule({
  declarations: [
    MediaInputComponent,
    TagInputComponent,
    KeyValueInputComponent,
    ThemeInputComponent,
    ParagraphInputComponent,
    ModalDialogComponent,
    ResultButtonComponent,
    KeysPipe
  ],
  exports: [
    MediaInputComponent,
    TagInputComponent,
    KeyValueInputComponent,
    ThemeInputComponent,
    ParagraphInputComponent,
    ModalDialogComponent,
    ResultButtonComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    CKEditorModule
  ]
})
export class SharedViewsModule {
  constructor(
  ) { }
}
