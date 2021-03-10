/*
 Navicat MariaDB Data Transfer

 Source Server         : 本地
 Source Server Type    : MariaDB
 Source Server Version : 50568
 Source Host           : localhost:3306
 Source Schema         : scit_edu_tool

 Target Server Type    : MariaDB
 Target Server Version : 50568
 File Encoding         : 65001

 Date: 09/03/2021 22:09:39
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for class_chart
-- ----------------------------
DROP TABLE IF EXISTS `class_chart`;
CREATE TABLE `class_chart`  (
  `f_id` smallint(6) NOT NULL,
  `s_id` smallint(6) NOT NULL,
  `c_id` tinyint(4) NOT NULL,
  `c_name` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  UNIQUE INDEX `class_chart`(`c_name`, `f_id`, `s_id`, `c_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for class_schedule
-- ----------------------------
DROP TABLE IF EXISTS `class_schedule`;
CREATE TABLE `class_schedule`  (
  `t_id` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `t_faculty` smallint(6) NOT NULL,
  `t_specialty` smallint(6) NOT NULL,
  `t_class` tinyint(4) NOT NULL,
  `t_grade` smallint(6) NOT NULL,
  `t_school_year` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `t_semester` tinyint(4) NOT NULL,
  `t_content` text CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `t_expired` int(10) UNSIGNED NOT NULL,
  UNIQUE INDEX `class_schedule`(`t_id`, `t_faculty`, `t_specialty`, `t_class`, `t_grade`, `t_school_year`, `t_semester`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for faculty_chart
-- ----------------------------
DROP TABLE IF EXISTS `faculty_chart`;
CREATE TABLE `faculty_chart`  (
  `f_id` smallint(6) NOT NULL,
  `f_name` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  UNIQUE INDEX `faculty_chart`(`f_id`, `f_name`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for sign_keys
-- ----------------------------
DROP TABLE IF EXISTS `sign_keys`;
CREATE TABLE `sign_keys`  (
  `app_key` tinytext CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `app_secret` tinytext CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `platform` tinytext CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `mail` tinytext CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `build` int(10) NOT NULL,
  `available` tinyint(1) NULL DEFAULT 1
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for specialty_chart
-- ----------------------------
DROP TABLE IF EXISTS `specialty_chart`;
CREATE TABLE `specialty_chart`  (
  `s_id` smallint(6) NOT NULL,
  `s_name` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `f_id` smallint(6) NOT NULL,
  INDEX `specialty_chart`(`s_name`, `s_id`, `f_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for student_achieve
-- ----------------------------
DROP TABLE IF EXISTS `student_achieve`;
CREATE TABLE `student_achieve`  (
  `u_id` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `u_faculty` smallint(6) NOT NULL,
  `u_specialty` smallint(6) NOT NULL,
  `u_class` tinyint(4) NOT NULL,
  `u_grade` smallint(6) NOT NULL,
  `a_content_01` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_02` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_03` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_04` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_05` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_06` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_07` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_08` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_09` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_10` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_11` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `a_content_12` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  PRIMARY KEY (`u_id`) USING BTREE,
  INDEX `student_achieve`(`u_id`, `u_faculty`, `u_specialty`, `u_class`, `u_grade`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for user_info
-- ----------------------------
DROP TABLE IF EXISTS `user_info`;
CREATE TABLE `user_info`  (
  `u_id` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `u_name` tinytext CHARACTER SET utf8 COLLATE utf8_general_ci NULL,
  `u_identify` tinyint(2) NOT NULL DEFAULT 0,
  `u_level` tinyint(2) NOT NULL DEFAULT 0,
  `u_faculty` smallint(6) NULL DEFAULT NULL,
  `u_specialty` smallint(6) NULL DEFAULT NULL,
  `u_class` tinyint(4) NULL DEFAULT NULL,
  `u_grade` smallint(6) NULL DEFAULT NULL,
  `u_info_expired` int(11) NOT NULL,
  PRIMARY KEY (`u_id`) USING BTREE,
  UNIQUE INDEX `user_info`(`u_id`, `u_identify`, `u_faculty`, `u_specialty`, `u_class`, `u_grade`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

-- ----------------------------
-- Table structure for user_token
-- ----------------------------
DROP TABLE IF EXISTS `user_token`;
CREATE TABLE `user_token`  (
  `u_id` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `u_password` varchar(600) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `u_session` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `u_session_expired` int(11) NOT NULL,
  `u_token_effective` tinyint(1) NOT NULL,
  PRIMARY KEY (`u_id`) USING BTREE,
  UNIQUE INDEX `user_token`(`u_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;

SET FOREIGN_KEY_CHECKS = 1;
