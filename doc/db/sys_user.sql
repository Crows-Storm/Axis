CREATE TABLE sys_user (
                          id BIGINT UNSIGNED PRIMARY KEY COMMENT 'user unique id',
                          login_id VARCHAR(50) NOT NULL COMMENT 'user login id',
                          username VARCHAR(50) NOT NULL COMMENT 'user name / nick name',
                          password VARCHAR(255) NOT NULL COMMENT 'password',
                          type VARCHAR(255) NOT NULL COMMENT 'android/ios/website/other(github/google/wx/...)',
                          email VARCHAR(100) NOT NULL ,
                          status TINYINT DEFAULT 1 COMMENT 'status: 0: pending 1: activated 2: disabled',
                          deleted TINYINT DEFAULT 0 COMMENT '逻辑删除: 0未删除 1已删除',
                          create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
                          update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                          UNIQUE KEY uk_login_id_deleted (login_id, deleted),
                          UNIQUE KEY uk_login_id (login_id),
                          UNIQUE KEY uk_email (email),
                          INDEX idx_deleted_status (deleted, status),
                          KEY idx_status (status)
) COMMENT='user basic info table';