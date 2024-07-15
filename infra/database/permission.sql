-- Account related permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_account', 'GET', '/v1/account', 'Read account information', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_account', 'PUT', '/v1/account', 'Update account information', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'change_password', 'PUT', '/v1/account/password', 'Change account password', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_account_tenant', 'GET', '/v1/account/tenant', 'Read account tenant information', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_account_tenants', 'GET', '/v1/account/tenants', 'Read all account tenants', false, false, '{}', null, null, NOW(), NOW());

-- Assets permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_assets', 'GET', '/v1/assets', 'Read all assets', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_asset', 'POST', '/v1/assets', 'Create a new asset', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_asset', 'GET', '/v1/assets/{slug}', 'Read a specific asset', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_asset', 'PUT', '/v1/assets/{slug}', 'Update a specific asset', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_asset', 'DELETE', '/v1/assets/{slug}', 'Delete a specific asset', false, false, '{}', null, null, NOW(), NOW());

-- Authorization permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'send_authorization', 'POST', '/v1/authorize/send', 'Send authorization', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'validate_authorization', 'POST', '/v1/authorize/{code}', 'Validate authorization code', false, false, '{}', null, null, NOW(), NOW());

-- Captcha permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'generate_captcha', 'POST', '/v1/captcha/generate', 'Generate a new captcha', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'validate_captcha', 'POST', '/v1/captcha/validate', 'Validate a captcha', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_captcha', 'GET', '/v1/captcha/{captcha_id}', 'Read a specific captcha', false, false, '{}', null, null, NOW(), NOW());

-- Authentication permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'login', 'POST', '/v1/login', 'User login', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'logout', 'POST', '/v1/logout', 'User logout', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'register', 'POST', '/v1/register', 'User registration', false, false, '{}', null, null, NOW(), NOW());

-- Menus permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_menus', 'GET', '/v1/menus', 'Read all menus', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_menu', 'POST', '/v1/menus', 'Create a new menu', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_menu', 'GET', '/v1/menus/{slug}', 'Read a specific menu', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_menu', 'PUT', '/v1/menus/{slug}', 'Update a specific menu', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_menu', 'DELETE', '/v1/menus/{slug}', 'Delete a specific menu', false, false, '{}', null, null, NOW(), NOW());

-- Permissions management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_permissions', 'GET', '/v1/permissions', 'Read all permissions', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_permission', 'POST', '/v1/permissions', 'Create a new permission', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_permission', 'GET', '/v1/permissions/{slug}', 'Read a specific permission', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_permission', 'PUT', '/v1/permissions/{slug}', 'Update a specific permission', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_permission', 'DELETE', '/v1/permissions/{slug}', 'Delete a specific permission', false, false, '{}', null, null, NOW(), NOW());

-- Policies management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_policies', 'GET', '/v1/policies', 'Read all policies', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_policy', 'POST', '/v1/policies', 'Create a new policy', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_policy', 'GET', '/v1/policies/{id}', 'Read a specific policy', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_policy', 'PUT', '/v1/policies/{id}', 'Update a specific policy', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_policy', 'DELETE', '/v1/policies/{id}', 'Delete a specific policy', false, false, '{}', null, null, NOW(), NOW());

-- Roles management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_roles', 'GET', '/v1/roles', 'Read all roles', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_role', 'POST', '/v1/roles', 'Create a new role', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_role', 'GET', '/v1/roles/{slug}', 'Read a specific role', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_role', 'PUT', '/v1/roles/{slug}', 'Update a specific role', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_role', 'DELETE', '/v1/roles/{slug}', 'Delete a specific role', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_role_permissions', 'POST', '/v1/roles/{slug}/permissions', 'Manage permissions for a specific role', false, false, '{}', null, null, NOW(), NOW());

-- Taxonomies management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_taxonomies', 'GET', '/v1/taxonomies', 'Read all taxonomies', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_taxonomy', 'POST', '/v1/taxonomies', 'Create a new taxonomy', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_taxonomy', 'GET', '/v1/taxonomies/{slug}', 'Read a specific taxonomy', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_taxonomy', 'PUT', '/v1/taxonomies/{slug}', 'Update a specific taxonomy', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_taxonomy', 'DELETE', '/v1/taxonomies/{slug}', 'Delete a specific taxonomy', false, false, '{}', null, null, NOW(), NOW());

-- Tenants management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_tenants', 'GET', '/v1/tenants', 'Read all tenants', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_tenant', 'POST', '/v1/tenants', 'Create a new tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_tenant', 'GET', '/v1/tenants/{slug}', 'Read a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_tenant', 'PUT', '/v1/tenants/{slug}', 'Update a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_tenant', 'DELETE', '/v1/tenants/{slug}', 'Delete a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_assets', 'POST', '/v1/tenants/{slug}/assets', 'Manage assets for a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_groups', 'POST', '/v1/tenants/{slug}/groups', 'Manage groups for a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_menu', 'POST', '/v1/tenants/{slug}/menu', 'Manage menu for a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_roles', 'POST', '/v1/tenants/{slug}/roles', 'Manage roles for a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_setting', 'POST', '/v1/tenants/{slug}/setting', 'Manage setting for a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_settings', 'POST', '/v1/tenants/{slug}/settings', 'Manage settings for a specific tenant', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenant_users', 'POST', '/v1/tenants/{slug}/users', 'Manage users for a specific tenant', false, false, '{}', null, null, NOW(), NOW());

-- Topics management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'read_topics', 'GET', '/v1/topics', 'Read all topics', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'create_topic', 'POST', '/v1/topics', 'Create a new topic', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_topic', 'GET', '/v1/topics/{slug}', 'Read a specific topic', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_topic', 'PUT', '/v1/topics/{slug}', 'Update a specific topic', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_topic', 'DELETE', '/v1/topics/{slug}', 'Delete a specific topic', false, false, '{}', null, null, NOW(), NOW());

-- Users management
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'update_user', 'PUT', '/v1/users/{username}', 'Update a specific user', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_user', 'DELETE', '/v1/users/{username}', 'Delete a specific user', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_user_tenant', 'POST', '/v1/users/{username}/tenant', 'Manage tenant for a specific user', false, false, '{}', null, null, NOW(), NOW());

-- general permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'all', '*', '*', 'Full access to all resources', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_all', 'GET', '*', 'Read access to all resources', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'write_all', 'POST', '*', 'Write access to all resources', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'update_all', 'PUT', '*', 'Update access to all resources', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'delete_all', 'DELETE', '*', 'Delete access to all resources', false, false, '{}', null, null, NOW(), NOW());

-- Tenant-specific permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'manage_tenant_all', '*', '/v1/tenants/{slug}/*', 'Full access to all tenant-specific resources', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'read_tenant_all', 'GET', '/v1/tenants/{slug}/*', 'Read access to all tenant-specific resources', false, false, '{}', null, null, NOW(), NOW());

-- API version permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'access_v1_api', '*', '/v1/*', 'Full access to all v1 API endpoints', false, false, '{}', null, null, NOW(), NOW());

-- Specific endpoint group permissions
INSERT INTO public.nb_permission (id, name, action, subject, description, "default", disabled, extras, created_by, updated_by, created_at, updated_at)
VALUES
  (nanoid(), 'manage_account', '*', '/v1/account*', 'Full access to account-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_assets', '*', '/v1/assets*', 'Full access to asset-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_authorization', '*', '/v1/authorize*', 'Full access to authorization-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_captcha', '*', '/v1/captcha*', 'Full access to captcha-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_menus', '*', '/v1/menus*', 'Full access to menu-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_permissions', '*', '/v1/permissions*', 'Full access to permission-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_policies', '*', '/v1/policies*', 'Full access to policy-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_roles', '*', '/v1/roles*', 'Full access to role-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_taxonomies', '*', '/v1/taxonomies*', 'Full access to taxonomy-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_tenants', '*', '/v1/tenants*', 'Full access to tenant-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_topics', '*', '/v1/topics*', 'Full access to topic-related endpoints', false, false, '{}', null, null, NOW(), NOW()),
  (nanoid(), 'manage_users', '*', '/v1/users*', 'Full access to user-related endpoints', false, false, '{}', null, null, NOW(), NOW());
