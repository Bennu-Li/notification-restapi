CREATE TABLE IF NOT EXISTS `user_behavior` (
  `id`          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user`        VARCHAR(30) ,
  `application` VARCHAR(30) ,
  `template`    TEXT,
  `params`      TEXT,
  `message`     TEXT,
  `status`      INT(10),
  `send_time`   DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;