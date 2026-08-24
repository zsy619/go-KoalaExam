-- ============================================================
-- KoalaExam 数据库初始化（ke_ 前缀）
-- 数据库：KoalaExam
-- 用户：root / 123456
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS ke_department;
DROP TABLE IF EXISTS ke_class;
DROP TABLE IF EXISTS ke_user;
DROP TABLE IF EXISTS ke_question_category;
DROP TABLE IF EXISTS ke_question;
DROP TABLE IF EXISTS ke_paper;
DROP TABLE IF EXISTS ke_paper_question;
DROP TABLE IF EXISTS ke_exam;
DROP TABLE IF EXISTS ke_exam_record;
DROP TABLE IF EXISTS ke_favorite_folder;
DROP TABLE IF EXISTS ke_favorite;
DROP TABLE IF EXISTS ke_wrong_log;

CREATE TABLE IF NOT EXISTS ke_department (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL COMMENT '名称',
  parent_id BIGINT NOT NULL DEFAULT 0 COMMENT '父ID',
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织/院系';

CREATE TABLE IF NOT EXISTS ke_class (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  grade VARCHAR(32),
  department_id BIGINT,
  teacher_id BIGINT,
  student_cnt INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_dept (department_id),
  INDEX idx_teacher (teacher_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班级';

CREATE TABLE IF NOT EXISTS ke_user (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password VARCHAR(128) NOT NULL,
  nickname VARCHAR(64),
  email VARCHAR(128),
  phone VARCHAR(32),
  avatar VARCHAR(255),
  role TINYINT NOT NULL DEFAULT 3 COMMENT '1:超管 2:教师 3:学生',
  gender TINYINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '0:禁用 1:正常',
  class_id BIGINT,
  department_id BIGINT,
  last_login_at DATETIME,
  last_login_ip VARCHAR(64),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE INDEX uniq_username (username),
  INDEX idx_email (email),
  INDEX idx_phone (phone),
  INDEX idx_role (role),
  INDEX idx_status (status),
  INDEX idx_class (class_id),
  INDEX idx_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户';

CREATE TABLE IF NOT EXISTS ke_question_category (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  parent_id BIGINT NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64),
  sort INT NOT NULL DEFAULT 0,
  creator_id BIGINT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_parent (parent_id),
  INDEX idx_creator (creator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题库分类';

CREATE TABLE IF NOT EXISTS ke_question (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  category_id BIGINT NOT NULL,
  type TINYINT NOT NULL COMMENT '1:单选 2:多选 3:判断 4:填空 5:不定项 6:编程',
  difficulty TINYINT NOT NULL DEFAULT 2 COMMENT '1:易 2:中 3:难',
  title TEXT NOT NULL,
  options JSON,
  answer TEXT NOT NULL,
  analysis TEXT,
  tags VARCHAR(255),
  score DOUBLE NOT NULL DEFAULT 1,
  creator_id BIGINT,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '0:草稿 1:已发布 2:归档',
  use_count BIGINT NOT NULL DEFAULT 0,
  correct_count BIGINT NOT NULL DEFAULT 0,
  wrong_count BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_category (category_id),
  INDEX idx_type (type),
  INDEX idx_difficulty (difficulty),
  INDEX idx_creator (creator_id),
  INDEX idx_status (status),
  INDEX idx_tags (tags)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题目';

CREATE TABLE IF NOT EXISTS ke_paper (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(128) NOT NULL,
  description TEXT,
  strategy TINYINT NOT NULL DEFAULT 1 COMMENT '1:固定 2:随机 3:GA',
  total_score DOUBLE NOT NULL DEFAULT 100,
  duration INT NOT NULL DEFAULT 60,
  pass_score DOUBLE NOT NULL DEFAULT 60,
  status TINYINT NOT NULL DEFAULT 1,
  creator_id BIGINT,
  config_rule JSON,
  question_ids JSON,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_creator (creator_id),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='试卷';

CREATE TABLE IF NOT EXISTS ke_paper_question (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  paper_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  type TINYINT,
  score DOUBLE NOT NULL DEFAULT 1,
  sort INT NOT NULL DEFAULT 0,
  section VARCHAR(32),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE INDEX uniq_paper_q (paper_id, question_id),
  INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='试卷题目关联';

CREATE TABLE IF NOT EXISTS ke_exam (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(128) NOT NULL,
  description TEXT,
  paper_id BIGINT NOT NULL,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  duration INT NOT NULL DEFAULT 60,
  max_attempts INT NOT NULL DEFAULT 1,
  shuffle_q TINYINT(1) NOT NULL DEFAULT 1,
  shuffle_opt TINYINT(1) NOT NULL DEFAULT 1,
  anti_cheat TINYINT(1) NOT NULL DEFAULT 1,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '0:未发布 1:进行中 2:已结束',
  creator_id BIGINT,
  target_users TEXT,
  target_classes TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_paper (paper_id),
  INDEX idx_status (status),
  INDEX idx_start (start_time),
  INDEX idx_end (end_time),
  INDEX idx_creator (creator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='考试';

CREATE TABLE IF NOT EXISTS ke_exam_record (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  exam_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  paper_snapshot JSON,
  answers JSON,
  status TINYINT NOT NULL DEFAULT 0 COMMENT '0:进行中 1:已交卷 2:已批改 3:异常',
  start_time DATETIME NOT NULL,
  submit_time DATETIME,
  duration INT NOT NULL DEFAULT 0,
  total_score DOUBLE NOT NULL DEFAULT 0,
  objective_score DOUBLE NOT NULL DEFAULT 0,
  subjective_score DOUBLE NOT NULL DEFAULT 0,
  passed TINYINT(1) NOT NULL DEFAULT 0,
  score_hash VARCHAR(128),
  tab_switch_cnt INT NOT NULL DEFAULT 0,
  audit_log TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE INDEX uniq_exam_user (exam_id, user_id, deleted_at),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='考试记录';

CREATE TABLE IF NOT EXISTS ke_favorite_folder (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  is_system TINYINT(1) NOT NULL DEFAULT 0,
  color VARCHAR(16),
  icon VARCHAR(64),
  question_cnt INT NOT NULL DEFAULT 0,
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收藏夹';

CREATE TABLE IF NOT EXISTS ke_favorite (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  target_type TINYINT NOT NULL COMMENT '1:题目 2:试卷 3:知识点',
  target_id BIGINT NOT NULL,
  folder_id BIGINT,
  source_type TINYINT NOT NULL DEFAULT 1 COMMENT '1:主动 2:错题自动 3:推荐',
  difficulty TINYINT NOT NULL DEFAULT 2,
  note VARCHAR(500),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE INDEX uniq_user_target (user_id, target_type, target_id, deleted_at),
  INDEX idx_folder (folder_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收藏';

CREATE TABLE IF NOT EXISTS ke_wrong_log (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  exam_id BIGINT,
  exam_record_id BIGINT,
  user_answer TEXT,
  correct_answer TEXT,
  wrong_count INT NOT NULL DEFAULT 1,
  last_wrong_at DATETIME,
  is_reviewed TINYINT(1) NOT NULL DEFAULT 0,
  mastery_level TINYINT NOT NULL DEFAULT 1 COMMENT '1-5',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  INDEX idx_user_q (user_id, question_id),
  INDEX idx_exam (exam_id),
  INDEX idx_exam_record (exam_record_id),
  INDEX idx_last (last_wrong_at),
  INDEX idx_mastery (mastery_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='错题日志';

SET FOREIGN_KEY_CHECKS = 1;
