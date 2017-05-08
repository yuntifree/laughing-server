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


