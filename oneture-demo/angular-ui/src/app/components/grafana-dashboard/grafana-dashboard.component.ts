import { Component } from '@angular/core';
import { DomSanitizer, SafeResourceUrl } from '@angular/platform-browser';
import { environment } from '../../../environment';

@Component({
  selector: 'app-grafana-dashboard',
  standalone: true,
  templateUrl: './grafana-dashboard.component.html',
  styleUrls: ['./grafana-dashboard.component.css']
})
export class GrafanaDashboardComponent {
  grafanaUrl: SafeResourceUrl;

  constructor(private sanitizer: DomSanitizer) {
    this.grafanaUrl = this.sanitizer.bypassSecurityTrustResourceUrl(
      environment.grafana_endpoint+'/dashboards'
    );
  }
}
