/**
 * Fixed getDetailsDataV2 + getProviderCourses
 *
 * Root cause of mismatch:
 * - Preview (8 per provider) ordered: FREE first, then cr DESC
 * - Full list ordered: cr DESC only + LIMIT 200
 *   → older/free courses shown in preview could fall outside the 200 newest
 * - Keep the SAME WHERE / EXISTS filters and the SAME ORDER BY in both queries
 *   so the preview set is always a subset of the full provider list.
 *
 * Drop this into your service and wire academyAppDB the same way as before.
 */

// Shared free-course helper (JS)
const isFreeCourse = (c: any): boolean => {
  const amount = c.imMeta?.subscribe?.fee?.default?.[0]?.amount ?? 0;
  if (amount === null || amount === undefined || amount === '' || Number(amount) === 0) {
    return true;
  }
  return false;
};

/** SQL expression: 0 = free, 1 = paid — must be identical in both queries */
const FREE_SORT_SQL = `
  CASE
    WHEN COALESCE(
      NULLIF(
        REPLACE(
          im.meta -> 'subscribe' -> 'fee' -> 'default' -> 0 ->> 'amount',
          ',',
          ''
        ),
        ''
      )::numeric,
      0
    ) = 0
    THEN 0
    ELSE 1
  END
`;

/** Shared SELECT + JOINs for course rows */
const COURSE_SELECT_SQL = `
  im.rand_id AS id,
  pl.rand_id AS "planId",
  pl.meta->>'thumb' AS thumb,
  im.meta AS "imMeta",
  im.status,
  im.academy,
  im.cr,
  im.meta AS doc,

  a.rand_id AS academy_id,
  a.name AS academy_name,
  a.title AS academy_title,

  jsonb_build_object(
    'services',
    COALESCE(
      (
        SELECT jsonb_agg(
          jsonb_build_object(
            'service_id', cs2.service_id,
            'show_in_service_wall', cs2.show_in_service_wall,
            'allow_all_academies', cs2.allow_all_academies,
            'academies', cs2.academies
          )
        )
        FROM entity.emis_course_service cs2
        WHERE cs2.rand_id = im.rand_id
          AND cs2.status = 1
      ),
      '[]'::jsonb
    )
  ) AS service
`;

/**
 * Shared WHERE for both endpoints.
 * :userAcademyId = آکادمی کاربر (برای allow_all_academies / academies)
 * :serviceId     = سرویس کلاینت
 */
const COURSE_WHERE_SQL = `
  im.rand_id LIKE '%IM%'
  AND im.status = 1
  AND a.status = 1
  AND 'provider' = ANY(a.type)
  AND (im.meta -> 'subscribe' ->> 'start')::timestamp < NOW()
  AND (im.meta -> 'schedule' ->> 'end')::timestamp > NOW()
  AND pl.rand_id NOT LIKE '%IM%'
  AND pl.status = 1
  AND pl.meta->>'thumb' IS NOT NULL
  AND pl.meta->>'thumb' != ''
  AND EXISTS (
    SELECT 1
    FROM entity.emis_course_service cs
    WHERE cs.rand_id = im.rand_id
      AND cs.service_id = :serviceId
      AND cs.status = 1
      AND (
        cs.allow_all_academies = true
        OR (
          cs.allow_all_academies = false
          AND :userAcademyId = ANY(cs.academies)
        )
      )
  )
`;

const COURSE_FROM_SQL = `
  FROM entity.emis_course AS im
  JOIN entity.emis_course AS pl
    ON pl.rand_id = split_part(im.rand_id, 'IM', 1)
  JOIN entity.emis_academy AS a
    ON a.rand_id = im.academy
`;

/**
 * All fields come flat on bodyData (no nested params).
 * If userAcademyId is missing, courses with allow_all_academies=false
 * (e.g. ویژه گمرک) are filtered OUT by COURSE_WHERE_SQL.
 */
function resolveRequestContext(bodyData: any) {
  let b: any = bodyData ?? {};

  if (typeof b === 'string') {
    try {
      b = JSON.parse(b);
    } catch {
      b = {};
    }
  }

  if (b === null || typeof b !== 'object' || Array.isArray(b)) {
    b = {};
  }

  const pick = (...keys: string[]) => {
    for (const key of keys) {
      const v = b[key];
      if (v === undefined || v === null) continue;
      if (typeof v === 'string') {
        const trimmed = v.trim();
        if (trimmed !== '') return trimmed;
        continue;
      }
      return v;
    }
    return null;
  };

  return {
    serviceId: pick('serviceId') || 'padafand.edus.ir',
    userId: pick('userId'),
    userAcademyId: pick('academyId', 'userAcademyId', 'academy'),
    referrer: pick('referrer'),
    providerAcademyId: pick(
      'providerAcademyId',
      'providerId',
      'targetAcademyId',
    ),
    providerName: pick('providerName'),
  };
}

async function attachDiscounts(
  academyAppDB: any,
  courses: any[],
  userId?: string,
): Promise<any[]> {
  const allImplIds = Array.from(
    new Set((courses || []).map((c: any) => c.id).filter(Boolean)),
  );

  const discountsByCourse: Record<string, any[]> = {};

  if (allImplIds.length === 0) {
    return (courses || []).map((c: any) => ({ ...c, discounts: [] }));
  }

  const discountsResult = await academyAppDB.query(
    `
      SELECT
        d.rand_id, d.code, d.title, d.type, d.value,
        d.course_ids, d.allow_all_courses,
        d.max_uses, d.used_count, d.max_uses_per_user,
        d.start_at, d.end_at, d.meta
      FROM entity.emis_discount_code d
      WHERE d.status = 1
        AND (d.start_at IS NULL OR d.start_at <= NOW())
        AND (d.end_at   IS NULL OR d.end_at   >= NOW())
        AND (
          d.allow_all_courses = true
          OR EXISTS (
            SELECT 1 FROM unnest(d.course_ids) AS cid
            WHERE cid IN (:implIds)
          )
        )
        AND (d.max_uses IS NULL OR d.used_count < d.max_uses)
      ORDER BY d.cr DESC
    `,
    { replacements: { implIds: allImplIds }, type: 'SELECT' },
  );

  const userUsageMap: Record<string, number> = {};

  if (userId && discountsResult?.length > 0) {
    const discountIds = discountsResult.map((d: any) => d.rand_id);
    const usageResult = await academyAppDB.query(
      `
        SELECT discount_id, COUNT(*)::int AS used_count
        FROM entity.emis_discount_usage
        WHERE user_id = :userId
          AND discount_id IN (:discountIds)
          AND status = 1
        GROUP BY discount_id
      `,
      { replacements: { userId, discountIds }, type: 'SELECT' },
    );

    (usageResult || []).forEach((u: any) => {
      userUsageMap[u.discount_id] = u.used_count;
    });
  }

  (discountsResult || []).forEach((d: any) => {
    const usedByUser = userUsageMap[d.rand_id] || 0;
    const remainingForUser =
      d.max_uses_per_user != null
        ? Math.max(0, d.max_uses_per_user - usedByUser)
        : null;

    if (remainingForUser !== null && remainingForUser <= 0) return;

    const discountItem = {
      id: d.rand_id,
      code: d.code,
      title: d.title,
      type: d.type,
      value: Number(d.value),
      maxUses: d.max_uses,
      usedCount: d.used_count,
      maxUsesPerUser: d.max_uses_per_user,
      usedByUser,
      remainingForUser,
      startAt: d.start_at,
      endAt: d.end_at,
      meta: d.meta || {},
    };

    const targets = d.allow_all_courses
      ? allImplIds
      : (d.course_ids || []).filter((cid: string) => allImplIds.includes(cid));

    targets.forEach((cid: string) => {
      if (!discountsByCourse[cid]) discountsByCourse[cid] = [];
      discountsByCourse[cid].push(discountItem);
    });
  });

  return (courses || []).map((c: any) => ({
    ...c,
    discounts: discountsByCourse[c.id] || [],
  }));
}

/** Same comparator for preview + full list: fewer services first, then newer cr */
function sortCoursesConsistent(a: any, b: any): number {
  const lenA = a?.service?.services?.length || 0;
  const lenB = b?.service?.services?.length || 0;

  if (lenA !== lenB) {
    return lenA - lenB;
  }

  const aCr = a.cr ? new Date(a.cr).getTime() : 0;
  const bCr = b.cr ? new Date(b.cr).getTime() : 0;
  return bCr - aCr;
}

export async function getDetailsDataV2(academyAppDB: any, bodyData: any) {
  try {
    const {
      serviceId: serviceFromClient,
      userId,
      userAcademyId: academyId,
      referrer,
    } = resolveRequestContext(bodyData);

    const providerAcademies = await academyAppDB.query(
      `
        SELECT
          a.rand_id,
          a.name,
          a.title,
          a.type,
          a.service,
          a.meta
        FROM entity.emis_academy a
        WHERE a.status = 1
          AND 'provider' = ANY(a.type)
        ORDER BY a.title ASC NULLS LAST, a.name ASC
      `,
      { type: 'SELECT' },
    );

    if (!providerAcademies?.length) {
      return {
        providers: [],
        userId,
        academyId,
        serviceId: serviceFromClient,
      };
    }

    const providerIds = providerAcademies.map((a: any) => a.rand_id);

    // Same filters + same free-first order as getProviderCourses; only rn <= 8 differs
    const courses = await academyAppDB.query(
      `
        SELECT *
        FROM (
          SELECT
            ${COURSE_SELECT_SQL},
            ROW_NUMBER() OVER (
              PARTITION BY im.academy
              ORDER BY ${FREE_SORT_SQL} ASC, im.cr DESC
            ) AS rn
          ${COURSE_FROM_SQL}
          WHERE
            ${COURSE_WHERE_SQL}
            AND im.academy IN (:providerIds)
        ) t
        WHERE t.rn <= 8
        ORDER BY t.academy_title ASC NULLS LAST, t.rn ASC
      `,
      {
        replacements: {
          providerIds,
          serviceId: serviceFromClient,
          userAcademyId: academyId || null,
        },
        type: 'SELECT',
      },
    );

    let coursesWithDiscounts = await attachDiscounts(
      academyAppDB,
      courses,
      userId,
    );

    coursesWithDiscounts = coursesWithDiscounts.sort(sortCoursesConsistent);

    const coursesByAcademy: Record<string, any[]> = {};
    coursesWithDiscounts.forEach((c: any) => {
      const key = c.academy_id || c.academy;
      if (!coursesByAcademy[key]) coursesByAcademy[key] = [];
      coursesByAcademy[key].push(c);
    });

    let providers = providerAcademies
      .map((a: any) => ({
        academyId: a.rand_id,
        name: a.name,
        title: a.title || a.name,
        service: a.service || [],
        meta: a.meta || {},
        courses: coursesByAcademy[a.rand_id] || [],
        hasMore: (coursesByAcademy[a.rand_id] || []).length >= 8,
        moreUrl: `/provider-courses/${a.name || a.rand_id}`,
      }))
      .filter((p: any) => p.courses.length > 0);

    if (academyId || referrer) {
      providers = providers.sort((a: any, b: any) => {
        const aUserMatch = a.academyId === academyId || a.name === academyId;
        const bUserMatch = b.academyId === academyId || b.name === academyId;
        if (aUserMatch && !bUserMatch) return -1;
        if (!aUserMatch && bUserMatch) return 1;

        if (referrer) {
          const aReferrerMatch =
            Array.isArray(a.service) && a.service.includes(referrer);
          const bReferrerMatch =
            Array.isArray(b.service) && b.service.includes(referrer);
          if (aReferrerMatch && !bReferrerMatch) return -1;
          if (!aReferrerMatch && bReferrerMatch) return 1;
        }
        return 0;
      });
    }

    return {
      providers,
      userId,
      academyId,
      serviceId: serviceFromClient,
    };
  } catch (err: any) {
    return {
      ok: false,
      error: 'ErrorMessage => ' + err,
    };
  }
}

export async function getProviderCourses(academyAppDB: any, bodyData: any) {
  try {
    const {
      serviceId: serviceFromClient,
      userId,
      userAcademyId,
      providerAcademyId,
      providerName,
    } = resolveRequestContext(bodyData);

    if (!providerAcademyId && !providerName) {
      return {
        ok: false,
        error: 'شناسه یا نام provider الزامی است',
      };
    }

    const providerRows: any = await academyAppDB.query(
      `
        SELECT
          a.rand_id,
          a.name,
          a.title,
          a.type,
          a.service,
          a.meta
        FROM entity.emis_academy a
        WHERE a.status = 1
          AND 'provider' = ANY(a.type)
          AND (
            (:providerAcademyId IS NOT NULL AND a.rand_id = :providerAcademyId)
            OR
            (:providerName IS NOT NULL AND a.name = :providerName)
          )
        LIMIT 1
      `,
      {
        replacements: {
          providerAcademyId: providerAcademyId || null,
          providerName: providerName || null,
        },
        type: 'SELECT',
      },
    );

    const provider = providerRows?.[0];
    if (!provider) {
      return {
        ok: false,
        error: 'آکادمی provider یافت نشد',
        courses: [],
      };
    }

    // Identical filters + identical ORDER BY as preview window (no arbitrary LIMIT 200)
    const courses = await academyAppDB.query(
      `
        SELECT
          ${COURSE_SELECT_SQL}
        ${COURSE_FROM_SQL}
        WHERE
          ${COURSE_WHERE_SQL}
          AND im.academy = :providerId
        ORDER BY ${FREE_SORT_SQL} ASC, im.cr DESC
      `,
      {
        replacements: {
          providerId: provider.rand_id,
          serviceId: serviceFromClient,
          userAcademyId,
        },
        type: 'SELECT',
      },
    );

    let coursesWithDiscounts = await attachDiscounts(
      academyAppDB,
      courses,
      userId,
    );

    // Keep same ordering as SQL / preview
    coursesWithDiscounts = coursesWithDiscounts.sort(sortCoursesConsistent);

    return {
      provider: {
        academyId: provider.rand_id,
        name: provider.name,
        title: provider.title || provider.name,
        meta: provider.meta || {},
      },
      courses: coursesWithDiscounts,
      total: coursesWithDiscounts.length,
      userId,
      academyId: userAcademyId,
      serviceId: serviceFromClient,
    };
  } catch (err: any) {
    return {
      ok: false,
      error: 'ErrorMessage => ' + err,
    };
  }
}
