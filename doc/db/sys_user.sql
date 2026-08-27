CREATE TABLE sys_user (
                          id BIGINT UNSIGNED PRIMARY KEY COMMENT 'user unique id',
                          login_id VARCHAR(50) NOT NULL COMMENT 'user login id',
                          username VARCHAR(50) NOT NULL COMMENT 'user name / nick name',
                          password VARCHAR(255) NOT NULL COMMENT 'password',
                          email VARCHAR(100),
                          status TINYINT DEFAULT 1 COMMENT '状态: 1正常 0禁用',
                          deleted TINYINT DEFAULT 0 COMMENT '逻辑删除: 0未删除 1已删除',
                          create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
                          update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                          UNIQUE KEY uk_username_deleted (username, deleted),
                          KEY idx_status (status)
) COMMENT='user basic info table';