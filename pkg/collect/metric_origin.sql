-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `metric`
(
    `id`             BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `name`           VARCHAR(255) NOT NULL COMMENT '名称',
    `help`           TEXT NOT NULL COMMENT '帮助信息',
    `updated_index`  BIGINT(20) NOT NULL COMMENT '更新索引',
    `metric`         LONGTEXT NOT NULL COMMENT '指标内容',
    `created_at`     DATETIME(3) NOT NULL DEFAULT current_timestamp (3) COMMENT '创建时间',
    `updated_at`     DATETIME(3) NOT NULL DEFAULT current_timestamp (3) ON UPDATE current_timestamp (3) COMMENT '更新时间',
    `deleted_at`     DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `idx_name_deleted_at` (`name`, `deleted_at`) USING BTREE,
    INDEX            `idx_deleted_at` (`deleted_at`) USING BTREE
) COMMENT ='监控指标' COLLATE = 'utf8_unicode_ci'
                       ENGINE = InnoDB;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS `metric`;

