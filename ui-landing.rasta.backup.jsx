const LANDING_MOCKUP_SRC =
  'https://padafand.edus.ir/emis-api/v1/idm-panel/file-get/users-drive/Umq4uj7xa297/mw7dwz8z-6nv.jpeg';

function LandingPage() {
  const { userOrgans } = getLandingAccess();

  const onEnter = (event) => {
    event?.stopPropagation?.();
    if (DC?.user?.isLogin) {
      redirectEducationalUser(userOrgans);
      return;
    }
    window.location.href = `https://${DC.service.id}/core-api/v1/data-provider/g/emis-edus--auth--goto-keestee/${DC.service.id}`;
  };

  return (
    <div dir="rtl" className="relative h-[100dvh] w-full overflow-hidden bg-[#f3ebe0]">
      <img
        src={LANDING_MOCKUP_SRC}
        alt="سامانه جامع توانمندسازی مشاوران — رستا"
        className="absolute inset-0 h-full w-full object-cover object-center select-none"
        draggable={false}
      />

      {/* Real CTA over the painted button in the mockup */}
      <button
        type="button"
        onClick={onEnter}
        aria-label="ورود به سامانه"
        className="absolute z-10 inline-flex items-center justify-center gap-2 rounded-full border-0 bg-[#3f5138] px-[clamp(1.1rem,2.2vw,1.85rem)] font-bold text-white shadow-none transition hover:bg-[#364530] focus-visible:outline focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-[#c98a3b] active:scale-[0.99]"
        style={{
          top: '58%',
          right: '10%',
          width: 'min(22rem, 28%)',
          height: '6.5%',
          minHeight: '2.75rem',
          fontSize: 'clamp(0.85rem, 1.35vw, 1.125rem)',
        }}
      >
        <span>ورود به سامانه</span>
        <span aria-hidden="true" className="text-[1.15em] leading-none">
          ←
        </span>
      </button>
    </div>
  );
}
