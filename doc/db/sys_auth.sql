CREATE TABLE sys_role (
                          id BIGINT PRIMARY KEY AUTO_INCREMENT,
                          role_code VARCHAR(50) NOT NULL COMMENT 'role code: ADMIN, MANAGER, USER',
                          role_name VARCHAR(50) NOT NULL COMMENT '角色名称',
                          role_level INT DEFAULT 0 COMMENT '角色级别: 数值越小权限越大',
                          status TINYINT DEFAULT 1 COMMENT '状态: 1正常 0禁用',
                          deleted TINYINT DEFAULT 0,
                          create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
                          update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                          UNIQUE KEY uk_role_code_deleted (role_code, deleted),
                          KEY idx_status (status)
) COMMENT='角色表';


CREATE TABLE sys_menu (
                          id BIGINT PRIMARY KEY AUTO_INCREMENT,
                          parent_id BIGINT DEFAULT 0 COMMENT '父菜单ID，0表示顶级菜单',
                          menu_code VARCHAR(50) NOT NULL COMMENT '菜单编码，唯一标识',
                          menu_name VARCHAR(50) NOT NULL COMMENT '菜单名称',
                          menu_type TINYINT NOT NULL COMMENT '类型: 1目录 2菜单 3按钮',
                          path VARCHAR(200) COMMENT '路由路径',
                          tree_path VARCHAR(500) COMMENT '树路径: 1/2/5/',
                          component VARCHAR(200) COMMENT '组件路径',
                          icon VARCHAR(50) COMMENT '图标',
                          permission VARCHAR(100) COMMENT '权限标识: user:add, user:edit',
                          sort_order INT DEFAULT 0 COMMENT '排序号，升序',
                          visible TINYINT DEFAULT 1 COMMENT '是否显示: 0隐藏 1显示',
                          status TINYINT DEFAULT 1 COMMENT '状态: 1正常 0禁用',
                          deleted TINYINT DEFAULT 0,
                          create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
                          update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                          UNIQUE KEY uk_menu_code_deleted (menu_code, deleted),
                          KEY idx_parent_id (parent_id),
                          KEY idx_menu_type (menu_type),
                          KEY idx_sort_order (sort_order)
) COMMENT='菜单表';

-- 插入示例数据
INSERT INTO sys_menu VALUES
                         (1, 0, 'system', 'System Management', 1, '/system', NULL, 'Layout', 'system', 'system:index', 1, 1, 1, 0, NOW(), NOW()),
                         (2, 1, 'user', 'User Management', 2, '/user', NULL, 'system/user/index', 'user', 'system:user:list', 1, 1, 1, 0, NOW(), NOW()),
                         (3, 2, 'user_add', 'User Add', 3, NULL, '/2/3', NULL, NULL, 'system:user:add', 1, 1, 1, 0, NOW(), NOW()),
                         (4, 2, 'user_edit', 'User Edit', 3, NULL, '/2/4', NULL, NULL, 'system:user:edit', 2, 1, 1, 0, NOW(), NOW()),
                         (5, 2, 'user_delete', 'User Delete', 3, NULL, '/2/5', NULL, NULL, 'system:user:delete', 3, 1, 1, 0, NOW(), NOW()),
                         (6, 1, 'role', 'Role Management', 2, '/role', NULL, 'system/role/index', 'role', 'system:role:list', 2, 1, 1, 0, NOW(), NOW());

CREATE TABLE sys_user_role (
                               id BIGINT PRIMARY KEY AUTO_INCREMENT,
                               user_id BIGINT NOT NULL,
                               role_id BIGINT NOT NULL,
                               create_time DATETIME DEFAULT CURRENT_TIMESTAMP,

                               UNIQUE KEY uk_user_role (user_id, role_id),
                               KEY idx_user_id (user_id),
                               KEY idx_role_id (role_id)
) COMMENT='用户角色关联表'

CREATE TABLE sys_role_menu (
                               id BIGINT PRIMARY KEY AUTO_INCREMENT,
                               role_id BIGINT NOT NULL,
                               menu_id BIGINT NOT NULL,
                               create_time DATETIME DEFAULT CURRENT_TIMESTAMP,

                               UNIQUE KEY uk_role_menu (role_id, menu_id),
                               KEY idx_role_id (role_id),
                               KEY idx_menu_id (menu_id)
) COMMENT='角色菜单关联表';