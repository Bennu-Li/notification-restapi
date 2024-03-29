CREATE TABLE IF NOT EXISTS `message_template` (
	`id`          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
	`name`        VARCHAR(20) NOT NULL unique,
	`message`     TEXT NOT NULL,
	`registrant`  VARCHAR(20) NOT NULL,
	`application` VARCHAR(20) NOT NULL,
	`created_at`  DATETIME DEFAULT CURRENT_TIMESTAMP,
	`updated_at`  DATETIME DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;