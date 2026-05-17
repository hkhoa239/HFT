import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { PmWorkspaceComponent } from './pm-workspace.component';
import { DataService } from '../../services/data.service';
import { APP_CONFIG } from '../../app.constants';
import { of, throwError } from 'rxjs';
import { DomSanitizer } from '@angular/platform-browser';

describe('PmWorkspaceComponent', () => {
  let component: PmWorkspaceComponent;
  let fixture: ComponentFixture<PmWorkspaceComponent>;
  let httpMock: HttpTestingController;
  let dataService: DataService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PmWorkspaceComponent, HttpClientTestingModule],
      providers: [DataService]
    }).compileComponents();

    fixture = TestBed.createComponent(PmWorkspaceComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    dataService = TestBed.inject(DataService);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load performance data from API and map to strategies', () => {
    const mockPerf = {
      success: true,
      data: [{
        alpha_name: 'TEST-ALPHA',
        author_name: 'TEST-AUTHOR',
        total_return: 10.5,
        sharpe: 1.5,
        win_rate: 0.6,
        max_drawdown: -2.0,
        status: 'completed',
        pnl_curve: [{ cumPnL: 0 }, { cumPnL: 10.5 }]
      }]
    };

    fixture.detectChanges(); // ngOnInit

    const perfReq = httpMock.expectOne(`${APP_CONFIG.apiUrl}/analytics/performance`);
    perfReq.flush(mockPerf);

    const corrReq = httpMock.expectOne(`${APP_CONFIG.apiUrl}/analytics/correlation`);
    corrReq.flush({ success: true, data: { matrix: [[1]], labels: ['TEST-ALPHA'] } });

    const modelsReq = httpMock.expectOne(`${APP_CONFIG.apiUrl}/models`);
    modelsReq.flush({ success: true, data: [] });

    expect(component.strategies.length).toBe(1);
    expect(component.strategies[0].id).toBe('TEST-ALPHA');
    expect(component.strategies[0].totalPnL).toBe(10.5);
    expect(component.strategies[0].status).toBe('Active');
  });

  it('should handle API errors gracefully', () => {
    fixture.detectChanges();

    const perfReq = httpMock.expectOne(`${APP_CONFIG.apiUrl}/analytics/performance`);
    perfReq.error(new ErrorEvent('Network error'));

    const corrReq = httpMock.expectOne(`${APP_CONFIG.apiUrl}/analytics/correlation`);
    corrReq.flush({ success: false });

    const modelsReq = httpMock.expectOne(`${APP_CONFIG.apiUrl}/models`);
    modelsReq.flush({ success: false });

    // Should have empty strategies on error as per DataService catchError
    expect(component.strategies.length).toBe(0);
  });

  it('should filter strategies by query', () => {
    // Manually set strategies for filtering test
    component['originalStrategies'] = [
      { id: 'ALPHA-1', author: 'User A', model: 'XGB' },
      { id: 'BETA-2', author: 'User B', model: 'RF' }
    ] as any;

    component.filterStrats('ALPHA');
    expect(component.strategies.length).toBe(1);
    expect(component.strategies[0].id).toBe('ALPHA-1');

    component.filterStrats('');
    expect(component.strategies.length).toBe(2);
  });

  it('should calculate heatmap cells correctly', () => {
    component.stratLabels = ['A', 'B'];
    component.stratCorr = [[1.0, 0.5], [0.5, 1.0]];
    
    // Trigger private generateHeatmap via any cast or by setting properties
    (component as any).generateHeatmap();
    
    expect(component.heatmapCells.length).toBe(4);
    expect(component.heatmapCells[0].val).toBe(1.0);
    expect(component.heatmapCells[1].val).toBe(0.5);
  });
});
