USE `reconcile`;

CREATE TABLE `orders` (
    `id` BIGINT NOT NULL COMMENT '订单ID(Snowflake)',
    `tenant_id` VARCHAR(32) NOT NULL COMMENT '租户ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `order_no` VARCHAR(64) NOT NULL COMMENT '业务订单号',
    `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '总金额(分)',
    `status` VARCHAR(20) NOT NULL DEFAULT 'CREATED' COMMENT '订单状态: CREATED/PAID/FINISHED/FAILED',
    `idempotent_key` VARCHAR(128) NOT NULL COMMENT '幂等键',
    `trace_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '追踪ID',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tenant_idempotent` (`tenant_id`, `idempotent_key`),
    KEY `idx_tenant_user` (`tenant_id`, `user_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';

CREATE TABLE `order_items` (
    `id` BIGINT NOT NULL COMMENT '明细ID(Snowflake)',
    `tenant_id` VARCHAR(32) NOT NULL COMMENT '租户ID',
    `order_id` BIGINT NOT NULL COMMENT '订单ID',
    `item_code` VARCHAR(64) NOT NULL COMMENT '商品编码',
    `item_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商品名称',
    `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
    `unit_price` BIGINT NOT NULL DEFAULT 0 COMMENT '单价(分)',
    `total_price` BIGINT NOT NULL DEFAULT 0 COMMENT '总价(分)',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_tenant_item` (`tenant_id`, `item_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';

CREATE TABLE `outbox_messages` (
    `id` BIGINT NOT NULL COMMENT '消息ID(Snowflake)',
    `tenant_id` VARCHAR(32) NOT NULL COMMENT '租户ID',
    `biz_no` VARCHAR(64) NOT NULL COMMENT '业务编号(如订单号)',
    `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型: ASSET_GRANT / NOTICE_SEND',
    `status` VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '状态: PENDING/SENT/FAILED',
    `payload` JSON NOT NULL COMMENT '消息体(JSON)',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '已重试次数',
    `max_retries` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
    `next_retry_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '下次重试时间',
    `trace_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '追踪ID',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `sent_at` DATETIME(3) NULL COMMENT '发送成功时间',
    `last_retry_at` DATETIME(3) NULL COMMENT '最后重试时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_biz_event` (`tenant_id`, `biz_no`, `event_type`),
    KEY `idx_status_retry` (`status`, `next_retry_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Outbox消息表';

CREATE TABLE `assets` (
    `id` BIGINT NOT NULL COMMENT '资产ID(Snowflake)',
    `tenant_id` VARCHAR(32) NOT NULL COMMENT '租户ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `item_code` VARCHAR(64) NOT NULL COMMENT '商品编码',
    `quantity` BIGINT NOT NULL DEFAULT 0 COMMENT '持有数量',
    `frozen` BIGINT NOT NULL DEFAULT 0 COMMENT '冻结数量',
    `total_granted` BIGINT NOT NULL DEFAULT 0 COMMENT '累计发放数量',
    `total_consumed` BIGINT NOT NULL DEFAULT 0 COMMENT '累计消耗数量',
    `version` INT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_item` (`tenant_id`, `user_id`, `item_code`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户资产余额表';

CREATE TABLE `asset_ledger` (
    `id` BIGINT NOT NULL COMMENT '流水ID(Snowflake)',
    `tenant_id` VARCHAR(32) NOT NULL COMMENT '租户ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `item_code` VARCHAR(64) NOT NULL COMMENT '商品编码',
    `change_type` VARCHAR(20) NOT NULL COMMENT '变更类型: GRANT/CONSUME/FROZEN/UNFROZEN/RECONCILE',
    `change_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '变更数量(正数增加/负数减少)',
    `balance_before` BIGINT NOT NULL DEFAULT 0 COMMENT '变更前余额',
    `balance_after` BIGINT NOT NULL DEFAULT 0 COMMENT '变更后余额',
    `source_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源订单号',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `trace_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '追踪ID',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_source_order` (`tenant_id`, `user_id`, `item_code`, `source_order_no`),
    KEY `idx_tenant_user_item` (`tenant_id`, `user_id`, `item_code`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资产流水明细表';

CREATE TABLE `notifications` (
    `id` BIGINT NOT NULL COMMENT '通知ID(Snowflake)',
    `tenant_id` VARCHAR(32) NOT NULL COMMENT '租户ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `biz_no` VARCHAR(64) NOT NULL COMMENT '业务编号(如订单号)',
    `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型: NOTICE_SEND',
    `channel` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '通知渠道: SMS/EMAIL/APP_PUSH',
    `status` VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '状态: PENDING/SENT/FAILED',
    `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '通知标题',
    `content` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '通知内容',
    `result` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '发送结果',
    `trace_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '追踪ID',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `sent_at` DATETIME(3) NULL COMMENT '发送成功时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_biz_event` (`tenant_id`, `biz_no`, `event_type`),
    KEY `idx_status` (`status`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';