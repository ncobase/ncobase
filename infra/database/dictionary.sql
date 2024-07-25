WITH tenant AS
         (SELECT id FROM ncse_tenant LIMIT 1),
     user_ids AS
         (SELECT id, username FROM ncse_user)
INSERT
INTO ncse_dictionary (id, name, slug, type, value, tenant_id, created_by, created_at, updated_by, updated_at)
VALUES
  -- 系统状态
  (nanoid(), '通用状态', 'common_status', 'object', '{"0":"禁用","1":"启用"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '审核状态', 'audit_status', 'object', '{"0":"待审核","1":"审核通过","2":"审核拒绝"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '发布状态', 'publish_status', 'object', '{"0":"草稿","1":"已发布","2":"已撤回","3":"已归档"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 优先级
  (nanoid(), '通用优先级', 'common_priority', 'object', '{"1":"低","2":"中","3":"高","4":"紧急"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '任务优先级', 'task_priority', 'object', '{"1":"P0","2":"P1","3":"P2","4":"P3"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 项目管理
  (nanoid(), '项目类型', 'project_type', 'object',
   '{"1":"研发项目","2":"运维项目","3":"咨询项目","4":"实施项目","5":"运营项目"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '项目状态', 'project_status', 'object',
   '{"1":"规划中","2":"进行中","3":"已完成","4":"已暂停","5":"已取消"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '任务状态', 'task_status', 'object',
   '{"1":"待开始","2":"进行中","3":"已完成","4":"已暂停","5":"已取消","6":"待审核"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '任务类型', 'task_type', 'object', '{"1":"需求","2":"设计","3":"开发","4":"测试","5":"部署","6":"文档"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 用户和权限
  (nanoid(), '用户角色', 'user_role', 'object',
   '{"admin":"系统管理员","pm":"项目经理","dev":"开发人员","qa":"测试人员","ops":"运维人员","user":"普通用户"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '权限类型', 'permission_type', 'object', '{"1":"菜单","2":"按钮","3":"接口"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 文档管理
  (nanoid(), '文档类型', 'document_type', 'object',
   '{"1":"需求文档","2":"设计文档","3":"开发文档","4":"测试文档","5":"运维文档","6":"用户手册"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '文档状态', 'document_status', 'object',
   '{"1":"草稿","2":"审核中","3":"已发布","4":"已归档","5":"已废弃"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 问题跟踪
  (nanoid(), '问题类型', 'issue_type', 'object', '{"1":"缺陷","2":"改进","3":"新功能","4":"任务","5":"安全漏洞"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '问题严重度', 'issue_severity', 'object', '{"1":"致命","2":"严重","3":"一般","4":"轻微","5":"建议"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '问题状态', 'issue_status', 'object',
   '{"1":"待处理","2":"处理中","3":"已解决","4":"已关闭","5":"重新打开","6":"无法重现"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 资源管理
  (nanoid(), '资源类型', 'resource_type', 'object',
   '{"1":"服务器","2":"数据库","3":"存储","4":"网络设备","5":"软件许可"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '资源状态', 'resource_status', 'object', '{"1":"空闲","2":"使用中","3":"维护中","4":"已报废"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 知识库
  (nanoid(), '知识类型', 'knowledge_type', 'object',
   '{"1":"产品知识","2":"技术文档","3":"常见问题","4":"最佳实践","5":"故障案例"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '知识状态', 'knowledge_status', 'object',
   '{"1":"草稿","2":"待审核","3":"已发布","4":"已归档","5":"已废弃"}', (SELECT id FROM tenant),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000),
   (SELECT id FROM user_ids WHERE username = 'admin'), EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 配置管理
  (nanoid(), '配置类型', 'config_type', 'object', '{"1":"系统配置","2":"业务配置","3":"安全配置","4":"性能配置"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '配置格式', 'config_format', 'object', '{"1":"键值对","2":"JSON","3":"YAML","4":"XML"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 通知和消息
  (nanoid(), '通知类型', 'notification_type', 'object', '{"1":"系统通知","2":"任务提醒","3":"审批通知","4":"安全警告"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '通知渠道', 'notification_channel', 'object', '{"1":"站内信","2":"邮件","3":"短信","4":"微信","5":"钉钉"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),

  -- 审计日志
  (nanoid(), '审计类型', 'audit_type', 'object', '{"1":"登录日志","2":"操作日志","3":"安全日志","4":"系统日志"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000)),
  (nanoid(), '审计结果', 'audit_result', 'object', '{"1":"成功","2":"失败","3":"警告","4":"异常"}',
   (SELECT id FROM tenant), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000), (SELECT id FROM user_ids WHERE username = 'admin'),
   EXTRACT(EPOCH FROM now()) * 1000 + floor(random() * 1000));
