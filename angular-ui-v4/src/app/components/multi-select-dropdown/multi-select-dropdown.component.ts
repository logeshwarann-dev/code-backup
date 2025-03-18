import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-multi-select-dropdown',
  templateUrl: './multi-select-dropdown.component.html',
  styleUrls: ['./multi-select-dropdown.component.css'],
  standalone: true,
  imports: [CommonModule, FormsModule]
})
export class MultiSelectDropdownComponent {
  @Input() list: any[] = []; 
  @Output() shareCheckedList = new EventEmitter();
  @Output() shareIndividualCheckedList = new EventEmitter();

  checkedList: any[] = [];
  appliedCheckedList: any[] = []; // Stores the final applied selection
  selectAllChecked: boolean = true; // Default to "All Partitions" selected
  showDropDown: boolean = false;

  /** Handles toggling 'All Partitions' checkbox **/
  toggleAllPartitions(status: boolean) {
    this.checkedList = [];
    this.list.forEach(item => {
      item.checked = status;
      if (status) {
        this.checkedList.push(item.name);
      }
    });

    if (status) {
      this.appliedCheckedList = ["All Partitions"];
    } else {
      this.appliedCheckedList = [];
    }
  }

  /** Updates individual partition selection **/
  updatePartitionSelection() {
    this.selectAllChecked = this.list.every(item => item.checked); // Auto-select "All" if all items are checked

    this.checkedList = this.list
      .filter(item => item.checked)
      .map(item => item.name);

    if (this.selectAllChecked) {
      this.appliedCheckedList = ["All Partitions"];
    } else {
      this.appliedCheckedList = this.checkedList.map(name =>
        name.replace('Partition ', '')
      );
    }
  }

  /** Applies the selection when 'Apply' is clicked **/
  applySelection() {
    if (this.selectAllChecked) {
      this.appliedCheckedList = ["All Partitions"];
    } else {
      this.appliedCheckedList = this.checkedList.map(name =>
        name.replace('Partition ', '')
      );
    }

    // Emit the final applied list
    this.shareCheckedList.emit(this.appliedCheckedList);

    // Close the dropdown
    this.showDropDown = false;
  }
}
