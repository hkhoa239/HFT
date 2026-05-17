import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { QrWorkspaceComponent } from './qr-workspace.component';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DataService } from '../../services/data.service';
import { NotificationService } from '../../services/notification.service';
import { APP_CONFIG } from '../../app.constants';
import { of, throwError } from 'rxjs';

describe('QrWorkspaceComponent', () => {
  let component: QrWorkspaceComponent;
  let fixture: ComponentFixture<QrWorkspaceComponent>;
  let httpMock: HttpTestingController;
  let notificationService: jasmine.SpyObj<NotificationService>;
  let dataService: jasmine.SpyObj<DataService>;

  beforeEach(async () => {
    const nsSpy = jasmine.createSpyObj('NotificationService', ['success', 'error', 'info', 'warning']);
    const dsSpy = jasmine.createSpyObj('DataService', ['getVariables']);
    dsSpy.getVariables.and.returnValue({});

    await TestBed.configureTestingModule({
      imports: [QrWorkspaceComponent, HttpClientTestingModule],
      providers: [
        { provide: NotificationService, useValue: nsSpy },
        { provide: DataService, useValue: dsSpy }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(QrWorkspaceComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    notificationService = TestBed.inject(NotificationService) as jasmine.SpyObj<NotificationService>;
    dataService = TestBed.inject(DataService) as jasmine.SpyObj<DataService>;
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should start backtest after checking alpha registry', () => {
    component.qrCode = 'print("hello")';
    component.runBacktest();

    const reqMe = httpMock.expectOne(`${APP_CONFIG.apiUrl}/alphas/me`);
    expect(reqMe.request.method).toBe('GET');
    reqMe.flush({ success: true, data: [{ id: 'alpha-123' }] });

    const reqRun = httpMock.expectOne(`${APP_CONFIG.apiUrl}/backtest/run`);
    expect(reqRun.request.method).toBe('POST');
    expect(reqRun.request.body.alpha_id).toBe('alpha-123');
    reqRun.flush({ success: true, data: { id: 'job-456' } });

    expect(component.isSimulating).toBeTrue();
    expect(component.simStatus).toBe('Job Queued...');
  });

  it('should poll status and handle completion', fakeAsync(() => {
    // Manually trigger startPolling
    (component as any).startPolling('job-456');
    
    // First poll
    tick(2000);
    const req1 = httpMock.expectOne(`${APP_CONFIG.apiUrl}/backtest/job-456`);
    req1.flush({ success: true, data: { status: 'running' } });
    expect(component.simStatus).toBe('Processing Windows...');

    // Second poll - completion
    tick(2000);
    const req2 = httpMock.expectOne(`${APP_CONFIG.apiUrl}/backtest/job-456`);
    req2.flush({ 
      success: true, 
      data: { 
        status: 'completed', 
        metrics: { total_pnl: 10, win_rate: 0.6, sharpe_ratio: 1.5, trade_count: 5, pnl_curve: [1, 2, 3] } 
      } 
    });

    expect(component.isSimulating).toBeFalse();
    expect(component.simProgress).toBe(100);
    expect(notificationService.success).toHaveBeenCalledWith('Backtest complete!');
  }));

  it('should handle backtest failure', fakeAsync(() => {
    (component as any).startPolling('job-456');
    
    tick(2000);
    const req = httpMock.expectOne(`${APP_CONFIG.apiUrl}/backtest/job-456`);
    req.flush({ success: true, data: { status: 'failed', error_log: 'Syntax Error' } });

    expect(component.isSimulating).toBeFalse();
    expect(notificationService.error).toHaveBeenCalledWith('Backtest failed: Syntax Error');
  }));
});
