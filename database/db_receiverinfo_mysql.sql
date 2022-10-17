-- DROP TABLE IF EXISTS `receiver_info`;
CREATE TABLE IF NOT EXISTS `receiver_info` (
  `receiverid`   char(15) NOT NULL,
  `receiver`     VARCHAR(30) ,
  `chatid`       VARCHAR(50) ,
  `time`         DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`receiverid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;