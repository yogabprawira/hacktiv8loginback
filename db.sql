CREATE DATABASE login;

USE login;

CREATE TABLE `userinfo` (
    `uid` INT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NULL DEFAULT NULL,
    `name` VARCHAR(64) NULL DEFAULT NULL,
    `email` VARCHAR(64) NULL DEFAULT NULL,
    `password` VARCHAR(64) NULL DEFAULT NULL,
    `role` VARCHAR(64) NULL DEFAULT NULL,
    PRIMARY KEY (`uid`)
);

