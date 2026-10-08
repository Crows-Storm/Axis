use axis_auth;

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
(1,  'READ',    'View',     0, 1),
(2,  'CREATE',  'Add',      1, 2),
(3,  'UPDATE',  'Edit',     2, 3),
(4,  'DELETE',  'Delete',   3, 4),
(5,  'ENABLE',  'Activate', 4, 5),
(6,  'CLOSE',   'Close',    5, 6),
(7,  'DISABLE', 'Disable',  6, 7),
(8,  'EXPORT',  'Export',   7, 8),
(9,  'AUDIT',   'Audit',    8, 9),
(10, 'PUBLISH', 'Publish',  9, 10);

INSERT INTO `sys_menu` (`id`, `parent_id`, `code`, `name`, `path`, `icon`, `sort`, `status`) VALUES
(1, 0, 'user',   'User Management',   '/users',  'user',   1, 1),
(2, 0, 'role',   'Role Management',   '/roles',  'role',   2, 1),
(3, 0, 'menu',   'Menu Management',   '/menus',  'menu',   3, 1),
(4, 0, 'audit',  'Audit Compliance',  '/audit',  'audit',  4, 1),
(5, 0, 'account', 'Account Management', '/accounts','account', 5, 1);

INSERT INTO `sys_role` (`id`, `code`, `name`, `description`, `status`) VALUES
(1, 'super_admin', 'Super Administrator', 'Full access to every menu and action', 1);

-- 1023 = READ|CREATE|UPDATE|DELETE|ENABLE|CLOSE|DISABLE|EXPORT|AUDIT|PUBLISH
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`, `mask`) VALUES
(1, 1, 1023),
(1, 2, 1023),
(1, 3, 1023),
(1, 4, 1023),
(1, 5, 1023);

-- user_id is the user service primary key
INSERT INTO `sys_user_role` (`user_id`, `role_id`) VALUES
(1, 1);
