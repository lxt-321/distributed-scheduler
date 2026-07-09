-- 分布式任务调度平台 初始化脚本
CREATE DATABASE IF NOT EXISTS xxl_job DEFAULT CHARACTER SET utf8mb4;
USE xxl_job;

CREATE TABLE `xxl_job_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `app_name` varchar(64) NOT NULL COMMENT '执行器AppName',
  `title` varchar(64) NOT NULL COMMENT '执行器名称',
  `address_type` tinyint(4) NOT NULL DEFAULT 0 COMMENT '0=自动注册,1=手动录入',
  `address_list` varchar(512) DEFAULT NULL COMMENT '手动录入地址,逗号分隔',
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `xxl_job_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `job_group` int(11) NOT NULL COMMENT '执行器分组ID',
  `job_cron` varchar(128) NOT NULL COMMENT '任务Cron表达式',
  `job_desc` varchar(255) NOT NULL COMMENT '任务描述',
  `author` varchar(64) DEFAULT NULL,
  `executor_route_strategy` varchar(50) DEFAULT 'FIRST' COMMENT '路由策略',
  `executor_handler` varchar(255) DEFAULT NULL COMMENT '任务处理器名',
  `executor_param` varchar(512) DEFAULT NULL COMMENT '任务参数',
  `executor_block_strategy` varchar(50) DEFAULT 'SERIAL_EXECUTION' COMMENT '阻塞策略',
  `executor_timeout` int(11) DEFAULT 0 COMMENT '任务超时(秒,0=不限制)',
  `executor_fail_retry_count` int(11) DEFAULT 0 COMMENT '失败重试次数',
  `trigger_status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '0=停止,1=运行',
  `trigger_last_time` bigint(13) NOT NULL DEFAULT 0,
  `trigger_next_time` bigint(13) NOT NULL DEFAULT 0,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `xxl_job_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `job_group` int(11) NOT NULL,
  `job_id` int(11) NOT NULL,
  `executor_address` varchar(255) DEFAULT NULL,
  `executor_handler` varchar(255) DEFAULT NULL,
  `executor_param` varchar(512) DEFAULT NULL,
  `executor_sharding_param` varchar(255) DEFAULT NULL COMMENT '分片参数',
  `executor_fail_retry_count` int(11) DEFAULT 0,
  `trigger_time` bigint(13) DEFAULT NULL,
  `trigger_code` int(11) DEFAULT NULL,
  `trigger_msg` varchar(2000) DEFAULT NULL,
  `handle_time` bigint(13) DEFAULT NULL,
  `handle_code` int(11) DEFAULT NULL,
  `handle_msg` varchar(2000) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `i_trigger_time` (`trigger_time`),
  KEY `i_handle_code` (`handle_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 种子数据：演示执行器分组 + 一个每5秒触发的示例任务
INSERT INTO `xxl_job_group`(`app_name`,`title`,`address_type`,`update_time`)
  VALUES ('dscheduler-demo','演示执行器',0,NOW());

INSERT INTO `xxl_job_info`(
  `job_group`,`job_cron`,`job_desc`,`author`,
  `executor_route_strategy`,`executor_handler`,`executor_param`,
  `executor_block_strategy`,`trigger_status`,`update_time`)
  VALUES (
  1,'*/5 * * * * ?','演示任务-每5秒执行一次','admin',
  'FIRST','demoJob','hello world',
  'SERIAL_EXECUTION',1,NOW());
