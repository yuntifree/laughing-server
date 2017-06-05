use laughing;

CREATE TABLE IF NOT EXISTS back_login
(
    uid     int unsigned NOT NULL AUTO_INCREMENT,
    username    varchar(32) NOT NULL,
    passwd      varchar(32) NOT NULL,
    salt        varchar(32) NOT NULL,
    skey        varchar(32) NOT NULL,
    role        tinyint unsigned NOT NULL DEFAULT 0,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    PRIMARY KEY(uid),
    UNIQUE KEY(username)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS user_lang
(
    id      int unsigned NOT NULL AUTO_INCREMENT,
    lang    varchar(16) NOT NULL DEFAULT 'zh',
    content varchar(128) NOT NULL DEFAULT '',
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    PRIMARY KEY(id),
    UNIQUE KEY(lang)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS lang_follower
(
    id      int unsigned NOT NULL AUTO_INCREMENT,
    lid     int unsigned NOT NULL,
    uid     int unsigned NOT NULL,
    ctime   datetime NOT NULL DEFAULT '2017-01-01',
    deleted tinyint unsigned NOT NULL DEFAULT 0,
    PRIMARY KEY(id),
    UNIQUE KEY(lid, uid)
) ENGINE = InnoDB;


