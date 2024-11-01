-- 获取租户
WITH tenant AS
         (SELECT id FROM ncse_tenant LIMIT 1),
     user_ids AS
         (SELECT id, username FROM ncse_user),
     -- 插入 天工集团
     group_1 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '邺都集团', 'yedu-group', NULL, (SELECT id FROM tenant),
                 (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id),
     -- 插入 信息科技公司 和 文化传媒公司
     company_1 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '信息科技公司', 'tech-company', (SELECT id FROM group_1), (SELECT id FROM tenant),
                 (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '文化传媒公司', 'media-company', (SELECT id FROM group_1), (SELECT id FROM tenant),
                 (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id, slug),
     -- 插入 信息科技公司的部门（技术部、市场部、财务部）
     department_1 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '技术部', 'tech-department', (SELECT id FROM company_1 WHERE slug = 'tech-company'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '市场部', 'marketing-department', (SELECT id FROM company_1 WHERE slug = 'tech-company'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '财务部', 'finance-department', (SELECT id FROM company_1 WHERE slug = 'tech-company'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id, slug),
     -- 插入 技术部的组（研发组、运维组、测试组）
     group_2 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '研发组', 'r&d-team', (SELECT id FROM department_1 WHERE slug = 'tech-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '运维组', 'operations-team', (SELECT id FROM department_1 WHERE slug = 'tech-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '测试组', 'qa-team', (SELECT id FROM department_1 WHERE slug = 'tech-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))),
     -- 插入 市场部的组（策划组、推广组、调研组）
     group_3 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '策划组', 'planning-team', (SELECT id FROM department_1 WHERE slug = 'marketing-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '推广组', 'promotion-team',
                 (SELECT id FROM department_1 WHERE slug = 'marketing-department'), (SELECT id FROM tenant),
                 (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '调研组', 'research-team', (SELECT id FROM department_1 WHERE slug = 'marketing-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))),
     -- 插入 财务部的组（会计组、审计组、税务组）
     group_4 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '会计组', 'accounting-team', (SELECT id FROM department_1 WHERE slug = 'finance-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '审计组', 'audit-team', (SELECT id FROM department_1 WHERE slug = 'finance-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '税务组', 'tax-team', (SELECT id FROM department_1 WHERE slug = 'finance-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))),
     -- 插入 文化传媒公司的部门（内容部、广告部、法务部）
     department_2 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '内容部', 'content-department', (SELECT id FROM company_1 WHERE slug = 'media-company'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '广告部', 'ad-department', (SELECT id FROM company_1 WHERE slug = 'media-company'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '法务部', 'legal-department', (SELECT id FROM company_1 WHERE slug = 'media-company'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id, slug),
     -- 插入 内容部的组（编辑组、审核组、发布组）
     group_5 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '编辑组', 'editor-team', (SELECT id FROM department_2 WHERE slug = 'content-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '审核组', 'review-team', (SELECT id FROM department_2 WHERE slug = 'content-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '发布组', 'publish-team', (SELECT id FROM department_2 WHERE slug = 'content-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))),
     -- 插入 广告部的组（策划组、执行组、分析组）
     group_6 AS (
       INSERT INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES (nanoid(), '策划组', 'ad-planning-team', (SELECT id FROM department_2 WHERE slug = 'ad-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '执行组', 'ad-execution-team', (SELECT id FROM department_2 WHERE slug = 'ad-department'),
                 (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
                 (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
                (nanoid(), '分析组', 'ad-analysis-team', (SELECT id
                                                          FROM department_2

                                                          WHERE slug = 'ad-department'), (SELECT id FROM tenant),
                 (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
                 EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)))
-- 插入 法务部的组（合同组、诉讼组、合规组）
INSERT
INTO ncse_group (id, name, slug, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
VALUES (nanoid(), '合同组', 'contract-team', (SELECT id FROM department_2 WHERE slug = 'legal-department'),
        (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
        (SELECT id FROM user_ids WHERE username = 'admin'),
        EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
        EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
       (nanoid(), '诉讼组', 'litigation-team', (SELECT id FROM department_2 WHERE slug = 'legal-department'),
        (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
        (SELECT id FROM user_ids WHERE username = 'admin'),
        EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
        EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
       (nanoid(), '合规组', 'compliance-team', (SELECT id FROM department_2 WHERE slug = 'legal-department'),
        (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
        (SELECT id FROM user_ids WHERE username = 'admin'),
        EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
        EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000));
