-- PostgreSQL example with double quoted identifiers
SELECT 
    "user_id",
    "first_name",
    "last_name",
    "email"
FROM "users"
WHERE "status" = 'active'
    AND "created_at" >= '2024-01-01'
ORDER BY "last_name", "first_name";

-- PostgreSQL-specific features
SELECT 
    "u"."user_id",
    "u"."email",
    COUNT("o"."order_id") AS "order_count"
FROM "users" "u"
LEFT JOIN "orders" "o" ON "u"."user_id" = "o"."user_id"
GROUP BY "u"."user_id", "u"."email"
HAVING "order_count" > 0
ORDER BY "order_count" DESC
LIMIT 10;
