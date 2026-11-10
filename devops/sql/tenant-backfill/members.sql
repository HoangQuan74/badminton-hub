COPY (
    SELECT organization_id, user_id, role, status, created_at
    FROM catalog.organization_members ORDER BY organization_id, user_id
) TO STDOUT WITH (FORMAT csv);
