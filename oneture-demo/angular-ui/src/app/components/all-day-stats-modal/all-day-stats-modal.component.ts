import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatTableModule } from '@angular/material/table';

@Component({
  selector: 'app-all-day-stats-modal',
  templateUrl: './all-day-stats-modal.component.html',
  standalone: true,
  imports: [MatTableModule],
  styleUrls: ['./all-day-stats-modal.component.css'],
})
export class AllDayStatsModalComponent {
  constructor(@Inject(MAT_DIALOG_DATA) public data: any) {}

  displayedColumns: string[] = ['key', 'value'];

  // Transform data into a flat array for table display
  getTableData(): any[] {
    if (!this.data) return [];
    const tableData: any[] = [];
    for (const [key, value] of Object.entries(this.data)) {
      if (typeof value === 'object' && !Array.isArray(value)) {
        tableData.push({ key, value: JSON.stringify(value) });
      } else {
        tableData.push({ key, value });
      }
    }
    return tableData;
  }
}
