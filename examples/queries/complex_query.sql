-- Complex SQL Server query example
SELECT 
    u.user_id,
    u.username,
    u.email,
    p.first_name,
    p.last_name,
    COUNT(o.order_id) as total_orders,
    SUM(o.total_amount) as total_spent,
    AVG(o.total_amount) as avg_order_value,
    MAX(o.order_date) as last_order_date
FROM 
    users u
    INNER JOIN profiles p ON u.user_id = p.user_id
    LEFT JOIN orders o ON u.user_id = o.customer_id
WHERE 
    u.created_date >= '2023-01-01'
    AND u.status = 'active'
    AND p.country IN ('US', 'CA', 'UK')
GROUP BY 
    u.user_id, u.username, u.email, p.first_name, p.last_name
HAVING 
    COUNT(o.order_id) > 5
ORDER BY 
    total_spent DESC, last_order_date DESC;
