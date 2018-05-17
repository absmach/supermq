import { Component, OnInit } from '@angular/core';
import { MatDialog } from '@angular/material';
import { toJS } from 'mobx';
import { Observable } from 'rxjs/Observable';

import { Thing } from '../../core/store/models';
import { ConfirmationDialogComponent } from '../shared/confirmation-dialog/confirmation-dialog.component';
import { AddClientDialogComponent } from './add-client-dialog/add-client-dialog.component';
import { ThingsStore } from '../../core/store/things.store';
import { ChannelsStore } from '../../core/store/channels.store';

@Component({
  selector: 'app-clients',
  templateUrl: './clients.component.html',
  styleUrls: ['./clients.component.scss']
})
export class ClientsComponent implements OnInit {
  clients: Observable<Thing[]>;
  displayedColumns = ['id', 'name', 'type', 'payload', 'actions'];

  constructor(
    private dialog: MatDialog,
    public thingsStore: ThingsStore,
    public channelsStore: ChannelsStore,
  ) { }

  ngOnInit() {
    this.thingsStore.getThings();
    this.channelsStore.getChannels();
  }

  addThing() {
    const dialogRef = this.dialog.open(AddClientDialogComponent);

    dialogRef.componentInstance.submit.subscribe((thing: Thing) => {
      this.thingsStore.addThing(thing);
    });
  }

  editThing(thing: Thing) {
    const dialogRef = this.dialog.open(AddClientDialogComponent, {
      data: thing
    });

    dialogRef.componentInstance.submit.subscribe((editedThing: Thing) => {
      this.thingsStore.editThing(toJS(editedThing));
    });
  }

  deleteThing(client: Thing) {
    const dialogRef = this.dialog.open(ConfirmationDialogComponent, {
      data: {
        question: 'Are you sure you want to delete the thing?'
      }
    });

    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        this.thingsStore.deleteThing(toJS(client));
      }
    });
  }
}
