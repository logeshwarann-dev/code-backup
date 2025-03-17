import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { IntervalData } from '../../models/partition.model'; // Import the IntervalData model

@Component({
  selector: 'app-topstocks',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './topstocks.component.html',
  styleUrls: ['./topstocks.component.css']
})
export class TopStocksComponent {
  @Input() intervalData: IntervalData | null = null; // Receive interval data from the parent

  topStocks = [
    { name: 'AAPL', price: 150, change: 1.2 },
    { name: 'GOOGL', price: 2800, change: -0.5 },
    { name: 'AMZN', price: 3400, change: 0.8 },
    { name: 'TSLA', price: 900, change: 2.1 },
    { name: 'MSFT', price: 299, change: -1.3 }
  ];

  // This can be extended to process intervalData if needed
}
