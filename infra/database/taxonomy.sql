-- 插入分类数据
WITH tenant AS
         (SELECT id FROM ncse_tenant LIMIT 1),
     user_ids AS
         (SELECT id, username FROM ncse_user),
     taxonomy_data_1 AS (
       INSERT INTO ncse_taxonomy (id, name, type, slug, cover, thumbnail, color, icon, url, keywords, description,
                                  status,
                                  extras, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES
           -- 技术
           (nanoid(), '技术', 'node', 'jishu', 'cover_tech.jpg', 'thumb_tech.jpg', '#0000FF', 'icon_tech.png',
            'https://ncobase.com/tech', '技术，创新', '与技术相关的所有内容。', 0, '{"extra_info": "技术类别"}', NULL,
            (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
            (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 科学
           (nanoid(), '科学', 'node', 'kexue', 'cover_sci.jpg', 'thumb_sci.jpg', '#00FF00', 'icon_sci.png',
            'https://ncobase.com/science', '科学，研究', '科学主题和研究。', 0, '{"extra_info": "科学类别"}', NULL,
            (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
            (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 项目管理
           (nanoid(), '项目管理', 'node', 'xiangmuguanli', 'cover_pm.jpg', 'thumb_pm.jpg', '#FFB6C1', 'icon_pm.png',
            'https://ncobase.com/projectmanagement', '项目管理，流程', '项目管理和流程。', 0,
            '{"extra_info": "项目管理类别"}', NULL,
            (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
            (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 营销
           (nanoid(), '营销', 'node', 'yingxiao', 'cover_marketing.jpg', 'thumb_marketing.jpg', '#FFD700',
            'icon_marketing.png',
            'https://ncobase.com/marketing', '营销，策略', '营销和策略。', 0, '{"extra_info": "营销类别"}', NULL,
            (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
            (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id, name),
     taxonomy_data_2 AS (
       INSERT INTO ncse_taxonomy (id, name, type, slug, cover, thumbnail, color, icon, url, keywords, description,
                                  status,
                                  extras, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES
           -- 编程 (技术的子类别)
           (nanoid(), '编程', 'node', 'biancheng', 'cover_prog.jpg', 'thumb_prog.jpg', '#FF0000', 'icon_prog.png',
            'https://ncobase.com/programming', '编程，开发', '编程语言和开发。', 0, '{"extra_info": "编程子类别"}',
            (SELECT id FROM taxonomy_data_1 WHERE name = '技术'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 人工智能 (技术的子类别)
           (nanoid(), '人工智能', 'node', 'rengongzhineng', 'cover_ai.jpg', 'thumb_ai.jpg', '#FFAAAA', 'icon_ai.png',
            'https://ncobase.com/ai', '人工智能，智能', '人工智能技术和应用。', 0, '{"extra_info": "人工智能子类别"}',
            (SELECT id FROM taxonomy_data_1 WHERE name = '技术'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 数据科学 (科学的子类别)
           (nanoid(), '数据科学', 'node', 'shujukexue', 'cover_ds.jpg', 'thumb_ds.jpg', '#00FFFF', 'icon_ds.png',
            'https://ncobase.com/datascience', '数据科学，分析', '数据科学和分析。', 0,
            '{"extra_info": "数据科学子类别"}',
            (SELECT id FROM taxonomy_data_1 WHERE name = '科学'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 战略规划 (项目管理的子类别)
           (nanoid(), '战略规划', 'node', 'zhanlueguihua', 'cover_strategy.jpg', 'thumb_strategy.jpg', '#8A2BE2',
            'icon_strategy.png',
            'https://ncobase.com/strategy', '战略规划，发展', '战略规划与发展。', 0, '{"extra_info": "战略规划子类别"}',
            (SELECT id FROM taxonomy_data_1 WHERE name = '项目管理'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 广告 (营销的子类别)
           (nanoid(), '广告', 'node', 'guanggao', 'cover_ad.jpg', 'thumb_ad.jpg', '#FF6347', 'icon_ad.png',
            'https://ncobase.com/advertising', '广告，宣传', '广告与宣传。', 0, '{"extra_info": "广告子类别"}',
            (SELECT id FROM taxonomy_data_1 WHERE name = '营销'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 市场调研 (营销的子类别)
           (nanoid(), '市场调研', 'node', 'shichangdiaocha', 'cover_market_research.jpg', 'thumb_market_research.jpg',
            '#98FB98', 'icon_market_research.png',
            'https://ncobase.com/marketresearch', '市场调研，分析', '市场调研和分析。', 0,
            '{"extra_info": "市场调研子类别"}',
            (SELECT id FROM taxonomy_data_1 WHERE name = '营销'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id, name),
     taxonomy_data_3 AS (
       INSERT INTO ncse_taxonomy (id, name, type, slug, cover, thumbnail, color, icon, url, keywords, description,
                                  status,
                                  extras, parent_id, tenant_id, created_by, updated_by, created_at, updated_at)
         VALUES
           -- 机器学习 (人工智能的子类别)
           (nanoid(), '机器学习', 'node', 'jishixuexi', 'cover_ml.jpg', 'thumb_ml.jpg', '#FF00FF', 'icon_ml.png',
            'https://ncobase.com/ml', '机器学习，人工智能', '机器学习和人工智能。', 0, '{"extra_info": "机器学习子类别"}',
            (SELECT id FROM taxonomy_data_2 WHERE name = '人工智能'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
           -- 大数据 (数据科学的子类别)
           (nanoid(), '大数据', 'node', 'dadata', 'cover_bd.jpg', 'thumb_bd.jpg', '#AAFFAA', 'icon_bd.png',
            'https://ncobase.com/bigdata', '大数据，分析', '大数据技术和应用。', 0, '{"extra_info": "大数据子类别"}',
            (SELECT id FROM taxonomy_data_2 WHERE name = '数据科学'), (SELECT id FROM tenant),
            (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
            EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000))
         RETURNING id, name),
     taxonomy_ids AS (SELECT id, name
                      FROM taxonomy_data_1
                      UNION ALL
                      SELECT id, name
                      FROM taxonomy_data_2
                      UNION ALL
                      SELECT id, name
                      FROM taxonomy_data_3)
-- 插入分类关系数据
INSERT
INTO ncse_taxonomy_relation (id, object_id, taxonomy_id, type, "order", created_by, created_at)
VALUES
  -- 技术相关的分类
  (nanoid(), 'object1', (SELECT id FROM taxonomy_ids WHERE name = '技术'), 'topic', 1,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object2', (SELECT id FROM taxonomy_ids WHERE name = '编程'), 'topic', 2,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object3', (SELECT id FROM taxonomy_ids WHERE name = '人工智能'), 'topic', 3,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object4', (SELECT id FROM taxonomy_ids WHERE name = '机器学习'), 'topic', 4,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 科学相关的分类
  (nanoid(), 'object5', (SELECT id FROM taxonomy_ids WHERE name = '科学'), 'topic', 5,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object6', (SELECT id FROM taxonomy_ids WHERE name = '数据科学'), 'topic', 6,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object7', (SELECT id FROM taxonomy_ids WHERE name = '大数据'), 'topic', 7,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 项目管理相关的分类
  (nanoid(), 'object8', (SELECT id FROM taxonomy_ids WHERE name = '项目管理'), 'topic', 8,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object9', (SELECT id FROM taxonomy_ids WHERE name = '战略规划'), 'topic', 9,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 营销相关的分类
  (nanoid(), 'object10', (SELECT id FROM taxonomy_ids WHERE name = '营销'), 'topic', 10,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object11', (SELECT id FROM taxonomy_ids WHERE name = '广告'), 'topic', 11,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'object12', (SELECT id FROM taxonomy_ids WHERE name = '市场调研'), 'topic', 12,
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000));
