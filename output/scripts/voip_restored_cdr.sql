/*
 Navicat Premium Data Transfer

 Source Server         : 192.168.1.205
 Source Server Type    : MySQL
 Source Server Version : 50729
 Source Host           : 192.168.1.205:3306
 Source Schema         : centnet_cdrrs

 Target Server Type    : MySQL
 Target Server Version : 50729
 File Encoding         : 65001

 Date: 27/11/2020 18:19:16
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for voip_restored_cdr
-- ----------------------------
DROP TABLE IF EXISTS `voip_restored_cdr`;
CREATE TABLE `voip_restored_cdr` (
  `id` bigint(20) NOT NULL COMMENT '话单唯一ID',
  `call_id` varchar(128) DEFAULT NULL COMMENT '通话ID',
  `caller_ip` varchar(64) DEFAULT NULL COMMENT '主叫IP',
  `caller_port` int(8) DEFAULT NULL COMMENT '主叫端口',
  `callee_ip` varchar(64) DEFAULT NULL COMMENT '被叫IP',
  `callee_port` int(8) DEFAULT NULL COMMENT '被叫端口',
  `caller_num` varchar(64) DEFAULT NULL COMMENT '主叫号码',
  `callee_num` varchar(64) DEFAULT NULL COMMENT '被叫号码',
  `caller_device` varchar(128) DEFAULT NULL COMMENT '主叫设备名',
  `callee_device` varchar(128) DEFAULT NULL COMMENT '被叫设备名',
  `callee_province` varchar(64) DEFAULT NULL COMMENT '被叫所属省',
  `callee_city` varchar(64) DEFAULT NULL COMMENT '被叫所属市',
  `connect_time` int(11) DEFAULT NULL COMMENT '通话开始时间',
  `disconnect_time` int(11) DEFAULT NULL COMMENT '通话结束时间',
  `duration` int(8) DEFAULT '0' COMMENT '通话时长',
  `fraud_type` varchar(32) DEFAULT NULL COMMENT '诈骗类型',
  `create_time` datetime DEFAULT NULL COMMENT '生成时间',
  PRIMARY KEY (`id`),
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

SET FOREIGN_KEY_CHECKS = 1;
