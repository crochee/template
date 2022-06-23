-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `dcs_author_control`
(
    `id`             BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `account_id`     VARCHAR(255) NOT NULL COMMENT '主账号ID',
    `author_control` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '权限控制：0不控制，1到期受限控制，2销毁受限控制',
    `deleted`        BIGINT(20) UNSIGNED NOT NULL COMMENT '软删除记录id',
    `created_at`     DATETIME(3) NOT NULL DEFAULT current_timestamp (3) COMMENT '创建时间',
    `updated_at`     DATETIME(3) NOT NULL DEFAULT current_timestamp (3) ON UPDATE current_timestamp (3) COMMENT '更新时间',
    `deleted_at`     DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `idx_account_id_deleted` (`account_id`, `deleted`) USING BTREE,
    INDEX            `idx_dcs_author_control_deleted_at` (`deleted_at`) USING BTREE
) COMMENT ='分布式云权限控制表' COLLATE = 'utf8_unicode_ci'
                       ENGINE = InnoDB;

CREATE TABLE `dcs_resource_pkg`
(
    `id`            BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `resource_id`   VARCHAR(255) NOT NULL COMMENT '资源id，按照一定的业务规则，生成唯一标识',
    `charge_type`   TINYINT(3) UNSIGNED NOT NULL COMMENT '资源包开通类型，1包年包月，2按量计费',
    `account_id`    VARCHAR(255) NOT NULL COMMENT '主账号ID',
    `user_id`       VARCHAR(255) NOT NULL COMMENT '用户ID',
    `product_id`    VARCHAR(100) NULL DEFAULT NULL COMMENT '产品编号',
    `pkg_status`    TINYINT(3) UNSIGNED NOT NULL COMMENT '资源包开通状态 0开通失败,1开通成功,2过期',
    `order_id`      VARCHAR(255) NOT NULL COMMENT '工单唯一标识',
    `order_type`    VARCHAR(255) NOT NULL COMMENT '工单类型,CREATE代表创建,RENEW代表续订,UPGRADE代表升配,DOWNGRADE代表降配,DESTROY代表销毁',
    `resource_type` BIGINT(20) NOT NULL COMMENT '工单资源类型，5784代表边缘云开通唯一标识',
    `configuration` LONGTEXT     NOT NULL COMMENT '工单资源构建json',
    `active_time`   DATETIME(3) NULL DEFAULT NULL COMMENT '生效时间',
    `inactive_time` DATETIME(3) NULL DEFAULT NULL COMMENT '失效时间',
    `deleted`       BIGINT(20) UNSIGNED NOT NULL COMMENT '软删除记录id',
    `created_at`    DATETIME(3) NOT NULL DEFAULT current_timestamp (3) COMMENT '创建时间',
    `updated_at`    DATETIME(3) NOT NULL DEFAULT current_timestamp (3) ON UPDATE current_timestamp (3) COMMENT '更新时间',
    `deleted_at`    DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `idx_resource_id_deleted` (`resource_id`, `deleted`) USING BTREE,
    INDEX           `idx_account_id` (`account_id`) USING BTREE,
    INDEX           `idx_dcs_resource_pkg_deleted_at` (`deleted_at`) USING BTREE,
    CONSTRAINT `configuration` CHECK (json_valid(`configuration`))
) COMMENT ='分布式云实例资源包表' COLLATE = 'utf8_unicode_ci'
                        ENGINE = InnoDB;

CREATE TABLE `dcs_resource_change_flow`
(
    `id`              BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `resource_id`     VARCHAR(255) NOT NULL COMMENT '资源id，按照一定的业务规则，生成唯一标识',
    `order_type`      VARCHAR(255) NOT NULL COMMENT '工单类型,CREATE代表创建,RENEW代表续订,UPGRADE代表升配,DOWNGRADE代表降配,DESTROY代表销毁',
    `purchase_unit`   BIGINT(20) NOT NULL COMMENT '购买单位',
    `purchase_number` BIGINT(20) NOT NULL COMMENT '购买数量',
    `reason`          LONGTEXT NULL DEFAULT NULL COMMENT '资源包开通失败原因',
    `created_at`      DATETIME(3) NOT NULL DEFAULT current_timestamp (3) COMMENT '创建时间',
    `updated_at`      DATETIME(3) NOT NULL DEFAULT current_timestamp (3) ON UPDATE current_timestamp (3) COMMENT '更新时间',
    `deleted_at`      DATETIME(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    INDEX             `idx_resource_id` (`resource_id`) USING BTREE,
    INDEX             `idx_dcs_resource_change_flow_deleted_at` (`deleted_at`) USING BTREE
) COMMENT ='分布式云实例资源变更事件表' COLLATE = 'utf8_unicode_ci'
                           ENGINE = InnoDB;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS `dcs_author_control`;
DROP TABLE IF EXISTS `dcs_resource_pkg`;
DROP TABLE IF EXISTS `dcs_resource_change_flow`;
