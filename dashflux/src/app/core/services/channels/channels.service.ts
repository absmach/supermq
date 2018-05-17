import { HttpClient, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { forkJoin } from 'rxjs/observable/forkJoin';

import { environment } from '../../../../environments/environment';
import { Channel, Thing } from '../../store/models';

interface ChannelsPayload {
  channels: Channel[];
}

@Injectable()
export class ChannelsService {

  constructor(private http: HttpClient) { }

  getChannels() {
    return this.http.get(environment.channelsUrl).switchMap((payload: ChannelsPayload) => {
      const allChannels = forkJoin(this.createChannelsRequests(payload.channels));
      return allChannels;
    }).switchMap((responses: Channel[]) => {
      responses.forEach(channel => {
        channel.connected = channel.connected ? channel.connected : [];
      });
      return Observable.of(responses);
    });
  }

  createChannelsRequests(channels) {
    return channels.map((channel => this.http.get(environment.channelsUrl + '/' + channel.id)));
  }

  addChannel(channel: Channel) {
    const payload = {
      name: channel.name
    };

    return this.http.post(environment.channelsUrl, payload, { observe: 'response' })
      .switchMap((res) => {
        const id = this.getChannelIdFrom(res);
        return forkJoin(this.createThingsConnectRequests(id, channel.connected));
      });
  }

  private getChannelIdFrom(res: HttpResponse<Object>) {
    const location = res.headers.get('Location');
    return location.replace('/channels/', '');
  }

  deleteChannel(channel: Channel) {
    return this.http.delete(environment.channelsUrl + '/' + channel.id);
  }

  editChannel(channelFormData: Channel, channel: Channel) {
    const payload = {
      name: channelFormData.name
    };

    const editChannel = this.http.put(environment.channelsUrl + '/' + channel.id, payload);

    return editChannel.switchMap(() => {
      const clientsToAdd = this.getThingsToAdd(channelFormData, channel);
      if (clientsToAdd.length) {
        return forkJoin(this.createThingsConnectRequests(channel.id, clientsToAdd));
      } else {
        return Observable.of([]);
      }
    }).switchMap(() => {
      const clientsToDelete = this.getClientsToDelete(channelFormData, channel);
      console.log(clientsToDelete);
      if (clientsToDelete.length) {
        return forkJoin(this.createThingsDisconnectRequests(channel.id, clientsToDelete));
      } else {
        return Observable.of([]);
      }
    });
  }

  getClientsToDelete(channelFormData: Channel, channel: Channel) {
    return channel.connected.filter(thing => {
      return channelFormData.connected.find(th => th.id === thing.id) === undefined;
    });
  }

  getThingsToAdd(channelFormData: Channel, channel: Channel) {
    return channelFormData.connected.filter(thing => {
      return channel.connected.find(th => th.id === thing.id) === undefined;
    });
  }

  createThingsConnectRequests(channelId: string , connected: Thing[]) {
    return connected.map((connection) => {
      return this.http.put(environment.channelsUrl + '/' + channelId + '/things/' + connection.id, {});
    });
  }

  createThingsDisconnectRequests(channelId: string , connected: Thing[]) {
    return connected.map((connection) => {
      return this.http.delete(environment.channelsUrl + '/' + channelId + '/things/' + connection.id, {});
    });
  }
}
