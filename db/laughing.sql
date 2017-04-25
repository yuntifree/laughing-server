use laughing;

-- 用户信息
CREATE TABLE IF NOT EXISTS users 
(
    uid     bigint unsigned NOT NULL AUTO_INCREMENT,
    token   varchar(32) NOT NULL,
    -- facebook id 和 token
    fb_id   varchar(32) NOT NULL,
    fb_token    varchar(128) NOT NULL,
    nickname    varchar(256) NOT NULL,
    headurl varchar(256) NOT NULL,
    imei    varchar(36) NOT NULL,
    -- 粉丝数
    fan_cnt int unsigned NOT NULL DEFAULT 0,
    -- 关注数
    follow_cnt  int unsigned NOT NULL DEFAULT 0,
    -- 分享数
    videos      int unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    mtime   datetime NOT NULL DEFAULT '2017-01-01',
    etime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(uid),
    KEY(fb_id)
) ENGINE = InnoDB;

-- 粉丝
CREATE TABLE IF NOT EXISTS fan 
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    uid     int unsigned NOT NULL,
    tuid    int unsigned NOT NULL,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    mtime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id),
    UNIQUE KEY(uid, tuid)
) ENGINE = InnoDB;

-- 关注
CREATE TABLE IF NOT EXISTS follower 
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    uid     int unsigned NOT NULL,
    tuid    int unsigned NOT NULL,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    mtime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id),
    UNIQUE KEY(uid, tuid)
) ENGINE = InnoDB;

-- 视频信息
CREATE TABLE IF NOT EXISTS media 
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    uid     int unsigned NOT NULL,
    img     varchar(128) NOT NULL,
    dst     varchar(128) NOT NULL, 
    title   varchar(256) NOT NULL,
    abstract    varchar(256) NOT NULL,
    views   int unsigned NOT NULL DEFAULT 0,
    -- origin 来源 0:上传 1:外部link
    origin  tinyint unsigned NOT NULL DEFAULT 0,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    unshare tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id)
) ENGINE = InnoDB;

-- 标签
CREATE TABLE IF NOT EXISTS tags
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    content varchar(64) NOT NULL,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id)
) ENGINE = InnoDB;

-- 视频标签
CREATE TABLE IF NOT EXISTS media_tags
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    mid     bigint unsigned NOT NULL,
    tid     int unsigned NOT NULL,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id),
    UNIQUE KEY(mid, tid)
) ENGINE = InnoDB;

-- 分享
CREATE TABLE IF NOT EXISTS shares
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    uid     int unsigned NOT NULL,
    -- medias id
    mid     int unsigned NOT NULL,
    allowshare  tinyint unsigned NOT NULL DEFAULT 0,
    reshare int unsigned NOT NULL DEFAULT 0,
    comments    int unsigned NOT NULL DEFAULT 0,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id),
    KEY(uid),
    KEY(mid)
) ENGINE = InnoDB;

-- 评论
CREATE TABLE IF NOT EXISTS comments
(
    id     bigint unsigned NOT NULL AUTO_INCREMENT,
    -- shares id
    sid     int unsigned NOT NULL,
    uid     int unsigned NOT NULL,
    content varchar(512) NOT NULL,
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(id),
    KEY(sid),
    KEY(uid)
) ENGINE = InnoDB;




