-- varchar(15) can't hold formatted numbers like "+90 555 111 22 33" (17
-- chars); E.164 is up to 15 digits before the '+' and any spacing. 20 covers
-- a fully formatted Turkish number.
alter table users alter column phone_number type varchar(20);
