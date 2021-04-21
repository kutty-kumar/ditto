USE pikachu;

CREATE TABLE `users`
(
  `external_id`   varchar(100) DEFAULT NULL,
  `id`            bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at`    datetime     DEFAULT NULL,
  `updated_at`    datetime     DEFAULT NULL,
  `deleted_at`    datetime     DEFAULT NULL,
  `status`        int(11)      DEFAULT NULL,
  `first_name`    varchar(255) DEFAULT NULL,
  `last_name`     varchar(255) DEFAULT NULL,
  `gender`        int(11)      DEFAULT NULL,
  `date_of_birth` datetime     DEFAULT NULL,
  `age`           bigint(20)   DEFAULT NULL,
  `height`        double       DEFAULT NULL,
  `weight`        double       DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_users_external_id` (`external_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `identities`
(
  `external_id`    varchar(100) DEFAULT NULL,
  `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at`     datetime     DEFAULT NULL,
  `updated_at`     datetime     DEFAULT NULL,
  `deleted_at`     datetime     DEFAULT NULL,
  `status`         int(11)      DEFAULT NULL,
  `identity_type`  int(11)      DEFAULT NULL,
  `identity_value` varchar(255) DEFAULT NULL,
  `user_id`        varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_identities_external_id` (`external_id`),
  KEY `identities_user_id_users_external_id_foreign` (`user_id`),
  CONSTRAINT `identities_user_id_users_external_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`external_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `addresses`
(
  `external_id` varchar(100) DEFAULT NULL,
  `id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at`  datetime     DEFAULT NULL,
  `updated_at`  datetime     DEFAULT NULL,
  `deleted_at`  datetime     DEFAULT NULL,
  `status`      int(11)      DEFAULT NULL,
  `line1`       varchar(255) DEFAULT NULL,
  `line2`       varchar(255) DEFAULT NULL,
  `zip_code`    varchar(255) DEFAULT NULL,
  `state`       varchar(255) DEFAULT NULL,
  `country`     varchar(255) DEFAULT NULL,
  `user_id`     varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_addresses_external_id` (`external_id`),
  KEY `addresses_user_id_users_external_id_foreign` (`user_id`),
  CONSTRAINT `addresses_user_id_users_external_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`external_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;