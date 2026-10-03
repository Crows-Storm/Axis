-- =====================================================
-- 1. 用户表
-- =====================================================
CREATE TABLE `sys_user` (
                            `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
                            `username`    VARCHAR(64)     NOT NULL COMMENT '用户名',
                            `password`    VARCHAR(128)    NOT NULL COMMENT '密码哈希(如 bcrypt)',
                            `nickname`    VARCHAR(64)     DEFAULT NULL COMMENT '昵称',
                            `email`       VARCHAR(128)    DEFAULT NULL COMMENT '邮箱',
                            `status`      TINYINT         NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
                            `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- =====================================================
-- 2. 角色表
-- =====================================================
CREATE TABLE `sys_role` (
                            `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
                            `code`        VARCHAR(64)     NOT NULL COMMENT '角色编码(如 admin, editor)',
                            `name`        VARCHAR(64)     NOT NULL COMMENT '角色名称',
                            `description` VARCHAR(255)    DEFAULT NULL COMMENT '描述',
                            `status`      TINYINT         NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
                            `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- =====================================================
-- 3. 菜单/资源表（每个菜单是一个资源）
-- =====================================================
CREATE TABLE `sys_menu` (
                            `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
                            `parent_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父菜单ID',
                            `code`        VARCHAR(64)     NOT NULL COMMENT '菜单编码(如 user, system, product, activity)',
                            `name`        VARCHAR(64)     NOT NULL COMMENT '菜单名称',
                            `path`        VARCHAR(128)    DEFAULT NULL COMMENT '前端路由',
                            `icon`        VARCHAR(64)     DEFAULT NULL COMMENT '图标',
                            `sort`        INT             NOT NULL DEFAULT 0 COMMENT '排序',
                            `status`      TINYINT         NOT NULL DEFAULT 1 COMMENT '1显示 0隐藏',
                            `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `uk_code` (`code`),
                            KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单/资源表';

-- =====================================================
-- 4. 角色-菜单权限表（核心表：用 bitmask 存权限）
-- =====================================================
-- mask 的含义：对某个菜单的操作位掩码
-- 例：mask=127 表示 READ|CREATE|UPDATE|DELETE|ENABLE|CLOSE|DISABLE
CREATE TABLE `sys_role_menu` (
                                 `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                 `role_id`     BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
                                 `menu_id`     BIGINT UNSIGNED NOT NULL COMMENT '菜单ID',
                                 `mask`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '权限位掩码',
                                 `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                 `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_role_menu` (`role_id`, `menu_id`),
                                 KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色-菜单权限表';

-- =====================================================
-- 5. 用户-角色表（一个用户可多角色）
-- =====================================================
CREATE TABLE `sys_user_role` (
                                 `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                 `user_id`     BIGINT UNSIGNED NOT NULL,
                                 `role_id`     BIGINT UNSIGNED NOT NULL,
                                 `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
                                 KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-角色表';

-- =====================================================
-- 6. 操作位定义表（可选，便于前端展示和扩展）
-- =====================================================
CREATE TABLE `sys_action` (
                              `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                              `code`        VARCHAR(32)     NOT NULL COMMENT '操作编码(如 READ, CREATE)',
                              `name`        VARCHAR(32)     NOT NULL COMMENT '操作名称',
                              `bit`         TINYINT         NOT NULL COMMENT '位序号(0~63)',
                              `sort`        INT             NOT NULL DEFAULT 0,
                              PRIMARY KEY (`id`),
                              UNIQUE KEY `uk_code` (`code`),
                              UNIQUE KEY `uk_bit` (`bit`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作位定义表';

-- =====================================================
-- 7. Seed data (bit = 0..9 → mask = 1<<bit)
-- READ=1 CREATE=2 UPDATE=4 DELETE=8 ENABLE=16 CLOSE=32
-- DISABLE=64 EXPORT=128 AUDIT=256 PUBLISH=512
-- PermAll = 1023
-- =====================================================
INSERT INTO `sys_action` (`id`, `code`, `name`, `bit`, `sort`) VALUES
(1,  'READ',    '查看', 0, 1),
(2,  'CREATE',  '新增', 1, 2),
(3,  'UPDATE',  '修改', 2, 3),
(4,  'DELETE',  '删除', 3, 4),
(5,  'ENABLE',  '激活', 4, 5),
(6,  'CLOSE',   '关闭', 5, 6),
(7,  'DISABLE', '禁用', 6, 7),
(8,  'EXPORT',  '导出', 7, 8),
(9,  'AUDIT',   '审核', 8, 9),
(10, 'PUBLISH', '发布', 9, 10);

INSERT INTO `sys_menu` (`id`, `parent_id`, `code`, `name`, `path`, `icon`, `sort`, `status`) VALUES
(1, 0, 'user',     '用户管理', '/user',     'user',     1, 1),
(2, 0, 'system',   '系统设置', '/system',   'setting',  2, 1),
(3, 0, 'product',  '商品管理', '/product',  'goods',    3, 1),
(4, 0, 'activity', '活动管理', '/activity', 'activity', 4, 1);

INSERT INTO `sys_role` (`id`, `code`, `name`, `description`, `status`) VALUES
(1, 'admin',  '超级管理员', '全部菜单全部操作', 1),
(2, 'editor', '内容编辑',   '商品读写+上下架，用户只读', 1),
(3, 'guest',  '访客',       '只读', 1);

-- admin: 1023 = 前 10 位全开
-- editor product: 1|2|4|16|32 = 55
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`, `mask`) VALUES
(1, 1, 1023),
(1, 2, 1023),
(1, 3, 1023),
(1, 4, 1023),
(2, 1, 1),
(2, 3, 55),
(3, 1, 1),
(3, 3, 1);

-- user_id 对应 user 服务的用户主键
INSERT INTO `sys_user_role` (`user_id`, `role_id`) VALUES
(1, 1),
(2, 2);
