import type { ReactNode } from 'react'
import { Link } from 'react-router'
import { CalendarCheck, MapPin, ShieldCheck } from 'lucide-react'

export function AuthShell({ children }: { children: ReactNode }) {
  return (
    <main className="auth-shell">
      <section className="auth-story" aria-label="Giới thiệu BadmintonHub">
        <Link className="brand brand-on-dark" to="/">
          <span className="brand-mark">B</span>
          <span>BadmintonHub</span>
        </Link>
        <div className="auth-story-content">
          <p className="eyebrow light">Ra sân đúng giờ. Chơi hết mình.</p>
          <h1>Sân cầu lông phù hợp, chỉ vài thao tác.</h1>
          <p>Tìm sân gần bạn, xem giờ trống theo thời gian thực và quản lý mọi lịch chơi tại một nơi.</p>
          <ul className="auth-benefits">
            <li><MapPin size={20} /> Khám phá sân theo khu vực</li>
            <li><CalendarCheck size={20} /> Đặt lịch rõ ràng, nhanh chóng</li>
            <li><ShieldCheck size={20} /> Thông tin đặt sân được bảo vệ</li>
          </ul>
        </div>
        <p className="auth-quote">“Một trận cầu hay bắt đầu từ một lịch hẹn đơn giản.”</p>
      </section>
      <section className="auth-panel">{children}</section>
    </main>
  )
}
