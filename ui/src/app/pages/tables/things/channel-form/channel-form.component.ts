import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { ViewCell, Cell, DefaultEditor, Editor } from 'ng2-smart-table';
import { NbWindowRef } from '@nebular/theme';
import { Observable } from 'rxjs';
import { Channel } from '../../../../@core/store/models';
import { ChannelsService } from '../../../../@core/services/channels/channels.service';

@Component({
  template: `
    <form class="form">
      <label for="subject">Subject:</label>
      <input nbInput id="subject" type="text">

      <label class="text-label" for="text">Text:</label>
      <textarea nbInput id="text"></textarea>
    </form>
  `,
  styleUrls: ['channel-form.component.scss'],
})

export class ChannelFormComponent {
  constructor(
	public windowRef: NbWindowRef,
	private service: ChannelsService,
	) {}


  channels: Observable<Channel[]>;


  ngOnInit() {

    this.service.getChannels().subscribe((payload: any) => {
        console.log("DATA2" ,payload);
        this.channels = payload
        for (let channel of payload) {
        	const id = channel.id;
        	if ('connected' in channel) {
            		const devices = channel['connected'];
			for (let device of devices) {
				if ( device.name == 'tlv1_BTS1' ) {
					console.log("id",id , "device", device.name);
				}
            		}
        	}
    	}
    });
  }


  close() {
    this.windowRef.close();
  }
}
