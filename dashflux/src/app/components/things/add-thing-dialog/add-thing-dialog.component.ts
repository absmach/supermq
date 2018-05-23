import { Component, EventEmitter, Inject, OnInit, Output } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material';

import { Thing } from '../../../core/store/models';

@Component({
  selector: 'app-add-thing-dialog',
  templateUrl: './add-thing-dialog.component.html',
  styleUrls: ['./add-thing-dialog.component.scss']
})
export class AddThingDialogComponent implements OnInit {
  addThingForm: FormGroup;
  @Output() submit: EventEmitter<Thing> = new EventEmitter<Thing>();

  constructor(
    private fb: FormBuilder,
    private dialogRef: MatDialogRef<AddThingDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: Thing
  ) { }

  ngOnInit() {
    this.addThingForm = this.fb.group(
      {
        id: [''],
        type: ['', [Validators.required]],
        name: ['', [Validators.required, Validators.minLength(5)]],
        payload: ['']
      }
    );

    if (this.data) {
      this.addThingForm.patchValue(this.data);
      this.addThingForm.get('payload').patchValue(JSON.stringify(this.data.payload));
    }
  }

  onAddThing() {
    const thing = this.addThingForm.value;
    this.submit.emit(thing);
    this.dialogRef.close();
  }
}
