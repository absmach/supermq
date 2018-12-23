import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

import { TablesComponent } from './tables.component';
import { SmartTableComponent } from './smart-table/smart-table.component';
import { ThingsComponent } from './things/things.component';
import { ChannelsComponent } from './channels/channels.component';
import { ButtonRenderComponent } from './channels/button.render.component';


const routes: Routes = [{
  path: '',
  component: TablesComponent,
  children: [{
    path: 'smart-table',
    component: SmartTableComponent,
  },
  {
    path: 'things',
    component: ThingsComponent,
  },
  {
    path: 'channels',
    component: ChannelsComponent,
  }],
}];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class TablesRoutingModule { }

export const routedComponents = [
  TablesComponent,
  SmartTableComponent,
  ChannelsComponent,
  ThingsComponent,
];
