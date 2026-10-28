
  // ------ Version 2 For Load Provider Academies With Courses ------
  async getDetailsDataV2(bodyData: any) {
    try {

      const isFreeCourse = (c: any): boolean => {
        const amount = c.imMeta?.subscribe?.fee?.default?.[0]?.amount ?? 0;

        if (amount === null || amount === undefined || amount === '' || Number(amount) === 0) {
          return true;
        }

        return false;
      };

      const serviceFromClient = bodyData?.serviceId || 'padafand.edus.ir';
      const referrer = bodyData?.referrer || null;

      const params = bodyData ?? {};
      const academyId = params.academyId;
      const userId = params.userId;

      // ─── ۱) آکادمی‌های provider ─────────────────────────────
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

      // ─── ۲) دوره‌های این آکادمی‌ها (حداکثر ۸ تا برای هر آکادمی) ─
      const courses = await academyAppDB.query(
        `
          SELECT *
          FROM (
            SELECT
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
              ) AS service,

              ROW_NUMBER() OVER (
                PARTITION BY im.academy
                ORDER BY
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
                  END ASC,
                  im.cr DESC
              ) AS rn

            FROM entity.emis_course AS im
            JOIN entity.emis_course AS pl
              ON pl.rand_id = split_part(im.rand_id, 'IM', 1)
            JOIN entity.emis_academy AS a
              ON a.rand_id = im.academy

            WHERE
              im.rand_id LIKE '%IM%'
              AND im.status = 1
              AND a.status = 1
              AND im.academy IN (:providerIds)
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
                      AND :academyId = ANY(cs.academies)
                    )
                  )
              )
          ) t
          WHERE t.rn <= 8
          ORDER BY t.academy_title ASC NULLS LAST, t.rn ASC
        `,
        {
          replacements: {
            providerIds,
            serviceId: serviceFromClient,
            academyId,
          },
          type: 'SELECT',
        },
      );

      // ─── حذف دوره‌های تکراری بر اساس imMeta.name (نگه داشتن جدیدترین) ───
      // if (Array.isArray(courses) && courses.length > 0) {
      //   const uniqueCoursesMap = new Map<string, any>();

      //   for (const course of courses) {
      //     const name = (course as any)?.imMeta?.name?.trim();
      //     if (!name) continue;

      //     const existing = uniqueCoursesMap.get(name);

      //     if (!existing) {
      //       uniqueCoursesMap.set(name, course);
      //       continue;
      //     }

      //     // مقایسه بر اساس cr (جدیدتر را نگه می‌داریم)
      //     const existingCr = existing.cr ? new Date(existing.cr).getTime() : 0;
      //     const currentCr = (course as any)?.cr ? new Date((course as any)?.cr).getTime() : 0;

      //     if (currentCr > existingCr) {
      //       uniqueCoursesMap.set(name, course);
      //     }
      //   }

      //   // جایگزینی آرایه با نسخه بدون تکرار
      //   courses.length = 0;
      //   courses.push(...uniqueCoursesMap.values());
      // }



      // ─── ۳) تخفیف‌ها ───────────────────────────────────────
      const allImplIds = Array.from(
        new Set((courses || []).map((c: any) => c.id).filter(Boolean)),
      );

      let discountsByCourse: Record<string, any[]> = {};

      if (allImplIds.length > 0) {
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
          {
            replacements: { implIds: allImplIds },
            type: 'SELECT',
          },
        );

        let userUsageMap: Record<string, number> = {};

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
            {
              replacements: { userId, discountIds },
              type: 'SELECT',
            },
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
            : (d.course_ids || []).filter((cid: string) =>
                allImplIds.includes(cid),
              );

          targets.forEach((cid: string) => {
            if (!discountsByCourse[cid]) discountsByCourse[cid] = [];
            discountsByCourse[cid].push(discountItem);
          });
        });
      }

      if (serviceFromClient === 'bonyad.edus.ir' && Array.isArray(courses) && courses.length > 0) {
        courses.sort((a: any, b: any) => {
          const aFree = isFreeCourse(a);
          const bFree = isFreeCourse(b);

          // اول رایگان‌ها
          if (aFree && !bFree) return -1;
          if (!aFree && bFree) return 1;

          // بعد بر اساس cr (جدیدتر اول)
          const aCr = a.cr ? new Date(a.cr).getTime() : 0;
          const bCr = b.cr ? new Date(b.cr).getTime() : 0;
          return bCr - aCr;
        });
      }


      // مرتب‌سازی بر اساس تعداد سرویس‌ها (تعداد کمتر اول)
      if (Array.isArray(courses) && courses.length > 0) {
        courses.sort((a: any, b: any) => {
          const lenA = a?.service?.services?.length || 0;
          const lenB = b?.service?.services?.length || 0;

          // اول آن‌هایی که services کمتری دارند
          if (lenA !== lenB) {
            return lenA - lenB;
          }

          // اگر تعداد برابر بود، بر اساس cr جدیدتر اول (اختیاری)
          const aCr = a.cr ? new Date(a.cr).getTime() : 0;
          const bCr = b.cr ? new Date(b.cr).getTime() : 0;
          return bCr - aCr;
        });
      }

      // ─── ۴) گروه‌بندی بر اساس آکادمی ──────────────────────
      const coursesByAcademy: Record<string, any[]> = {};

      (courses || []).forEach((c: any) => {
        const key = c.academy_id || c.academy;
        if (!coursesByAcademy[key]) coursesByAcademy[key] = [];
        coursesByAcademy[key].push({
          ...c,
          discounts: discountsByCourse[c.id] || [],
        });
      });

      // فقط آکادمی‌هایی که حداقل یک دوره دارند
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

      // آکادمی کاربر (academyId) بالاتر از بقیه
     // آکادمی کاربر در اولویت اول
      // سپس آکادمی‌هایی که Referrer در service آن‌ها وجود دارد
      if (academyId || referrer) {
        providers = providers.sort((a: any, b: any) => {
          // اولویت آکادمی کاربر
          const aUserMatch =
            a.academyId === academyId || a.name === academyId;

          const bUserMatch =
            b.academyId === academyId || b.name === academyId;

          if (aUserMatch && !bUserMatch) return -1;
          if (!aUserMatch && bUserMatch) return 1;

          // اولویت دامین Referrer
          if (referrer) {
            const aReferrerMatch =
              Array.isArray(a.service) &&
              a.service.includes(referrer);

            const bReferrerMatch =
              Array.isArray(b.service) &&
              b.service.includes(referrer);

            if (aReferrerMatch && !bReferrerMatch) return -1;
            if (!aReferrerMatch && bReferrerMatch) return 1;
          }

          // حفظ ترتیب قبلی در صورت برابر بودن اولویت‌ها
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




  // ------------------------------------------------

    // async getProviderCourses(bodyData: any) {
  //   try {
  //     const serviceFromClient = bodyData?.serviceId || 'padafand.edus.ir';
  //     const params = bodyData ?? {};

  //     const userId = params.userId;
  //     // آکادمی کاربر (برای فیلتر service)
  //     const userAcademyId = params.academyId || params.userAcademyId;
  //     // آکادمی/provider هدف (از moreUrl)
  //     const providerAcademyId =
  //       params.providerAcademyId ||
  //       params.providerId ||
  //       params.targetAcademyId ||
  //       null;
  //     const providerName = params.providerName || params.name || null;

  //     if (!providerAcademyId && !providerName) {
  //       return {
  //         ok: false,
  //         error: 'شناسه یا نام provider الزامی است',
  //       };
  //     }

  //     // ─── ۱) اطلاعات provider ─────────────────────────────
  //     const providerRows: any = await academyAppDB.query(
  //       `
  //         SELECT
  //           a.rand_id,
  //           a.name,
  //           a.title,
  //           a.type,
  //           a.service,
  //           a.meta
  //         FROM entity.emis_academy a
  //         WHERE a.status = 1
  //           AND 'provider' = ANY(a.type)
  //           AND (
  //             (:providerAcademyId IS NOT NULL AND a.rand_id = :providerAcademyId)
  //             OR
  //             (:providerName IS NOT NULL AND a.name = :providerName)
  //           )
  //         LIMIT 1
  //       `,
  //       {
  //         replacements: {
  //           providerAcademyId: providerAcademyId || null,
  //           providerName: providerName || null,
  //         },
  //         type: 'SELECT',
  //       },
  //     );

  //     const provider = providerRows?.[0];

  //     if (!provider) {
  //       return {
  //         ok: false,
  //         error: 'آکادمی provider یافت نشد',
  //         courses: [],
  //       };
  //     }

  //     // ─── ۲) همه دوره‌های این provider (حداکثر ۲۰۰) ───────
  //     const courses = await academyAppDB.query(
  //       `
  //         SELECT
  //           im.rand_id AS id,
  //           pl.rand_id AS "planId",
  //           pl.meta->>'thumb' AS thumb,
  //           im.meta AS "imMeta",
  //           im.status,
  //           im.academy,
  //           im.cr,
  //           im.meta AS doc,

  //           a.rand_id AS academy_id,
  //           a.name AS academy_name,
  //           a.title AS academy_title,

  //           jsonb_build_object(
  //             'services',
  //             COALESCE(
  //               (
  //                 SELECT jsonb_agg(
  //                   jsonb_build_object(
  //                     'service_id', cs2.service_id,
  //                     'show_in_service_wall', cs2.show_in_service_wall,
  //                     'allow_all_academies', cs2.allow_all_academies,
  //                     'academies', cs2.academies
  //                   )
  //                 )
  //                 FROM entity.emis_course_service cs2
  //                 WHERE cs2.rand_id = im.rand_id
  //                   AND cs2.status = 1
  //               ),
  //               '[]'::jsonb
  //             )
  //           ) AS service

  //         FROM entity.emis_course AS im
  //         JOIN entity.emis_course AS pl
  //           ON pl.rand_id = split_part(im.rand_id, 'IM', 1)
  //         JOIN entity.emis_academy AS a
  //           ON a.rand_id = im.academy

  //         WHERE
  //           im.rand_id LIKE '%IM%'
  //           AND im.status = 1
  //           AND a.status = 1
  //           AND im.academy = :providerId
  //           AND 'provider' = ANY(a.type)
  //           AND (im.meta -> 'subscribe' ->> 'start')::timestamp < NOW()
  //           AND (im.meta -> 'schedule' ->> 'end')::timestamp > NOW()
  //           AND pl.rand_id NOT LIKE '%IM%'
  //           AND pl.status = 1
  //           AND pl.meta->>'thumb' IS NOT NULL
  //           AND pl.meta->>'thumb' != ''

  //           AND EXISTS (
  //             SELECT 1
  //             FROM entity.emis_course_service cs
  //             WHERE cs.rand_id = im.rand_id
  //               AND cs.service_id = :serviceId
  //               AND cs.status = 1
  //               AND (
  //                 cs.allow_all_academies = true
  //                 OR (
  //                   cs.allow_all_academies = false
  //                   AND :userAcademyId = ANY(cs.academies)
  //                 )
  //               )
  //           )

  //         ORDER BY im.cr DESC
  //         LIMIT 200
  //       `,
  //       {
  //         replacements: {
  //           providerId: provider.rand_id,
  //           serviceId: serviceFromClient,
  //           userAcademyId: userAcademyId || null,
  //         },
  //         type: 'SELECT',
  //       },
  //     );

  //     // ─── ۳) تخفیف‌ها ─────────────────────────────────────
  //     const allImplIds = Array.from(
  //       new Set((courses || []).map((c: any) => c.id).filter(Boolean)),
  //     );

  //     let discountsByCourse: Record<string, any[]> = {};

  //     if (allImplIds.length > 0) {
  //       const discountsResult = await academyAppDB.query(
  //         `
  //           SELECT
  //             d.rand_id, d.code, d.title, d.type, d.value,
  //             d.course_ids, d.allow_all_courses,
  //             d.max_uses, d.used_count, d.max_uses_per_user,
  //             d.start_at, d.end_at, d.meta
  //           FROM entity.emis_discount_code d
  //           WHERE d.status = 1
  //             AND (d.start_at IS NULL OR d.start_at <= NOW())
  //             AND (d.end_at   IS NULL OR d.end_at   >= NOW())
  //             AND (
  //               d.allow_all_courses = true
  //               OR EXISTS (
  //                 SELECT 1 FROM unnest(d.course_ids) AS cid
  //                 WHERE cid IN (:implIds)
  //               )
  //             )
  //             AND (d.max_uses IS NULL OR d.used_count < d.max_uses)
  //           ORDER BY d.cr DESC
  //         `,
  //         {
  //           replacements: { implIds: allImplIds },
  //           type: 'SELECT',
  //         },
  //       );

  //       let userUsageMap: Record<string, number> = {};

  //       if (userId && discountsResult?.length > 0) {
  //         const discountIds = discountsResult.map((d: any) => d.rand_id);
  //         const usageResult = await academyAppDB.query(
  //           `
  //             SELECT discount_id, COUNT(*)::int AS used_count
  //             FROM entity.emis_discount_usage
  //             WHERE user_id = :userId
  //               AND discount_id IN (:discountIds)
  //               AND status = 1
  //             GROUP BY discount_id
  //           `,
  //           {
  //             replacements: { userId, discountIds },
  //             type: 'SELECT',
  //           },
  //         );

  //         (usageResult || []).forEach((u: any) => {
  //           userUsageMap[u.discount_id] = u.used_count;
  //         });
  //       }

  //       (discountsResult || []).forEach((d: any) => {
  //         const usedByUser = userUsageMap[d.rand_id] || 0;
  //         const remainingForUser =
  //           d.max_uses_per_user != null
  //             ? Math.max(0, d.max_uses_per_user - usedByUser)
  //             : null;

  //         if (remainingForUser !== null && remainingForUser <= 0) return;

  //         const discountItem = {
  //           id: d.rand_id,
  //           code: d.code,
  //           title: d.title,
  //           type: d.type,
  //           value: Number(d.value),
  //           maxUses: d.max_uses,
  //           usedCount: d.used_count,
  //           maxUsesPerUser: d.max_uses_per_user,
  //           usedByUser,
  //           remainingForUser,
  //           startAt: d.start_at,
  //           endAt: d.end_at,
  //           meta: d.meta || {},
  //         };

  //         const targets = d.allow_all_courses
  //           ? allImplIds
  //           : (d.course_ids || []).filter((cid: string) =>
  //               allImplIds.includes(cid),
  //             );

  //         targets.forEach((cid: string) => {
  //           if (!discountsByCourse[cid]) discountsByCourse[cid] = [];
  //           discountsByCourse[cid].push(discountItem);
  //         });
  //       });
  //     }

  //     const coursesWithDiscounts = (courses || []).map((c: any) => ({
  //       ...c,
  //       discounts: discountsByCourse[c.id] || [],
  //     }));

  //     return {
  //       provider: {
  //         academyId: provider.rand_id,
  //         name: provider.name,
  //         title: provider.title || provider.name,
  //         meta: provider.meta || {},
  //       },
  //       courses: coursesWithDiscounts,
  //       total: coursesWithDiscounts.length,
  //       userId,
  //       academyId: userAcademyId,
  //       serviceId: serviceFromClient,
  //     };
  //   } catch (err: any) {
  //     return {
  //       ok: false,
  //       error: 'ErrorMessage => ' + err,
  //     };
  //   }
  // }