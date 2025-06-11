WITH tenant AS
         (SELECT id FROM ncse_sys_tenant LIMIT 1),
     user_ids AS
       (SELECT id, username
        FROM ncse_sys_user),
     taxonomy_ids AS
       (SELECT id, name
        FROM ncse_cms_taxonomy)
INSERT
INTO ncse_cms_topic (id, name, title, slug, content, thumbnail, temp, markdown, private, status,
                     released, taxonomy_id, space_id, created_by, updated_by, created_at, updated_at)
VALUES
  -- 技术相关的话题
  (nanoid(), '人工智能简介', '人工智能：改变世界的技术', 'ai-introduction',
   '人工智能是计算机科学的一个分支，致力于创造智能机器。本文将介绍 AI 的基础概念、发展历史和未来展望。', 'ai_thumb.jpg',
   false, true, false, 1, EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM taxonomy_ids WHERE name = '人工智能'), (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), 'Python 编程入门', 'Python 编程：从零开始', 'python-basics',
   'Python 是一种简单易学 yet 功能强大的编程语言。本文将带你了解 Python 的基本语法、数据类型和控制结构。',
   'python_thumb.jpg', false, true, false, 1, EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM taxonomy_ids WHERE name = '编程'), (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 科学相关的话题
  (nanoid(), '大数据分析技术', '大数据分析：洞察数据的艺术', 'big-data-analysis',
   '大数据分析涉及检查、清理、转换和建模数据以发现有用信息。本文将介绍常用的大数据分析工具和技术。', 'big_data_thumb.jpg',
   false, true, false, 1, EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM taxonomy_ids WHERE name = '大数据'), (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 项目管理相关的话题
  (nanoid(), '敏捷开发方法论', '敏捷开发：提高项目效率的新方法', 'agile-methodology',
   '敏捷开发是一种以人为核心、迭代、递增的开发方法。本文将详细介绍 Scrum、Kanban 等敏捷方法。', 'agile_thumb.jpg', false,
   true, false, 1, EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM taxonomy_ids WHERE name = '项目管理'), (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 营销相关的话题
  (nanoid(), '数字营销策略', '数字时代的营销革命', 'digital-marketing-strategy',
   '数字营销利用数字技术和平台来推广产品和服务。本文将探讨 SEO、社交媒体营销、内容营销等数字营销策略。',
   'digital_marketing_thumb.jpg', false, true, false, 1, EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM taxonomy_ids WHERE name = '营销'), (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 草稿状态的话题
  (nanoid(), '机器学习算法概述', '机器学习算法：从理论到实践', 'machine-learning-algorithms',
   '本文将介绍常见的机器学习算法，包括监督学习、无监督学习和强化学习算法。', 'ml_algorithms_thumb.jpg', false, true, false,
   0, NULL, (SELECT id FROM taxonomy_ids WHERE name = '机器学习'), (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000));
