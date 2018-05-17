import { HttpClientModule } from '@angular/common/http';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';

import { MaterialModule } from '../../core/material/material.module';
import { AuthenticationService } from '../../core/services/auth/authentication.service';
import { TokenStorage } from '../../core/services/auth/token-storage.service';
import { ChannelsService } from '../../core/services/channels/channels.service';
import { ThingsService } from '../../core/services/clients/clients.service';
import { ClientsComponent } from './clients.component';
import { UiStore } from '../../core/store/ui.store';
import { ThingsStore } from '../../core/store/clients.store';
import { ChannelsStore } from '../../core/store/channels.store';

describe('ClientsComponent', () => {
  let component: ClientsComponent;
  let fixture: ComponentFixture<ClientsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ClientsComponent ],
      imports: [
        MaterialModule,
        MatDialogModule,
        HttpClientModule,
        RouterTestingModule,
        FormsModule,
        ReactiveFormsModule,
        NoopAnimationsModule,
      ],
      providers: [
        UiStore,
        {
          provide: ThingsStore,
          useClass: class {
            getThings = jasmine.createSpy('getThings');
          }
        },
        ChannelsStore,
        AuthenticationService,
        TokenStorage,
        ThingsService,
        ChannelsService,
        { provide: MatDialogRef, useValue: {} },
        { provide: MAT_DIALOG_DATA, useValue: [] },
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ClientsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
