SET NAMES utf8mb4;

INSERT INTO shop_category (name, icon, sort) VALUES
('火锅', 'hotpot.png', 1),
('咖啡', 'coffee.png', 2),
('烧烤', 'bbq.png', 3);

INSERT INTO shop (name, category_id, address, lng, lat, score, avg_price, description) VALUES
('蜀香火锅', 1, '上海市浦东新区世纪大道100号', 121.506, 31.245, 4.7, 120, '川味火锅，环境热闹'),
('Manner Coffee', 2, '上海市黄浦区南京东路200号', 121.490, 31.240, 4.5, 28, '平价精品咖啡'),
('深夜烧烤', 3, '上海市徐汇区漕溪北路300号', 121.436, 31.188, 4.6, 85, '适合夜宵聚会');

INSERT INTO user (phone, nickname, password_hash) VALUES
('13800000001', 'alice', ''),
('13800000002', 'bob', '');

INSERT INTO review (user_id, shop_id, content, like_count) VALUES
(1, 1, '味道不错，牛肉很嫩。', 3),
(2, 1, '排队有点久，但值得。', 5),
(1, 2, '咖啡香气很好，出杯快。', 2);

INSERT INTO voucher (shop_id, stock, begin_time, end_time) VALUES
(1, 5, NOW() - INTERVAL 1 HOUR, NOW() + INTERVAL 1 DAY),
(2, 3, NOW() - INTERVAL 1 HOUR, NOW() + INTERVAL 1 DAY);