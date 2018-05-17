import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { Thing } from '../store/models';

const MOCK_CLIENTS = [
];

@Injectable()
export class MockClientsService {
  getClients() {
      return Observable.of(MOCK_CLIENTS).delay(1000);
  }

  addClient(thing: Thing) {
    MOCK_CLIENTS.push(thing);
    return Observable.of(1).delay(1000);
  }
}
