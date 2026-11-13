-- Application rollback should retain this table and keys. This explicit down
-- migration discards replay history; system roles remain for existing accounts.
DROP TABLE identity.onboard_requests;
