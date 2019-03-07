use ipasm;

truncate table sys_config;
INSERT INTO `sys_config` (`section`, `keyword`, `value_s`, `value_n`, `created`, `updated`) VALUES
    ('login', 'max_failed_login_attempts', '', 3, now(), now()),
    ('login', 'failure_block_time', '', 3, now(), now()),
    ('system', 'data_retention_days', '', 180, now(), now())
;

insert into ast_server(category1, category2, data_type, hostname, port, name)
values(1, 1, 1, '127.0.0.1', 0, 'localhost');

INSERT INTO `mbr_member` (`member_id`, `org_id`, `username`, `email`, `position`, `name`, `birth`, `nickname`, `zipcode`, `country`, `state`, `city`, `address1`, `address2`, `phone1`, `phone2`, `login_count`, `status`, `timezone`, `failed_login_count`, `last_success_login`, `last_failed_login`, `last_read_message`, `session_id`, `created`, `last_updated`) VALUES
 (1, 'root', 'kws', 'kws@kyungwoo.com', 1024, 'Kyungwoo admin.', '1970-01-01', '관리자', '93105', 'kr', 'Guro', 'Seoul', '', '', '010-1234-1234', '010-3456-3456', 244, 1, 'Asia/Seoul', 0, '2018-04-07 15:51:12', '2018-03-22 16:40:22', 0, '03af8bde2bb07cf78f3720a7175f0707', '2017-02-28 14:55:40', '2018-04-07 15:51:12')
;

INSERT INTO `mbr_password` (`member_id`, `password`, `salt`, `created`, `updated`) VALUES
  (1, 'ee6e4493f3becc79853337c41517f08340b9e1d3eab30f0f3ec5cce81066c098', 'GINMw@au%f', '2017-02-20 10:21:28', '2018-04-07 15:51:12')
;

