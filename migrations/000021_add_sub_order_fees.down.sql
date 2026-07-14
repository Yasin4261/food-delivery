alter table sub_orders
    drop column if exists delivery_fee,
    drop column if exists commission;
