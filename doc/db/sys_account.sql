use axis_account;

CREATE TABLE `sys_currency` (
`id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
`class`         VARCHAR(20)     NOT NULL                COMMENT '资产大类: crypto/fiat/stock/commodity',
`symbol`        VARCHAR(32)     NOT NULL                COMMENT '符号: CNY/USD/USDT/BTC/ETH',
`name`          VARCHAR(128)    NOT NULL                COMMENT '全称: 人民币/Tether USD/Bitcoin',
`decimals`      TINYINT UNSIGNED NOT NULL DEFAULT 0     COMMENT '最小单位精度: USD=2/JPY=0/ETH=18',
`smallest_unit` VARCHAR(20)     NOT NULL                COMMENT '最小单位名称: cent/分/円/wei/satoshi',
`status`        TINYINT         NOT NULL DEFAULT 1      COMMENT '状态: 1=启用 0=停用',
`sort`          INT             NOT NULL DEFAULT 0      COMMENT '排序权重',
`created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
UNIQUE KEY `uk_symbol` (`symbol`),
KEY `idx_class` (`class`),
KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='币种/资产元数据表';

INSERT INTO `sys_currency` (`class`, `symbol`, `name`, `decimals`, `smallest_unit`, `status`, `sort`) VALUES
-- ============ 法币 (Fiat) ============
('fiat', 'USD', 'US Dollar',              2, 'cent',    1, 100),
('fiat', 'CNY', 'Chinese Yuan',           2, '分',      1, 101),
('fiat', 'EUR', 'Euro',                   2, 'cent',    1, 102),
('fiat', 'GBP', 'British Pound',          2, 'penny',   1, 103),
('fiat', 'JPY', 'Japanese Yen',           0, '円',      1, 104),
('fiat', 'KRW', 'South Korean Won',       0, '원',      1, 105),
('fiat', 'HKD', 'Hong Kong Dollar',       2, 'cent',    1, 106),
('fiat', 'SGD', 'Singapore Dollar',       2, 'cent',    1, 107),
('fiat', 'AUD', 'Australian Dollar',      2, 'cent',    1, 108),
('fiat', 'CAD', 'Canadian Dollar',        2, 'cent',    1, 109),
('fiat', 'CHF', 'Swiss Franc',            2, 'rappen',  1, 110),
('fiat', 'INR', 'Indian Rupee',           2, 'paisa',   1, 111),
('fiat', 'BRL', 'Brazilian Real',         2, 'centavo', 1, 112),
('fiat', 'RUB', 'Russian Ruble',          2, 'kopek',   1, 113),
('fiat', 'BHD', 'Bahraini Dinar',         3, 'fils',    1, 114),
('fiat', 'KWD', 'Kuwaiti Dinar',          3, 'fils',    1, 115),
('fiat', 'TWD', 'New Taiwan Dollar',      2, 'cent',    1, 116),
('fiat', 'MOP', 'Macanese Pataca',        2, 'avo',     1, 117),

-- ============ 加密货币 (Crypto) ============
('crypto', 'BTC',  'Bitcoin',             8,  'satoshi', 1, 200),
('crypto', 'ETH',  'Ethereum',            18, 'wei',     1, 201),
('crypto', 'BNB',  'BNB',                 18, 'wei',     1, 202),
('crypto', 'SOL',  'Solana',              9,  'lamport', 1, 203),
('crypto', 'XRP',  'XRP',                 6,  'drop',    1, 204),
('crypto', 'ADA',  'Cardano',             6,  'lovelace',1, 205),
('crypto', 'DOGE', 'Dogecoin',            8,  'koinu',   1, 206),
('crypto', 'TRX',  'TRON',                6,  'sun',     1, 207),
('crypto', 'TON',  'Toncoin',             9,  'nanoton', 1, 208),
('crypto', 'AVAX', 'Avalanche',           18, 'wei',     1, 209),
('crypto', 'DOT',  'Polkadot',            10, 'planck',  1, 210),
('crypto', 'MATIC','Polygon',             18, 'wei',     1, 211),
('crypto', 'LTC',  'Litecoin',            8,  'litoshi', 1, 212),
('crypto', 'BCH',  'Bitcoin Cash',        8,  'satoshi', 1, 213),
('crypto', 'LINK', 'Chainlink',           18, 'wei',     1, 214),
('crypto', 'UNI',  'Uniswap',             18, 'wei',     1, 215),
('crypto', 'ATOM', 'Cosmos',              6,  'uatom',   1, 216),
('crypto', 'XLM',  'Stellar',             7,  'stroop',  1, 217),

-- ============ 稳定币 (Stablecoin) ============
('crypto', 'USDT', 'Tether USD',          6,  'micro',   1, 250),
('crypto', 'USDC', 'USD Coin',            6,  'micro',   1, 251),
('crypto', 'DAI',  'Dai Stablecoin',      18, 'wei',     1, 252),
('crypto', 'BUSD', 'Binance USD',         18, 'wei',     1, 253),
('crypto', 'TUSD', 'TrueUSD',             18, 'wei',     1, 254),
('crypto', 'FDUSD','First Digital USD',   18, 'wei',     1, 255),

-- ============ 股票 (Stock) ============
('stock', 'AAPL',  'Apple Inc.',          4, 'share1e-4', 1, 300),
('stock', 'TSLA',  'Tesla Inc.',          4, 'share1e-4', 1, 301),
('stock', 'BABA',  'Alibaba Group',       4, 'share1e-4', 1, 302),
('stock', 'GOOGL', 'Alphabet Inc.',       4, 'share1e-4', 1, 303),
('stock', 'MSFT',  'Microsoft Corp.',     4, 'share1e-4', 1, 304),
('stock', 'NVDA',  'NVIDIA Corp.',        4, 'share1e-4', 1, 305),
('stock', 'AMZN',  'Amazon.com Inc.',     4, 'share1e-4', 1, 306),
('stock', 'META',  'Meta Platforms Inc.', 4, 'share1e-4', 1, 307),
('stock', 'SPACEX','SpaceX (private)',    4, 'share1e-4', 1, 308),
('stock', '0700',  'Tencent Holdings',    4, 'share1e-4', 1, 309),
('stock', '9988',  'Alibaba HK',          4, 'share1e-4', 1, 310),

-- ============ 大宗商品 (Commodity) ============
('commodity', 'XAU', 'Gold (troy ounce)',       4, 'mg',       1, 400),
('commodity', 'XAG', 'Silver (troy ounce)',     4, 'mg',       1, 401),
('commodity', 'XPT', 'Platinum',                4, 'mg',       1, 402),
('commodity', 'XPD', 'Palladium',               4, 'mg',       1, 403),
('commodity', 'OIL', 'Crude Oil (barrel)',      4, 'ml',       1, 404);
-- ('commodity', 'GAS', 'Natural Gas (MMBtu)',     4, 'btu1e-4',  1, 405),
-- ('commodity', 'WHEAT','Wheat (bushel)',         4, 'grain1e-4',1, 406),
-- ('commodity', 'CORN','Corn (bushel)',           4, 'grain1e-4',1, 407),
-- ('commodity', 'RE-HZ','Real Estate - Hangzhou', 2, 'cent',     1, 408),
-- ('commodity', 'EV-TSLA-M3','Tesla Model 3 (used)', 2, 'cent', 1, 409);