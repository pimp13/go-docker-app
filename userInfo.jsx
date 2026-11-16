const TABS = [
  { id: 'personal', label: 'اطلاعات فردی', icon: 'person' },
  { id: 'school', label: 'مدرسه و سازمان محل خدمت', icon: 'school' },
  { id: 'job', label: 'اطلاعات شغلی و قرارداد', icon: 'work' },
  { id: 'extra', label: 'اطلاعات تکمیلی', icon: 'assignment_ind' },
];

function getPath(obj, path) {
  if (!obj || !path) return undefined;
  return path.split('.').reduce((cur, key) => (cur == null ? undefined : cur[key]), obj);
}

function toFormValue(value) {
  if (value == null || value === '—' || value === '— —') return '';
  if (Array.isArray(value)) return value.filter(Boolean).join(' / ');
  return String(value).trim();
}

function emptyToNull(value) {
  if (value == null) return null;
  if (typeof value === 'string' && value.trim() === '') return null;
  return value;
}

function toNullableNumber(value) {
  if (value == null || value === '') return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

function pickInfoValue(items, label) {
  if (!items) return '';
  const found = items.find((x) => x.label === label);
  if (!found) return '';
  return toFormValue(found.value);
}

/** اول از userData، در صورت خالی بودن از مدل نمایشی info */
function fromUserOrInfo(userData, info, paths = {}) {
  if (paths.user) {
    const raw = getPath(userData, paths.user);
    const value = toFormValue(raw);
    if (value !== '') return value;
  }
  if (paths.userAlt) {
    const raw = getPath(userData, paths.userAlt);
    const value = toFormValue(raw);
    if (value !== '') return value;
  }
  if (paths.infoSection && paths.infoLabel) {
    return pickInfoValue(info?.[paths.infoSection], paths.infoLabel);
  }
  return '';
}

const FIELD_DEFS = {
  personal: [
    {
      key: 'firstName',
      label: 'نام',
      icon: 'badge',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'person.firstName', infoSection: 'personal', infoLabel: 'نام' }),
    },
    {
      key: 'lastName',
      label: 'نام خانوادگی',
      icon: 'badge',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'person.lastName', infoSection: 'personal', infoLabel: 'نام خانوادگی' }),
    },
    {
      key: 'nationalCode',
      label: 'کد ملی',
      icon: 'credit_card',
      type: 'text',
      readOnly: true,
      from: (u, i) => fromUserOrInfo(u, i, { user: 'person.nationalCode', infoSection: 'personal', infoLabel: 'کد ملی' }),
    },
    {
      key: 'fatherName',
      label: 'نام پدر',
      icon: 'family_restroom',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'person.fatherName', infoSection: 'personal', infoLabel: 'نام پدر' }),
    },
    {
      key: 'birthDate',
      label: 'تاریخ تولد',
      icon: 'cake',
      type: 'date',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'person.birthDate', infoSection: 'personal', infoLabel: 'تاریخ تولد' }),
    },
    {
      key: 'gender',
      label: 'جنسیت',
      icon: 'wc',
      type: 'select',
      options: [
        { value: 'female', label: 'زن' },
        { value: 'male', label: 'مرد' },
      ],
      from: (u, i) => fromUserOrInfo(u, i, { user: 'person.gender', infoSection: 'personal', infoLabel: 'جنسیت' }),
    },
  ],

  job: [
    {
      key: 'personnelCode',
      label: 'کد پرسنلی',
      icon: 'badge',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.personnelCode', infoSection: 'job', infoLabel: 'کد پرسنلی' }),
    },
    {
      key: 'positionTitle',
      label: 'عنوان موقعیت شغلی',
      icon: 'business_center',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.positionTitle', infoSection: 'job', infoLabel: 'عنوان موقعیت شغلی' }),
    },
    {
      key: 'positionCode',
      label: 'کد موقعیت شغلی',
      icon: 'pin',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.positionCode', infoSection: 'job', infoLabel: 'کد موقعیت شغلی' }),
    },
    {
      key: 'jobFieldTitle',
      label: 'عنوان رشته شغلی',
      icon: 'category',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.jobFieldTitle', infoSection: 'job', infoLabel: 'عنوان رشته شغلی' }),
    },
    {
      key: 'jobFieldCode',
      label: 'کد رشته شغلی',
      icon: 'code',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.jobFieldCode', infoSection: 'job', infoLabel: 'کد رشته شغلی' }),
    },
    {
      key: 'jobCategory',
      label: 'دسته‌بندی شغلی',
      icon: 'school',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.jobCategory', infoSection: 'job', infoLabel: 'دسته‌بندی شغلی' }),
    },
    {
      key: 'contractType',
      label: 'نوع قرارداد',
      icon: 'assignment',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.contractType', infoSection: 'job', infoLabel: 'نوع قرارداد' }),
    },
    {
      key: 'rasteh',
      label: 'رسته',
      icon: 'groups',
      type: 'select',
      options: [
        { value: 'amoozegari', label: 'آموزگاری' },
        { value: 'amoozeshi_farhangi', label: 'آموزشی و فرهنگی' },
        { value: 'dabiri', label: 'دبیری' },
        { value: 'other', label: 'سایر' },
      ],
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.rasteh', infoSection: 'job', infoLabel: 'رسته' }),
    },
    {
      key: 'workingHours',
      label: 'ساعت کاری موظفی',
      icon: 'schedule',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.workingHours', infoSection: 'job', infoLabel: 'ساعت کاری موظفی' }),
    },
    {
      key: 'salaryCondition',
      label: 'شرایط حقوقی',
      icon: 'payments',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.salaryCondition', infoSection: 'job', infoLabel: 'شرایط حقوقی' }),
    },
    {
      key: 'serviceYears',
      label: 'سابقه خدمت به سال',
      icon: 'history',
      type: 'number',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'employment.serviceYears', infoSection: 'job', infoLabel: 'سابقه خدمت' }),
    },
    {
      key: 'counselingServiceYears',
      label: 'سابقه خدمت در حوزه مشاوره به سال',
      icon: 'psychology',
      type: 'number',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'employment.counselingServiceYears',
          infoSection: 'job',
          infoLabel: 'سابقه مشاوره',
        }),
    },
  ],

  school: [
    {
      key: 'schoolName',
      label: 'نام مدرسه',
      icon: 'domain',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'organization.schoolName', infoSection: 'school', infoLabel: 'نام مدرسه' }),
    },
    {
      key: 'schoolCode',
      label: 'کد مدرسه',
      icon: 'numbers',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'organization.schoolCode', infoSection: 'school', infoLabel: 'کد مدرسه' }),
    },
    {
      key: 'organizationName',
      label: 'نام سازمان',
      icon: 'apartment',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.organizationName',
          userAlt: 'organization.organName',
          infoSection: 'school',
          infoLabel: 'نام سازمان',
        }),
    },
    {
      key: 'organizationCode',
      label: 'کد سازمان',
      icon: 'corporate_fare',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.organizationCode',
          userAlt: 'organization.organLocalId',
          infoSection: 'school',
          infoLabel: 'کد سازمان',
        }),
    },
    {
      key: 'organizationType',
      label: 'نوع سازمان',
      icon: 'business',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.organizationType',
          infoSection: 'school',
          infoLabel: 'نوع سازمان',
        }),
    },
    {
      key: 'schoolType',
      label: 'نوع مدرسه',
      icon: 'home_work',
      type: 'text',
      from: (u, i) => fromUserOrInfo(u, i, { user: 'organization.schoolType', infoSection: 'school', infoLabel: 'نوع مدرسه' }),
    },
    {
      key: 'schoolGender',
      label: 'جنسیت مدرسه',
      icon: 'wc',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.schoolGender',
          infoSection: 'school',
          infoLabel: 'جنسیت مدرسه',
        }),
    },
    {
      key: 'educationalPeriod',
      label: 'مقطع تحصیلی',
      icon: 'menu_book',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.educationalPeriod',
          infoSection: 'school',
          infoLabel: 'مقطع تحصیلی',
        }),
    },
    {
      key: 'activityRegionCode',
      label: 'کد منطقه فعالیت',
      icon: 'code',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.activityRegionCode',
          infoSection: 'school',
          infoLabel: 'کد منطقه',
        }),
    },
    {
      key: 'organizationalLevel',
      label: 'سطح سازمانی',
      icon: 'account_tree',
      type: 'text',
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'organization.organizationalLevel',
          infoSection: 'school',
          infoLabel: 'سطح سازمانی',
        }),
    },
  ],

  /** فرم اولیه تکمیل اطلاعات مشاوران */
  extra: [
    {
      key: 'maritalStatus',
      label: 'وضعیت تأهل',
      icon: 'favorite',
      type: 'select',
      options: [
        { value: 'married', label: 'متأهل' },
        { value: 'single', label: 'مجرد' },
      ],
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'additional.maritalStatus',
          infoSection: 'extra',
          infoLabel: 'وضعیت تأهل',
        }),
    },
    {
      key: 'childrenCount',
      label: 'تعداد فرزندان (در صورت تأهل)',
      icon: 'child_care',
      type: 'select',
      options: [
        { value: '1', label: 'یک' },
        { value: '2', label: 'دو' },
        { value: '3', label: 'سه' },
        { value: '4_plus', label: 'چهار و بیشتر' },
      ],
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'additional.childrenCount',
          infoSection: 'extra',
          infoLabel: 'تعداد فرزندان',
        }),
    },
    {
      key: 'lastEducationDegree',
      label: 'آخرین مدرک تحصیلی',
      icon: 'school',
      type: 'select',
      options: [
        { value: 'bachelor', label: 'کارشناسی' },
        { value: 'master', label: 'فوق کارشناسی' },
        { value: 'phd_student', label: 'دانشجوی دکتری' },
        { value: 'phd', label: 'دکتری' },
      ],
      from: (u, i) =>
        fromUserOrInfo(u, i, {
          user: 'additional.lastEducationDegree',
          infoSection: 'extra',
          infoLabel: 'آخرین مدرک تحصیلی',
        }),
    },
    {
      key: 'bachelorDegree',
      label: 'مدرک تحصیلی کارشناسی',
      icon: 'menu_book',
      type: 'text',
      from: (u) => toFormValue(u?.additional?.bachelorDegree),
    },
    {
      key: 'bachelorUniversity',
      label: 'دانشگاه اخذ مدرک کارشناسی',
      icon: 'account_balance',
      type: 'text',
      from: (u) => toFormValue(u?.additional?.bachelorUniversity),
    },
    {
      key: 'masterDegree',
      label: 'مدرک تحصیلی فوق کارشناسی',
      icon: 'menu_book',
      type: 'text',
      from: (u) => toFormValue(u?.additional?.masterDegree),
    },
    {
      key: 'masterUniversity',
      label: 'دانشگاه اخذ مدرک فوق کارشناسی',
      icon: 'account_balance',
      type: 'text',
      from: (u) => toFormValue(u?.additional?.masterUniversity),
    },
    {
      key: 'doctorateDegree',
      label: 'مدرک تحصیلی دکتری',
      icon: 'menu_book',
      type: 'text',
      from: (u) => toFormValue(u?.additional?.doctorateDegree),
    },
    {
      key: 'doctorateUniversity',
      label: 'دانشگاه اخذ مدرک / درحال تحصیل دکتری',
      icon: 'account_balance',
      type: 'text',
      from: (u) => toFormValue(u?.additional?.doctorateUniversity),
    },
    {
      key: 'doctorateInProgress',
      label: 'در حال تحصیل دکتری',
      icon: 'pending_actions',
      type: 'select',
      options: [
        { value: 'true', label: 'بله' },
        { value: 'false', label: 'خیر' },
      ],
      from: (u) => {
        const v = u?.additional?.doctorateInProgress;
        if (v == null || v === '') return '';
        return String(Boolean(v));
      },
    },
    {
      key: 'willingToAttendRastehCourse',
      label:
        'مایل به حضور در دوره آموزشی حضوری دو روزه (ارتقای شایستگی‌های حرفه‌ای و شبکه‌سازی مشاوران: رستا) در مرکز استان هستید؟',
      icon: 'event',
      type: 'select',
      options: [
        { value: 'yes', label: 'بله' },
        { value: 'no', label: 'خیر' },
        { value: 'undecided', label: 'هنوز تصمیم نگرفته‌ام' },
      ],
      fullWidth: true,
      from: (u) => toFormValue(u?.additional?.willingToAttendRastehCourse),
    },
  ],
};

function buildFormFromData(userData, info) {
  const next = { personal: {}, job: {}, school: {}, extra: {} };
  Object.keys(FIELD_DEFS).forEach((section) => {
    FIELD_DEFS[section].forEach((field) => {
      next[section][field.key] = field.from(userData, info) ?? '';
    });
  });
  return next;
}


/** payload هم‌شکل با userData برای ارسال به بک‌اند */
function buildSavePayload(form, userData) {
  const firstName = emptyToNull(form.personal.firstName);
  const lastName = emptyToNull(form.personal.lastName);
  const fullName =
    [firstName, lastName].filter(Boolean).join(' ') ||
    emptyToNull(userData?.person?.fullName);

  const emp = userData?.employment || {};
  const org = userData?.organization || {};

  return {
    person: {
      randId: userData?.person?.randId ?? null,
      firstName,
      lastName,
      fullName,
      nationalCode: emptyToNull(form.personal.nationalCode) ?? userData?.person?.nationalCode ?? null,
      fatherName: emptyToNull(form.personal.fatherName),
      birthDate: emptyToNull(form.personal.birthDate),
      gender: emptyToNull(form.personal.gender),
    },
    employment: {
      personnelCode: emptyToNull(form.job.personnelCode),
      positionTitle: emptyToNull(form.job.positionTitle),
      positionCode: emptyToNull(form.job.positionCode),
      jobFieldTitle: emptyToNull(form.job.jobFieldTitle),
      jobFieldCode: emptyToNull(form.job.jobFieldCode),
      jobFieldId: emp.jobFieldId ?? null,
      isEducationalJobField: emp.isEducationalJobField ?? null,
      jobCategory: emptyToNull(form.job.jobCategory),
      jobCategoryCode: emp.jobCategoryCode ?? null,
      contractType: emptyToNull(form.job.contractType),
      rasteh: emptyToNull(form.job.rasteh),
      workingHours: emptyToNull(form.job.workingHours),
      teachingHourType: emp.teachingHourType ?? null,
      teachingHourTypeCode: emp.teachingHourTypeCode ?? null,
      salaryCondition: emptyToNull(form.job.salaryCondition),
      serviceYears: toNullableNumber(form.job.serviceYears),
      counselingServiceYears: toNullableNumber(form.job.counselingServiceYears),
    },
    organization: {
      schoolName: emptyToNull(form.school.schoolName),
      schoolCode: emptyToNull(form.school.schoolCode),
      organizationName: emptyToNull(form.school.organizationName),
      organizationCode: emptyToNull(form.school.organizationCode),
      organizationType: emptyToNull(form.school.organizationType),
      organizationTypeCode: org.organizationTypeCode ?? null,
      schoolType: emptyToNull(form.school.schoolType),
      schoolGender: emptyToNull(form.school.schoolGender),
      schoolGenderCode: org.schoolGenderCode ?? null,
      educationalPeriod: emptyToNull(form.school.educationalPeriod),
      educationalPeriodCode: org.educationalPeriodCode ?? null,
      activityRegionCode: emptyToNull(form.school.activityRegionCode),
      organizationalLevel: emptyToNull(form.school.organizationalLevel),
      organRandId: org.organRandId ?? null,
      organPath: org.organPath ?? null,
      organLocalId: org.organLocalId ?? emptyToNull(form.school.organizationCode),
      organName: org.organName ?? emptyToNull(form.school.organizationName),
      profileOrganPath: org.profileOrganPath ?? null,
      profileOrganRandId: org.profileOrganRandId ?? null,
    },
    additional: {
      maritalStatus: emptyToNull(form.extra.maritalStatus),
      childrenCount:
        form.extra.maritalStatus === 'married'
          ? emptyToNull(form.extra.childrenCount)
          : null,
      lastEducationDegree: emptyToNull(form.extra.lastEducationDegree),
      bachelorDegree: emptyToNull(form.extra.bachelorDegree),
      bachelorUniversity: emptyToNull(form.extra.bachelorUniversity),
      masterDegree: emptyToNull(form.extra.masterDegree),
      masterUniversity: emptyToNull(form.extra.masterUniversity),
      doctorateDegree: emptyToNull(form.extra.doctorateDegree),
      doctorateUniversity: emptyToNull(form.extra.doctorateUniversity),
      doctorateInProgress:
        form.extra.doctorateInProgress === ''
          ? null
          : form.extra.doctorateInProgress === 'true',
      willingToAttendRastehCourse: emptyToNull(form.extra.willingToAttendRastehCourse),
    },
  };
}

const JALALI_MONTHS = [
  'فروردین',
  'اردیبهشت',
  'خرداد',
  'تیر',
  'مرداد',
  'شهریور',
  'مهر',
  'آبان',
  'آذر',
  'دی',
  'بهمن',
  'اسفند',
];

const JALALI_WEEKDAYS = ['ش', 'ی', 'د', 'س', 'چ', 'پ', 'ج'];

function div(a, b) {
  return Math.floor(a / b);
}

function gregorianToJalali(gy, gm, gd) {
  const gdm = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334];
  let gy2 = gm > 2 ? gy + 1 : gy;
  let days =
    355666 +
    365 * gy +
    div(gy2 + 3, 4) -
    div(gy2 + 99, 100) +
    div(gy2 + 399, 400) +
    gd +
    gdm[gm - 1];
  let jy = -1595 + 33 * div(days, 12053);
  days %= 12053;
  jy += 4 * div(days, 1461);
  days %= 1461;
  if (days > 365) {
    jy += div(days - 1, 365);
    days = (days - 1) % 365;
  }
  const jm = days < 186 ? 1 + div(days, 31) : 7 + div(days - 186, 30);
  const jd = 1 + (days < 186 ? days % 31 : (days - 186) % 30);
  return { jy, jm, jd };
}

function jalaliToGregorian(jy, jm, jd) {
  jy += 1595;
  let days =
    -355668 +
    365 * jy +
    div(jy, 33) * 8 +
    div((jy % 33) + 3, 4) +
    jd +
    (jm < 7 ? (jm - 1) * 31 : (jm - 7) * 30 + 186);
  let gy = 400 * div(days, 146097);
  days %= 146097;
  if (days > 36524) {
    gy += 100 * div(--days, 36524);
    days %= 36524;
    if (days >= 365) days += 1;
  }
  gy += 4 * div(days, 1461);
  days %= 1461;
  if (days > 365) {
    gy += div(days - 1, 365);
    days = (days - 1) % 365;
  }
  let gd = days + 1;
  const salA = [
    0,
    31,
    (gy % 4 === 0 && gy % 100 !== 0) || gy % 400 === 0 ? 29 : 28,
    31,
    30,
    31,
    30,
    31,
    31,
    30,
    31,
    30,
    31,
  ];
  let gm = 0;
  for (gm = 1; gm <= 12 && gd > salA[gm]; gm += 1) gd -= salA[gm];
  return { gy, gm, gd };
}

function isJalaliLeap(jy) {
  const breaks = [
    -61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210, 1635, 2060, 2097, 2192, 2262, 2324, 2394,
    2456, 3178,
  ];
  const bl = breaks.length;
  let jp = breaks[0];
  let jump = 0;
  for (let i = 1; i < bl; i += 1) {
    const jm = breaks[i];
    jump = jm - jp;
    if (jy < jm) break;
    jp = jm;
  }
  let n = jy - jp;
  if (jump - n < 6) n = n - jump + div(jump + 4, 33) * 33;
  let leap = ((((n + 1) % 33) - 1) % 4);
  if (leap === -1) leap = 4;
  return leap === 0;
}

function jalaliMonthLength(jy, jm) {
  if (jm <= 6) return 31;
  if (jm <= 11) return 30;
  return isJalaliLeap(jy) ? 30 : 29;
}

function pad2(n) {
  return String(n).padStart(2, '0');
}

function formatJalali(jy, jm, jd) {
  return `${jy}/${pad2(jm)}/${pad2(jd)}`;
}

function toEnglishDigits(value) {
  return String(value ?? '')
    .replace(/[۰-۹]/g, (d) => '۰۱۲۳۴۵۶۷۸۹'.indexOf(d))
    .replace(/[٠-٩]/g, (d) => '٠١٢٣٤٥٦٧٨٩'.indexOf(d));
}

function parseJalaliDate(value) {
  const raw = toEnglishDigits(value).trim();
  if (!raw) return null;

  const jalaliMatch = raw.match(/^(\d{4})[\/\-.](\d{1,2})[\/\-.](\d{1,2})$/);
  if (jalaliMatch) {
    const jy = Number(jalaliMatch[1]);
    const jm = Number(jalaliMatch[2]);
    const jd = Number(jalaliMatch[3]);
    if (jy >= 1200 && jy <= 1500 && jm >= 1 && jm <= 12 && jd >= 1 && jd <= jalaliMonthLength(jy, jm)) {
      return { jy, jm, jd };
    }
  }

  const isoMatch = raw.match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (isoMatch) {
    const gy = Number(isoMatch[1]);
    const gm = Number(isoMatch[2]);
    const gd = Number(isoMatch[3]);
    if (gy >= 1800 && gy <= 2100) {
      return gregorianToJalali(gy, gm, gd);
    }
  }

  return null;
}

function getTodayJalali() {
  const now = new Date();
  return gregorianToJalali(now.getFullYear(), now.getMonth() + 1, now.getDate());
}

function jalaliWeekStartOffset(jy, jm) {
  const { gy, gm, gd } = jalaliToGregorian(jy, jm, 1);
  // JS: 0=Sun ... 6=Sat → Jalali week starts Saturday
  const weekDay = new Date(gy, gm - 1, gd).getDay();
  return (weekDay + 1) % 7;
}

function JalaliDatePicker({ id, name, value, onChange, disabled, className }) {
  const selected = React.useMemo(() => parseJalaliDate(value), [value]);
  const today = React.useMemo(() => getTodayJalali(), []);
  const [open, setOpen] = React.useState(false);
  const [view, setView] = React.useState(() => selected || today);
  const rootRef = React.useRef(null);

  React.useEffect(() => {
    if (selected) setView(selected);
  }, [selected?.jy, selected?.jm, selected?.jd]);

  React.useEffect(() => {
    if (!open) return undefined;
    const onPointerDown = (event) => {
      if (rootRef.current && !rootRef.current.contains(event.target)) {
        setOpen(false);
      }
    };
    const onKeyDown = (event) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onPointerDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousedown', onPointerDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [open]);

  const emit = (nextValue) => {
    onChange?.({
      target: {
        name,
        value: nextValue,
      },
    });
  };

  const monthLength = jalaliMonthLength(view.jy, view.jm);
  const offset = jalaliWeekStartOffset(view.jy, view.jm);
  const cells = [];
  for (let i = 0; i < offset; i += 1) cells.push(null);
  for (let day = 1; day <= monthLength; day += 1) cells.push(day);

  const shiftMonth = (delta) => {
    let nextJm = view.jm + delta;
    let nextJy = view.jy;
    if (nextJm < 1) {
      nextJm = 12;
      nextJy -= 1;
    } else if (nextJm > 12) {
      nextJm = 1;
      nextJy += 1;
    }
    setView({ jy: nextJy, jm: nextJm, jd: Math.min(view.jd || 1, jalaliMonthLength(nextJy, nextJm)) });
  };

  const years = [];
  for (let y = today.jy; y >= today.jy - 80; y -= 1) years.push(y);

  return (
    <div className="relative w-full" ref={rootRef}>
      <div className="relative">
        <input
          id={id}
          name={name}
          type="text"
          inputMode="numeric"
          autoComplete="bday"
          placeholder="۱۳۷۰/۰۱/۱۵"
          className={`${className} pe-10`}
          value={value ?? ''}
          disabled={disabled}
          readOnly={disabled}
          dir="ltr"
          onFocus={() => {
            if (!disabled) setOpen(true);
          }}
          onClick={() => {
            if (!disabled) setOpen(true);
          }}
          onChange={(event) => {
            emit(toEnglishDigits(event.target.value).replace(/[^\d\/\-.]/g, ''));
          }}
        />
        <button
          type="button"
          tabIndex={-1}
          disabled={disabled}
          aria-label="باز کردن تقویم"
          className="absolute inset-y-0 end-0 z-10 flex items-center pe-3 text-[var(--tavan-blue)] disabled:text-gray-300"
          onClick={() => {
            if (!disabled) setOpen((prev) => !prev);
          }}
        >
          <span className="material-icons text-[1.2rem]">calendar_month</span>
        </button>
      </div>

      {open && !disabled ? (
        <div className="relative z-20 mt-2 w-full max-w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-3 shadow-md">
          <div className="mb-3 flex items-center justify-between gap-2">
            <button
              type="button"
              className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg hover:bg-gray-100"
              aria-label="ماه بعد"
              onClick={() => shiftMonth(1)}
            >
              <span className="material-icons text-[1.1rem]">chevron_right</span>
            </button>

            <div className="flex min-w-0 flex-1 items-center justify-center gap-2">
              <select
                className="max-w-[45%] rounded-md border border-gray-200 bg-white px-2 py-1 text-sm"
                value={view.jm}
                onChange={(event) => {
                  const jm = Number(event.target.value);
                  setView((prev) => ({
                    ...prev,
                    jm,
                    jd: Math.min(prev.jd || 1, jalaliMonthLength(prev.jy, jm)),
                  }));
                }}
              >
                {JALALI_MONTHS.map((month, index) => (
                  <option key={month} value={index + 1}>
                    {month}
                  </option>
                ))}
              </select>
              <select
                className="max-w-[40%] rounded-md border border-gray-200 bg-white px-2 py-1 text-sm"
                value={view.jy}
                onChange={(event) => {
                  const jy = Number(event.target.value);
                  setView((prev) => ({
                    ...prev,
                    jy,
                    jd: Math.min(prev.jd || 1, jalaliMonthLength(jy, prev.jm)),
                  }));
                }}
              >
                {years.map((year) => (
                  <option key={year} value={year}>
                    {year}
                  </option>
                ))}
              </select>
            </div>

            <button
              type="button"
              className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg hover:bg-gray-100"
              aria-label="ماه قبل"
              onClick={() => shiftMonth(-1)}
            >
              <span className="material-icons text-[1.1rem]">chevron_left</span>
            </button>
          </div>

          <div className="mb-1 grid grid-cols-7 gap-1 text-center text-[0.7rem] text-gray-400">
            {JALALI_WEEKDAYS.map((day) => (
              <span key={day}>{day}</span>
            ))}
          </div>

          <div className="grid grid-cols-7 gap-1">
            {cells.map((day, index) => {
              if (!day) {
                return <span key={`empty-${index}`} className="h-9" />;
              }

              const isSelected =
                selected &&
                selected.jy === view.jy &&
                selected.jm === view.jm &&
                selected.jd === day;
              const isToday =
                today.jy === view.jy && today.jm === view.jm && today.jd === day;

              return (
                <button
                  key={`${view.jy}-${view.jm}-${day}`}
                  type="button"
                  className={[
                    'h-9 rounded-lg text-sm transition',
                    isSelected
                      ? 'bg-[var(--tavan-blue)] font-bold text-white'
                      : isToday
                        ? 'border border-[var(--tavan-blue)] text-[var(--tavan-blue)]'
                        : 'hover:bg-gray-100',
                  ].join(' ')}
                  onClick={() => {
                    emit(formatJalali(view.jy, view.jm, day));
                    setOpen(false);
                  }}
                >
                  {day}
                </button>
              );
            })}
          </div>

          <div className="mt-3 flex items-center justify-between gap-2 border-t border-gray-100 pt-2">
            <button
              type="button"
              className="rounded-lg px-2 py-1 text-xs text-gray-500 hover:bg-gray-100"
              onClick={() => {
                emit('');
                setOpen(false);
              }}
            >
              پاک کردن
            </button>
            <button
              type="button"
              className="rounded-lg px-2 py-1 text-xs text-[var(--tavan-blue)] hover:bg-gray-100"
              onClick={() => {
                emit(formatJalali(today.jy, today.jm, today.jd));
                setView(today);
                setOpen(false);
              }}
            >
              امروز
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function UserInfoFormField({
  icon,
  label,
  name,
  value,
  onChange,
  disabled,
  readOnly,
  type = 'text',
  options = [],
  fullWidth = false,
}) {
  const isSelect = type === 'select';
  const isNumber = type === 'number';
  const isDate = type === 'date';
  const inputId = name;
  const isDisabled = disabled || readOnly;

  const baseInputClass = 'tavan-modaresan-form__input w-full transition-all duration-200';

  const disabledClass = isDisabled ? 'bg-gray-100 text-gray-400 cursor-not-allowed opacity-70 border-gray-200 placeholder:text-gray-300' : 'bg-white text-[var(--tavan-text)]';

  const inputClassName = `${baseInputClass} ${disabledClass}`;

  return (
    <div className={`tavan-modaresan-form__field overflow-visible${fullWidth || isDate ? ' md:col-span-2' : ''}`}>
      <label className={`tavan-modaresan-form__label flex items-start gap-2 ${isDisabled ? 'text-gray-400' : ''}`} htmlFor={inputId}>
        {icon ? (
          <span className={`material-icons text-[1.125rem] mt-0.5 shrink-0 ${isDisabled ? 'text-gray-400' : 'text-[var(--tavan-blue)]'}`}>
            {icon}
          </span>
        ) : null}
        <span>{label}</span>
      </label>

      {isSelect ? (
        <select
          id={inputId}
          name={name}
          className={inputClassName}
          value={value ?? ''}
          onChange={onChange}
          disabled={isDisabled}
        >
          <option value="">انتخاب کنید…</option>
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      ) : isDate ? (
        <JalaliDatePicker
          id={inputId}
          name={name}
          value={value}
          onChange={onChange}
          disabled={isDisabled}
          className={inputClassName}
        />
      ) : (
          <input
            id={inputId}
            name={name}
            type={isNumber ? 'number' : 'text'}
            min={isNumber ? 0 : undefined}
            className={inputClassName}
            value={value ?? ''}
            onChange={onChange}
            disabled={isDisabled}
            readOnly={isDisabled}
            dir="rtl"
          />
        )}
    </div>
  );
}

function SectionUserInfo() {
  const { userData, reload } = getUserHrmData();

  const [activeTab, setActiveTab] = React.useState('personal');
  const [isEditing, setIsEditing] = React.useState(false);
  const [form, setForm] = React.useState({
    personal: {},
    school: {},
    job: {},
    extra: {},
  });
  const [saving, setSaving] = React.useState(false);

  const info = userData && typeof buildUserInfoModel === 'function' ? buildUserInfoModel(userData) : null;

  React.useEffect(() => {
    if (!userData) return;
    setForm(buildFormFromData(userData, info));
  }, [userData]);

  if (!userData) {
    return (
      <div className="tavan-wall tavan-userinfo" dir="rtl">
        <div className="tavan-userinfo__loading">
          <LoadingOval visible />
        </div>
      </div>
    );
  }

  const handleChange = (section) => (e) => {
    const { name, value } = e.target;
    setForm((prev) => {
      const nextSection = { ...prev[section], [name]: value };

      if (section === 'extra' && name === 'maritalStatus' && value !== 'married') {
        nextSection.childrenCount = '';
      }

      if (section === 'extra' && name === 'lastEducationDegree' && value === 'bachelor') {
        nextSection.bachelorDegree = '';
        nextSection.masterDegree = '';
        nextSection.masterUniversity = '';
        nextSection.doctorateDegree = '';
        nextSection.doctorateUniversity = '';
        nextSection.doctorateInProgress = '';
      }

      return { ...prev, [section]: nextSection };
    });
  };

  const handleCancel = () => {
    setIsEditing(false);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload = buildSavePayload(form, userData);
      const hasSavedProfile = userData?.meta?.hasSavedProfile === true;

      const provider = hasSavedProfile
        ? 'fr/emis-edus/profile/person-profiles/update'
        : 'fr/emis-edus/profile/person-profiles/save';

      const res = await $self.fr({
        provider,
        data: payload,
      });

      if (!res?.data?.Result?.ok) {
        throw new Error(!res?.data?.Result?.message || 'خطا در ذخیره اطلاعات');
      }

      setIsEditing(false);
      $self.alert({ type: 'info', message: 'اطلاعات پروفایل شما با موفقیت ذخیره شد' });
      await reload();
    } catch (error) {
      console.error('خطا در ذخیره پروفایل:', error);
      alert(error?.message || 'خطا در ذخیره اطلاعات. دوباره تلاش کنید.');
    } finally {
      setSaving(false);
    }
  };

  const fields = FIELD_DEFS[activeTab] || [];
  const isMarried = form.extra?.maritalStatus === 'married';
  const activeTabMeta = TABS.find((t) => t.id === activeTab);
  const lastDegree = form.extra?.lastEducationDegree;
  const isBachelorOnly = lastDegree === 'bachelor';

  const editActions = !isEditing ? (
    <button
      title="ویرایش"
      type="button"
      className="tavan-wall-btn tavan-wall-btn--soft tavan-wall-btn--sm justify-center whitespace-nowrap"
      onClick={() => setIsEditing(true)}
    >
      <span className="material-icons text-[1.125rem]">edit</span>
    </button>
  ) : (
      <>
        <button
          title="انصراف"
          type="button"
          className="tavan-wall-btn tavan-wall-btn--ghost tavan-wall-btn--sm justify-center whitespace-nowrap"
          onClick={handleCancel}
          disabled={saving}
        >
          انصراف
      </button>
        <button
          title="ذخیره"
          type="button"
          className="tavan-wall-btn tavan-wall-btn--primary tavan-wall-btn--sm justify-center whitespace-nowrap"
          onClick={handleSave}
          disabled={saving}
        >
          {saving ? (
            'در حال ذخیره…'
          ) : (
              <>
                <span className="material-icons text-[1.125rem]">save</span>
            ذخیره
          </>
            )}
        </button>
      </>
    );

  return (
    <div className="tavan-wall tavan-userinfo w-full min-w-0 max-w-full pb-6" dir="rtl">
      <header className="tavan-userinfo__header mb-4 sm:mb-5 lg:mb-6 flex flex-wrap sm:flex-nowrap items-start lg:items-center justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <div className="tavan-userinfo__accent shrink-0" aria-hidden="true" />
          <div className="min-w-0">
            <h2 className="tavan-userinfo__title text-xl sm:text-2xl">پروفایل شما</h2>
            <p className="tavan-userinfo__subtitle line-clamp-2 sm:line-clamp-none !text-gray-500/80">
              اطلاعات فردی، شغلی، محل خدمت و تکمیلی مشاوران
            </p>
          </div>
        </div>
      </header>

      <div className="mb-4 sm:mb-5 lg:mb-6 w-full lg:flex lg:justify-center">
        <div className="w-full overflow-x-auto overflow-y-hidden pb-0.5 -mx-1 px-1 lg:mx-0 lg:px-0 lg:overflow-visible lg:w-auto">
          <div className="tavan-modaresan-tabs mb-0 inline-flex flex-nowrap sm:flex-wrap w-max min-w-full sm:w-full lg:w-auto lg:min-w-0" role="tablist">
            {TABS.map((tab) => (
              <button
                key={tab.id}
                type="button"
                role="tab"
                aria-selected={activeTab === tab.id}
                className={`tavan-modaresan-tabs__btn shrink-0 whitespace-nowrap text-xs sm:text-sm px-3.5 py-2 sm:px-[1.125rem] sm:py-2.5${activeTab === tab.id ? ' is-active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                <span className="material-icons text-base sm:text-[1.125rem]">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      <article className="tavan-userinfo-card overflow-visible">
        <header className="tavan-userinfo-card__header flex flex-wrap items-center gap-2.5">
          <div className="tavan-userinfo-card__header-icon">
            <span className="material-icons">{activeTabMeta?.icon}</span>
          </div>
          <h3 className="tavan-userinfo-card__header-title text-sm sm:text-base">
            {activeTabMeta?.label}
          </h3>

          {/* متن حالت + دکمه ویرایش در سمت راست متن */}
          <div className="w-full sm:w-auto sm:mr-auto flex items-center gap-2.5">
            {/* <div className="flex items-center gap-2">
              {editActions}
            </div>
             <span className="text-xs font-semibold text-[var(--tavan-text-muted)] whitespace-nowrap">
              {isEditing ? 'حالت ویرایش' : 'حالت مشاهده'}
            </span>  */}
          </div>
        </header>

        <div className="tavan-userinfo-card__body overflow-visible p-4 sm:p-[1.375rem]">
          <form
            className="tavan-modaresan-form overflow-visible"
            onSubmit={(e) => {
              e.preventDefault();
              if (isEditing) handleSave();
            }}
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-x-5 overflow-visible">
              {fields.map((field) => {
                const disabledChildren = field.key === 'childrenCount' && !isMarried;

                const disabledByDegree =
                  isBachelorOnly &&
                  [
                    'bachelorDegree',
                    'masterDegree',
                    'masterUniversity',
                    'doctorateDegree',
                    'doctorateUniversity',
                    'doctorateInProgress',
                  ].includes(field.key);

                return (
                  <UserInfoFormField
                    key={field.key}
                    icon={field.icon}
                    label={field.label}
                    name={field.key}
                    value={form[activeTab]?.[field.key]}
                    onChange={handleChange(activeTab)}
                    disabled={!isEditing || disabledChildren || disabledByDegree}
                    readOnly={field.readOnly}
                    type={field.type}
                    options={field.options}
                    fullWidth={field.fullWidth}
                  />
                );
              })}
            </div>

            {activeTab === 'extra' && (
              <p className="mt-3 text-xs text-gray-500/90 leading-relaxed">
                این بخش مربوط به فرم اولیه تکمیل اطلاعات مشاوران است. فیلدهایی که هنوز در
                سامانه ثبت نشده‌اند خالی هستند و پس از ذخیره اطلاعات شما تکمیل می‌شوند.
              </p>
            )}

            {/* پایین فرم — دکمه ویرایش یا دکمه‌های ذخیره/انصراف */}
            <div className="tavan-modaresan-actions mt-6 flex flex-col-reverse sm:flex-row sm:justify-end gap-2.5">
              {!isEditing ? (
                <button
                  title="ویرایش"
                  type="button"
                  className="tavan-wall-btn tavan-wall-btn--soft w-full sm:w-auto justify-center"
                  onClick={() => setIsEditing(true)}
                >
                  <span className="material-icons text-[1.125rem]">edit</span>
      ویرایش
                </button>
              ) : (
                  <>
                    <button
                      title="انصراف"
                      type="button"
                      className="tavan-wall-btn tavan-wall-btn--ghost w-full sm:w-auto justify-center"
                      onClick={handleCancel}
                      disabled={saving}
                    >
                      انصراف
      </button>
                    <button
                      title="ذخیره تغییرات"
                      type="submit"
                      className="tavan-wall-btn tavan-wall-btn--primary w-full sm:w-auto justify-center"
                      disabled={saving}
                    >
                      <span className="material-icons text-[1.125rem]">save</span>
                      <span>{saving ? 'در حال ذخیره…' : 'ذخیره تغییرات'}</span>
                    </button>
                  </>
                )}
            </div>
          </form>
        </div>
      </article>
    </div>
  );
}