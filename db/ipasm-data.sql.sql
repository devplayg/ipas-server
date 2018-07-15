use ipasm;

truncate table sys_config;
INSERT INTO `sys_config` (`section`, `keyword`, `value_s`, `value_n`, `created`, `updated`) VALUES
    ('login', 'max_failed_login_attempts', '', 3, now(), now()),
    ('login', 'failure_block_time', '', 3, now(), now()),
    ('system', 'data_retention_days', '', 180, now(), now())
;




insert into ast_server(category1, category2, data_type, hostname, port, name)
values(1, 1, 1, '127.0.0.1', 0, 'localhost');
