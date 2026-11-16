/* Temporary mockup landing — restore the commented UI when design is ready.

const COLORS = {
  forest: 'var(--rasta-forest)',
  leaf: 'var(--rasta-leaf)',
  sand: 'var(--rasta-sand)',
  orange: 'var(--rasta-orange)',
};

function RastaLeafLogo({ className = 'h-9 w-9' }) {
  return (
    <svg className={className} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M24 4C14 10 8 20 10 34c6-2 12-2 18 1 4-10 4-21-4-31Z" fill={COLORS.leaf} />
      <path d="M24 8c-6 8-8 16-7 26 4-1.5 8-1 12 1 2.5-8 2-18-5-27Z" fill={COLORS.forest} opacity="0.55" />
      <path d="M22 12c1 8 2 16 5 24" stroke="#fff" strokeWidth="1.6" strokeLinecap="round" opacity="0.7" />
      <path d="M22 20c4 2 7 5 9 9M20 28c3 1.5 5.5 3 8 6" stroke="#fff" strokeWidth="1.2" strokeLinecap="round" opacity="0.55" />
    </svg>
  );
}

function OrangeAccent({ className = '' }) {
  return (
    <svg className={className} viewBox="0 0 40 48" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M20 2C10 14 6 26 12 44c6-4 12-6 18-4C28 24 28 12 20 2Z" fill={COLORS.orange} />
      <path d="M20 10c1 8 2 16 3 26" stroke="#fff" strokeWidth="1.4" strokeLinecap="round" opacity="0.55" />
    </svg>
  );
}

function BackgroundArcs() {
  return (
    <svg className="pointer-events-none absolute inset-0 h-full w-full" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <circle cx="980" cy="120" r="520" stroke={COLORS.sand} strokeWidth="1.25" opacity="0.85" />
      <circle cx="980" cy="120" r="440" stroke={COLORS.sand} strokeWidth="1" opacity="0.55" />
      <circle cx="980" cy="120" r="360" stroke={COLORS.sand} strokeWidth="0.9" opacity="0.4" />
    </svg>
  );
}

function OrganicShapes() {
  return (
    <svg className="pointer-events-none absolute inset-0 h-full w-full" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M-120 40C40 -40 260 -20 380 120C500 260 480 480 340 620C200 760 -40 780 -120 620C-180 500 -200 200 -120 40Z" fill={COLORS.leaf} />
      <path d="M-100 80C40 10 230 30 330 150C430 270 410 460 290 580C170 700 -20 710 -90 570C-140 470 -160 200 -100 80Z" fill={COLORS.forest} opacity="0.38" />
      <path d="M-80 560C60 500 200 560 280 680C360 800 220 900 40 920C-60 930 -120 760 -80 560Z" fill={COLORS.leaf} opacity="0.95" />
      <path d="M-50 600C70 560 180 610 240 700C300 790 180 870 50 880C-30 885 -80 720 -50 600Z" fill={COLORS.forest} opacity="0.32" />
      <path d="M60 200C130 140 220 150 280 230C340 310 300 420 210 480" stroke="#fff" strokeWidth="1.5" fill="none" opacity="0.4" />
      <path d="M30 280C100 240 170 255 220 320" stroke="#fff" strokeWidth="1.15" fill="none" opacity="0.32" />
      <path d="M40 620C110 590 170 620 210 680" stroke="#fff" strokeWidth="1.2" fill="none" opacity="0.28" />
    </svg>
  );
}

function LandingPage_UI() {
  const { userOrgans, canEnter } = getLandingAccess();

  const onEnter = () => {
    if (DC?.user?.isLogin) {
      redirectEducationalUser(userOrgans);
      return;
    }
    window.location.href = `https://${DC.service.id}/core-api/v1/data-provider/g/emis-edus--auth--goto-keestee/${DC.service.id}`;
  };

  const heroImageSrc = 'https://padafand.edus.ir/emis-api/v1/idm-panel/file-get/users-drive/Umq4uj7xa297/mw90cu5c-6is.jpg';
  const entering = false;

  return (
    <div
      dir="rtl"
      className="relative min-h-[100dvh] w-full overflow-hidden bg-[color-mix(in_srgb,var(--rasta-sand)_42%,#ffffff)] text-[color:var(--rasta-forest)]"
    >
      <BackgroundArcs />
      <OrganicShapes />

      <OrangeAccent className="pointer-events-none absolute left-[46%] top-[7%] z-10 h-10 w-8 -translate-x-1/2 md:left-[42%] md:h-12 md:w-10" />

      <header className="relative z-20 flex items-center justify-between px-5 pt-5 sm:px-8 sm:pt-7 lg:px-12">
        <div className="flex items-center gap-2.5">
          <RastaLeafLogo className="h-9 w-9 sm:h-10 sm:w-10" />
          <span
            className="text-[1.65rem] font-black leading-none tracking-tight sm:text-[1.85rem]"
            style={{ color: COLORS.forest }}
          >
            رستا
          </span>
        </div>
        <p className="max-w-[9.5rem] text-left text-[0.7rem] font-medium leading-snug text-[color:color-mix(in_srgb,var(--rasta-forest)_70%,#ffffff)] sm:max-w-none sm:text-sm">
          معاونت پرورشی آموزش و پرورش
        </p>
      </header>

      <main className="relative z-20 mx-auto flex min-h-[calc(100dvh-5.5rem)] w-full max-w-6xl flex-col-reverse items-center gap-8 px-5 pb-10 pt-6 sm:px-8 lg:flex-row lg:items-center lg:justify-between lg:gap-6 lg:px-12 lg:pb-16 lg:pt-4">
        <section className="w-full max-w-xl shrink-0 text-right lg:w-[48%]">
          <p className="mb-3 text-sm font-semibold text-[color:color-mix(in_srgb,var(--rasta-leaf)_75%,var(--rasta-orange))] sm:text-base">
            ویژه مشاوران مدارس سراسر کشور
          </p>

          <h1 className="mb-4 text-[1.85rem] font-black leading-[1.35] text-[color:var(--rasta-forest)] sm:text-4xl lg:text-[2.65rem] lg:leading-[1.3]">
            سامانه جامع توانمندسازی مشاوران
          </h1>

          <p className="mb-8 max-w-md text-sm leading-7 text-[color:color-mix(in_srgb,var(--rasta-forest)_55%,#ffffff)] sm:text-base sm:leading-8">
            ارتقای شایستگی‌های حرفه‌ای و شبکه‌سازی مشاوران
          </p>

          <button
            type="button"
            onClick={onEnter}
            disabled={entering}
            className="inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-[color:var(--rasta-forest)] px-8 py-3 text-base font-bold text-white shadow-[0_10px_28px_rgba(var(--rasta-forest-rgb),0.28)] transition hover:-translate-y-0.5 hover:bg-[color:color-mix(in_srgb,var(--rasta-forest)_88%,#000000)] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-[color:var(--rasta-orange)] disabled:cursor-wait disabled:opacity-75 disabled:hover:translate-y-0"
          >
            {entering ? (
              <span
                className="h-4 w-4 animate-spin rounded-full border-2 border-white/35 border-t-white"
                aria-hidden="true"
              />
            ) : null}
            <span>ورود به سامانه</span>
            <span aria-hidden="true" className="text-lg leading-none">←</span>
          </button>
        </section>

        <section className="relative flex w-full max-w-[22rem] shrink-0 justify-center sm:max-w-md lg:max-w-[30rem] lg:flex-1">
          <div
            className="relative z-10 aspect-[5/6] w-[90%] overflow-hidden bg-[color:var(--rasta-sand)] shadow-[0_22px_55px_rgba(var(--rasta-forest-rgb),0.2)]"
            style={{ borderRadius: '56% 44% 50% 50% / 46% 40% 60% 54%' }}
          >
            <img
              src={heroImageSrc}
              alt="مشاوره و گفت‌وگوی حرفه‌ای در مدرسه"
              className="h-full w-full scale-[1.08] object-cover object-[center_18%]"
              draggable={false}
            />
          </div>

          <div className="pointer-events-none absolute inset-y-[4%] -start-2 z-20 w-[42%] sm:-start-4" aria-hidden="true">
            <svg className="h-full w-full" viewBox="0 0 180 420" preserveAspectRatio="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M160 20C90 60 30 140 28 220C26 300 70 370 150 400C80 360 50 290 52 220C54 140 100 70 160 20Z" fill={COLORS.leaf} />
              <path d="M148 40C95 75 48 145 46 220C44 290 75 350 140 380" fill="none" stroke="#fff" strokeWidth="1.4" opacity="0.35" />
            </svg>
          </div>
        </section>
      </main>
    </div>
  );
}
*/
