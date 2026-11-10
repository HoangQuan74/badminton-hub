COPY (
    SELECT id, organization_id,
           replace(encode(convert_to(name, 'UTF8'), 'base64'), E'\n', ''),
           replace(encode(convert_to(slug, 'UTF8'), 'base64'), E'\n', ''),
           replace(encode(convert_to(address, 'UTF8'), 'base64'), E'\n', ''),
           replace(encode(convert_to(timezone, 'UTF8'), 'base64'), E'\n', ''),
           status, created_at, updated_at
    FROM catalog.venues ORDER BY id
) TO STDOUT WITH (FORMAT csv);
