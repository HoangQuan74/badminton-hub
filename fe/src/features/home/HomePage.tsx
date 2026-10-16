import { useMutation } from '@tanstack/react-query'
import { CalendarDays, ChevronRight, Clock3, LogOut, MapPin, Search, Star, UserRound } from 'lucide-react'
import { useNavigate } from 'react-router'
import { authApi } from '@/features/auth/auth.api'
import { tokenStorage } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'
import { queryClient } from '@/lib/query-client'

const venues = [
  { name: 'Sân Cầu Lông Thành Công', area: 'Bình Thạnh, TP.HCM', price: '120.000đ', rating: '4.9', tone: 'lime', badge: 'Còn sân tối nay' },
  { name: 'Smash Arena', area: 'Phú Nhuận, TP.HCM', price: '150.000đ', rating: '4.8', tone: 'blue', badge: 'Được yêu thích' },
  { name: 'The Shuttle House', area: 'Quận 3, TP.HCM', price: '135.000đ', rating: '4.7', tone: 'orange', badge: 'Mới trên BadmintonHub' }
]

export function HomePage() {
  const user = useAuthStore((state) => state.user)!
  const reset = useAuthStore((state) => state.reset)
  const navigate = useNavigate()
  const logout = useMutation({
    mutationFn: authApi.logout,
    onSettled: () => {
      tokenStorage.clear()
      queryClient.clear()
      reset()
      navigate('/login', { replace: true })
    }
  })
  const firstName = user.full_name.trim().split(/\s+/).at(-1) || 'bạn'

  return (
    <div className="home-page">
      <header className="site-header">
        <a className="brand" href="/"><span className="brand-mark">B</span><span>BadmintonHub</span></a>
        <nav aria-label="Điều hướng chính"><a className="active" href="#discover">Khám phá sân</a><a href="#bookings">Lịch của tôi</a><a href="#favorites">Yêu thích</a></nav>
        <div className="header-actions">
          <button className="avatar-button" type="button" aria-label="Tài khoản"><UserRound size={18} /><span>{user.full_name}</span></button>
          <button className="icon-button bordered" type="button" onClick={() => logout.mutate()} disabled={logout.isPending} aria-label="Đăng xuất"><LogOut size={18} /></button>
        </div>
      </header>

      <main>
        <section className="hero" id="discover">
          <div className="hero-copy">
            <p className="eyebrow light">Xin chào {firstName}, sẵn sàng ra sân?</p>
            <h1>Tìm sân. Chọn giờ.<br /><em>Vào trận.</em></h1>
            <p>Hàng trăm sân cầu lông đang chờ bạn. Đặt lịch rõ ràng, không cần gọi điện qua lại.</p>
          </div>
          <form className="search-panel" onSubmit={(event) => event.preventDefault()}>
            <label><span><MapPin size={16} /> Khu vực</span><input placeholder="Bạn muốn chơi ở đâu?" /></label>
            <label><span><CalendarDays size={16} /> Ngày chơi</span><input type="date" /></label>
            <label><span><Clock3 size={16} /> Khung giờ</span><select defaultValue=""><option value="" disabled>Chọn giờ</option><option>06:00 – 09:00</option><option>17:00 – 20:00</option><option>20:00 – 23:00</option></select></label>
            <button type="submit"><Search size={19} /> Tìm sân</button>
          </form>
        </section>

        <section className="content-section">
          <div className="section-heading"><div><p className="eyebrow">Gợi ý gần bạn</p><h2>Sân được cộng đồng yêu thích</h2></div><button className="text-button" type="button">Xem tất cả <ChevronRight size={17} /></button></div>
          <div className="venue-grid">
            {venues.map((venue, index) => (
              <article className="venue-card" key={venue.name}>
                <div className={`venue-visual ${venue.tone}`}><span className="court-lines" aria-hidden="true" /><span className="venue-badge">{venue.badge}</span><span className="venue-number">0{index + 1}</span></div>
                <div className="venue-info"><div className="venue-topline"><h3>{venue.name}</h3><span className="rating"><Star size={15} fill="currentColor" /> {venue.rating}</span></div><p><MapPin size={15} /> {venue.area}</p><div className="venue-footer"><span>Từ <strong>{venue.price}</strong>/giờ</span><button type="button" aria-label={`Xem ${venue.name}`}><ChevronRight size={18} /></button></div></div>
              </article>
            ))}
          </div>
        </section>

        <section className="quick-strip" id="bookings"><div><span className="quick-icon"><CalendarDays /></span><div><p className="eyebrow">Lịch sắp tới</p><h2>Bạn chưa có lịch chơi nào.</h2><p>Đặt sân đầu tiên và rủ đồng đội cùng tham gia.</p></div></div><button className="secondary-button" type="button">Khám phá sân</button></section>
      </main>
      <footer><span>© 2026 BadmintonHub</span><span>Đặt sân dễ dàng. Chơi cầu hết mình.</span></footer>
    </div>
  )
}
