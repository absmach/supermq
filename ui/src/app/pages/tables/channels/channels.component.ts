import { Component } from '@angular/core';

import { ViewCell, LocalDataSource } from 'ng2-smart-table';
import { ButtonRenderComponent } from './button.render.component'
import { ConnectedRenderComponent } from './connected.render.component';
import { SmartTableService } from '../../../@core/data/smart-table.service';
import { Observable } from 'rxjs';

import { Channel } from '../../../@core/store/models';
import { ThingsStore } from '../../../@core/store/things.store';
import { ChannelsStore } from '../../../@core/store/channels.store';
import { ChannelsService } from '../../../@core/services/channels/channels.service';



@Component({
  selector: 'ngx-smart-table',
  templateUrl: './channels.component.html',
  styles: [`
    nb-card {
      transform: translate3d(0, 0, 0);
    }
  `],
})

export class ChannelsComponent {

  settings = {
    add: {
      addButtonContent: '<i class="nb-plus"></i>',
      createButtonContent: '<i class="nb-checkmark"></i>',
      cancelButtonContent: '<i class="nb-close"></i>',
    },
    edit: {
      editButtonContent: '<i class="nb-edit"></i>',
      saveButtonContent: '<i class="nb-checkmark"></i>',
      cancelButtonContent: '<i class="nb-close"></i>',
    },
    delete: {
      deleteButtonContent: '<i class="nb-trash"></i>',
      confirmDelete: true,
    },
    columns: {
      id: {
        title: 'ID',
        type: 'custom',
        renderComponent: ButtonRenderComponent,
        defaultValue: 'Connect'
      },
      name: {
        title: 'Name',
        type: 'string',
      },
      conncted: {
        title: 'List Devices',
        type: 'custom',
        renderComponent: ConnectedRenderComponent,
	filterFunction(cell?: any, search?: string): boolean {
      		const match = cell.indexOf(search) > -1;
		 console.log("Cell : ", match, search);
                 return true ;
        }
      },
    },
  };

  source: LocalDataSource = new LocalDataSource();
  channels: Observable<Channel[]>;
	
  constructor(
    private service: ChannelsService,
    public thingsStore: ThingsStore,
    public channelsStore: ChannelsStore,
  ) {

  }

  ngOnInit() {
    const data = this.channelsStore.getChannels();
    console.log("DATA1" ,data);
    //this.source.load(data);
    this.service.getChannels().subscribe((payload: any) => {
        console.log("DATA2" ,payload);
        this.channels = payload
        this.source.load(payload);
    });
    this.channelsStore.getChannels();
  }



  //constructor(private service: SmartTableService) {
    //const data = this.service.getData();
    //this.source.load(data);
  //}

  onDeleteConfirm(event): void {
    if (window.confirm('Are you sure you want to delete?')) {
      event.confirm.resolve();
    } else {
      event.confirm.reject();
    }
  }
}
