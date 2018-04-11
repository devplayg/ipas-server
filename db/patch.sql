use ipasm;


drop table ast_asset;
CREATE TABLE `ast_asset` (
  `asset_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `class` int(10) unsigned NOT NULL COMMENT '1:IPAS-Org, 2: reserved',
  `parent_id` int(10) unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `code` varchar(16) not null default '',
  `type1` int(10) unsigned NOT NULL,
  `type2` int(10) unsigned NOT NULL DEFAULT '0',
  `seq` int(10) unsigned NOT NULL DEFAULT '0',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`asset_id`),
  KEY `ix_parent_id` (`parent_id`),
  KEY `ix_code` (`code`),
  KEY `ix_class` (`class`),
  KEY `ix_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `ast_ipas` (
  `equip_id` varchar(16) NOT NULL,
  `equip_type` int(11) NOT NULL COMMENT 'vt, zt, pt',
  `org_id` int(10) unsigned NOT NULL DEFAULT '0',
  `group_id` int(10) unsigned NOT NULL DEFAULT '0',
  `latitude` float(10,6) NOT NULL DEFAULT '0.000000',
  `longitude` float(10,6) NOT NULL DEFAULT '0.000000',
  `speed` int(11) NOT NULL DEFAULT '0',
  `snr` int(11) NOT NULL DEFAULT '0',
  `usim` varchar(32) NOT NULL DEFAULT '',
  `speeding_count` int(11) NOT NULL DEFAULT '0',
  `shock_count` int(11) NOT NULL DEFAULT '0',
  `ip` int(10) unsigned NOT NULL DEFAULT '0',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`equip_id`),

  KEY `ix_ast_ipas_equiptype` (`equip_type`),
  KEY `ix_ast_ipas_orgid` (`org_id`),
  KEY `ix_ast_ipas_groupid` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8