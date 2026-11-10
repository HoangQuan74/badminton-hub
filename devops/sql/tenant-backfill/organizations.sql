COPY (
    SELECT id, replace(encode(convert_to(name, 'UTF8'), 'base64'), E'\n', ''),
           replace(encode(convert_to(slug, 'UTF8'), 'base64'), E'\n', ''),
           status, created_at, updated_at
    FROM catalog.organizations ORDER BY id
) TO STDOUT WITH (FORMAT csv);
