import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotificationService, Notification } from '../../services/notification.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-toast',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './toast.component.html',
  styleUrl: './toast.component.scss'
})
export class ToastComponent implements OnInit, OnDestroy {
  messages: Notification[] = [];
  private sub?: Subscription;

  constructor(private ns: NotificationService) {}

  ngOnInit() {
    this.sub = this.ns.notifications$.subscribe(n => {
      this.messages.push(n);
      setTimeout(() => this.remove(0), n.duration || 3000);
    });
  }

  ngOnDestroy() {
    this.sub?.unsubscribe();
  }

  remove(idx: number) {
    this.messages.splice(idx, 1);
  }
}
