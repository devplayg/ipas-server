

create user 'devplayg'@'%' identified by 'devplayg';
grant all privileges on ipasm.* to 'devplayg'@'%';
create user 'devplayg'@'localhost' identified by 'devplayg123';
grant all privileges on ipasm.* to 'devplayg'@'localhost';
flush privileges;




use ipasm;

select * from log_ipas_event;
select * from log_ipas_status;
select * from ast_ipas;
select * from ast_asset;

truncate table log_ipas_event;
truncate table log_ipas_status;
truncate table ast_ipas;





INSERT INTO `mbr_member` (`member_id`, `org_id`, `username`, `email`, `position`, `name`, `birth`, `nickname`, `zipcode`, `country`, `state`, `city`, `address1`, `address2`, `phone1`, `phone2`, `login_count`, `status`, `timezone`, `failed_login_count`, `last_success_login`, `last_failed_login`, `last_read_message`, `session_id`, `created`, `last_updated`) VALUES
 (1, 'root', 'kws', 'kws@kyungwoo.com', 1024, 'Kyungwoo admin.', '1970-01-01', '관리자', '93105', 'kr', 'Guro', 'Seoul', '', '', '010-1234-1234', '010-3456-3456', 244, 1, 'Asia/Seoul', 0, '2018-04-07 15:51:12', '2018-03-22 16:40:22', 0, '03af8bde2bb07cf78f3720a7175f0707', '2017-02-28 14:55:40', '2018-04-07 15:51:12')
,(2, '', 'kws_kr', 'help_kr@kyungwoo.com', 1, '경우-KR', '1970-01-01', '', '', '', '', '', '', '', '', '', 0, 0, 'Asia/Seoul', 0, '1970-01-01 00:00:00', '1970-01-01 00:00:00', 0, '', '2018-04-07 15:53:26', '2018-04-07 15:55:25')
,(3, '', 'kws_en', 'help_en@kyungwoo.com', 1, '경우-EN', '1970-01-01', '', '', '', '', '', '', '', '', '', 0, 0, '', 0, '1970-01-01 00:00:00', '1970-01-01 00:00:00', 0, '', '2018-04-07 15:54:00', '2018-04-07 15:54:00')
,(4, '', 'kws_jp', 'help_jp@kyungwoo.com', 1, '경우-JP', '1970-01-01', '', '', '', '', '', '', '', '', '', 0, 0, '', 0, '1970-01-01 00:00:00', '1970-01-01 00:00:00', 0, '', '2018-04-07 15:55:53', '2018-04-07 15:55:53')
;


INSERT INTO `mbr_password` (`member_id`, `password`, `salt`, `created`, `updated`) VALUES
  (1, 'ee6e4493f3becc79853337c41517f08340b9e1d3eab30f0f3ec5cce81066c098', 'GINMw@au%f', '2017-02-20 10:21:28', '2018-04-07 15:51:12')
, (2, 'bb521feb6156a56b54194ea67096e5ccbddb846055f8ead62a549435c6dd2ad4', '', '2018-04-07 15:53:26', '2018-04-07 15:53:26')
, (3, '59dc8a04edc727387c84dbd5dc648c77b2a54d1d5b77b0b04c88bd085cbd6e70', '', '2018-04-07 15:54:00', '2018-04-07 15:54:00')
, (4, '8c20854fb44d3b5c9250c0507c316f284193eab235c46e14537e60cfd441d155', '', '2018-04-07 15:55:53', '2018-04-07 15:55:53')
;


INSERT INTO `ast_asset` (`asset_id`, `class`, `parent_id`, `name`, `type1`, `type2`, `seq`, `created`, `updated`) VALUES
    (1,     1,  0, '한국 #1',   1, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (2,     1,  0, '한국 #2',   1, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (3,     1,  0, 'USA #1',    1, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (4,     1,  0, 'USA #2',    1, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (5,     1,  0, '日本 #1',   1, 1, 0, '2018-03-23 16:35:24', '2018-04-01 15:11:46'),
    (6,     1,  0, '日本 #2',   1, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (7,     1,  1, '서울',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-25 10:30:24'),
    (8,     1,  1, '경기',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-25 10:30:26'),
    (9,     1,  2, '충청',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-25 10:30:28'),
    (10,    1,  2, '경상',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-25 10:30:31'),
    (11,    1,  3, 'CA',        2, 1, 0, '2018-03-23 16:35:24', '2018-04-01 15:11:37'),
    (12,    1,  3, 'WA',        2, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (13,    1,  4, 'AZ',        2, 1, 0, '2018-03-23 16:35:24', '2018-03-28 23:40:01'),
    (14,    1,  4, 'SC',        2, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (15,    1,  5, '東北',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (16,    1,  5, '関東',      2, 2, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (17,    1,  6, '主婦',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
    (18,    1,  6, '関西',      2, 1, 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24')
;

INSERT INTO `ast_asset` (`asset_id`, `class`, `parent_id`, `name`, `type1`, `type2`, `icon`, `seq`, `created`, `updated`) VALUES
 (1, 1, 0, 'Korea #1', 1, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (2, 1, 0, 'Korea #2', 1, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (3, 1, 0, 'USA West', 1, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (4, 1, 0, 'USA East', 1, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (5, 1, 0, '日本 #1', 1, 1, '', 0, '2018-03-23 16:35:24', '2018-04-01 15:11:46'),
 (6, 1, 0, '日本 #2', 1, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (7, 1, 1, 'Seoul', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-25 10:30:24'),
 (8, 1, 1, 'Busan', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-25 10:30:26'),
 (9, 1, 2, 'Pyeongchang', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-25 10:30:28'),
 (10, 1, 2, 'Bundang', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-25 10:30:31'),
 (11, 1, 3, 'Seatle', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-04-01 15:11:37'),
 (12, 1, 3, 'Los Angeles', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (13, 1, 4, 'New York', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-28 23:40:01'),
 (14, 1, 4, 'Washington D.C.', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-04-22 18:39:03'),
 (15, 1, 5, '東北', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (16, 1, 5, '関東', 2, 2, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (17, 1, 6, '主婦', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24'),
 (18, 1, 6, '関西', 2, 1, '', 0, '2018-03-23 16:35:24', '2018-03-23 16:35:24')
;

insert into ast_code(asset_id, code) values(1, 'kr1'),(2, 'kr2'),(3, 'us1'),(4, 'us2'),(5, 'jp1'),(6, 'jp2');





update log_ipas_event
set date = date_add(date, interval datediff(now(), date) day)
where date >= '2017-03-17 00:00:00';




CREATE TABLE `stats_equip_count` (
  `date` datetime NOT NULL,
  `org_id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `equip_type` int(11) NOT NULL COMMENT 'vt, zt, pt',
  `count` int(10) unsigned NOT NULL,
  KEY `ix_date` (`date`),
  KEY `ix_orgid` (`date`,`org_id`),
  KEY `ix_groupid` (`date`,`org_id`, `group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE IF NOT EXISTS stats_timeline (
			date      datetime NOT NULL,
			asset_id  int(11) NOT NULL,
			time      datetime NOT NULL,
			summary varchar(32) not null,

			KEY       ix_date (date),
			KEY       ix_assetid (date, asset_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;