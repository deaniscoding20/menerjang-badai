-- name: ListBrands :many
SELECT b.id::text AS id,b.name,b.description,b.website,b.status FROM brands b
JOIN user_brand_access g ON (g.org_id,g.brand_id)=(b.org_id,b.id)
JOIN organization_members m ON (m.org_id,m.user_id)=(g.org_id,g.user_id)
JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.org_id
WHERE b.org_id=sqlc.arg(org_id)::uuid AND g.user_id=sqlc.arg(user_id)::uuid AND g.active AND m.active AND u.active AND o.active ORDER BY b.created_at,b.id;
-- name: GetBrand :one
SELECT b.id::text AS id,b.name,b.description,b.website,b.status FROM brands b
JOIN user_brand_access g ON (g.org_id,g.brand_id)=(b.org_id,b.id)
JOIN organization_members m ON (m.org_id,m.user_id)=(g.org_id,g.user_id)
JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.org_id
WHERE b.org_id=sqlc.arg(org_id)::uuid AND b.id=sqlc.arg(id)::uuid AND g.user_id=sqlc.arg(user_id)::uuid AND g.active AND m.active AND u.active AND o.active;
-- name: InsertBrand :one
INSERT INTO brands(org_id,id,name,description,website) SELECT sqlc.arg(org_id)::uuid,sqlc.arg(id)::uuid,sqlc.arg(name),sqlc.arg(description),sqlc.arg(website)
FROM organization_members m JOIN user_roles r USING(org_id,user_id) JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.org_id
WHERE m.org_id=sqlc.arg(org_id)::uuid AND m.user_id=sqlc.arg(user_id)::uuid AND m.active AND u.active AND o.active AND EXISTS(SELECT 1 FROM role_permissions rp WHERE rp.role_id=r.role_id AND rp.permission_id='brands.create')
RETURNING id::text AS id,name,description,website,status;
-- name: UpdateBrand :one
UPDATE brands b SET name=sqlc.arg(name),description=sqlc.arg(description),website=sqlc.arg(website),status=sqlc.arg(status),updated_at=now()
WHERE b.org_id=sqlc.arg(org_id)::uuid AND b.id=sqlc.arg(id)::uuid AND EXISTS(SELECT 1 FROM user_brand_access g JOIN organization_members m ON (m.org_id,m.user_id)=(g.org_id,g.user_id) JOIN user_roles r ON (r.org_id,r.user_id)=(m.org_id,m.user_id) JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.org_id WHERE (g.org_id,g.brand_id)=(b.org_id,b.id) AND g.user_id=sqlc.arg(user_id)::uuid AND g.active AND m.active AND u.active AND o.active AND EXISTS(SELECT 1 FROM role_permissions rp WHERE rp.role_id=r.role_id AND rp.permission_id='brands.write'))
RETURNING id::text AS id,name,description,website,status;
