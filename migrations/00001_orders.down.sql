drop table if exists orders.order_items;
drop table if exists orders.items;
drop index if exists orders.idx_delivery_order_id;
drop table if exists orders.delivery;
drop index if exists orders.idx_orders_customer_id;
drop table if exists orders.orders;
drop schema if exists orders;