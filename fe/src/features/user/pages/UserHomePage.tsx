import { useMutation } from '@tanstack/react-query'
import { CalendarDays, ChevronRight, Clock3, LogOut, MapPin, Search, Star, UserRound } from 'lucide-react'
import { useNavigate } from 'react-router'
import { authApi } from '@/features/auth/auth.api'
import { tokenStorage } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'
import { queryClient } from '@/lib/query-client'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useI18n } from '@/lib/use-i18n'

const venues = [
  { name: 'Sân Cầu Lông Thành Công', area: 'Bình Thạnh, TP.HCM', price: '120.000đ', rating: '4.9', tone: 'lime', badgeKey: 'user.venueAvailability' as const },
  { name: 'Smash Arena', area: 'Phú Nhuận, TP.HCM', price: '150.000đ', rating: '4.8', tone: 'blue', badgeKey: 'user.venueFavorite' as const },
  { name: 'The Shuttle House', area: 'Quận 3, TP.HCM', price: '135.000đ', rating: '4.7', tone: 'orange', badgeKey: 'user.venueNew' as const }
]

export function UserHomePage() {
  const { t } = useI18n()
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
        <nav aria-label={t('user.nav.discover')}><a className="active" href="#discover">{t('user.nav.discover')}</a><a href="#bookings">{t('user.nav.bookings')}</a><a href="#favorites">{t('user.nav.favorites')}</a></nav>
        <div className="header-actions">
          <LanguageSwitcher />
          <button className="avatar-button" type="button" aria-label={t('user.account')}><UserRound size={18} /><span>{user.full_name}</span></button>
          <button className="icon-button bordered" type="button" onClick={() => logout.mutate()} disabled={logout.isPending} aria-label={t('user.logout')}><LogOut size={18} /></button>
        </div>
      </header>

      <main>
        <section className="hero" id="discover">
          <div className="hero-copy">
            <p className="eyebrow light">{t('user.greeting', { name: firstName })}</p>
            <h1>{t('user.heroTitle')}<br /><em>{t('user.heroEmphasis')}</em></h1>
            <p>{t('user.heroDescription')}</p>
          </div>
          <form className="search-panel" onSubmit={(event) => event.preventDefault()}>
            <label><span><MapPin size={16} /> {t('user.area')}</span><input placeholder={t('user.areaPlaceholder')} /></label>
            <label><span><CalendarDays size={16} /> {t('user.playDate')}</span><input type="date" /></label>
            <label><span><Clock3 size={16} /> {t('user.timeSlot')}</span><select defaultValue=""><option value="" disabled>{t('user.timePlaceholder')}</option><option>{t('user.morning')}</option><option>{t('user.evening')}</option><option>{t('user.night')}</option></select></label>
            <button type="submit"><Search size={19} /> {t('user.search')}</button>
          </form>
        </section>

        <section className="content-section">
          <div className="section-heading"><div><p className="eyebrow">{t('user.nearby')}</p><h2>{t('user.popularVenues')}</h2></div><button className="text-button" type="button">{t('user.seeAll')} <ChevronRight size={17} /></button></div>
          <div className="venue-grid">
            {venues.map((venue, index) => (
              <article className="venue-card" key={venue.name}>
                <div className={`venue-visual ${venue.tone}`}><span className="court-lines" aria-hidden="true" /><span className="venue-badge">{t(venue.badgeKey)}</span><span className="venue-number">0{index + 1}</span></div>
                <div className="venue-info"><div className="venue-topline"><h3>{venue.name}</h3><span className="rating"><Star size={15} fill="currentColor" /> {venue.rating}</span></div><p><MapPin size={15} /> {venue.area}</p><div className="venue-footer"><span>{t('user.from')} <strong>{venue.price}</strong>{t('user.perHour')}</span><button type="button" aria-label={t('user.viewVenue', { name: venue.name })}><ChevronRight size={18} /></button></div></div>
              </article>
            ))}
          </div>
        </section>

        <section className="quick-strip" id="bookings"><div><span className="quick-icon"><CalendarDays /></span><div><p className="eyebrow">{t('user.upcoming')}</p><h2>{t('user.noBookings')}</h2><p>{t('user.noBookingsDescription')}</p></div></div><button className="secondary-button" type="button">{t('user.exploreVenues')}</button></section>
      </main>
      <footer><span>© 2026 BadmintonHub</span><span>{t('user.footer')}</span></footer>
    </div>
  )
}
