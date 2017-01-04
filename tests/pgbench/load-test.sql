select /* read */ now();
/* read */ select now();
insert into proxytest values (1, 'foo', 'bar');
select /* read */ count(*) from proxytest;
update proxytest set name = 'doo' where id = 1;
delete from proxytest where id = 1;
insert into proxytest values (2, 'aaa', 'bbb');
/* read */ select value from proxytest;
delete from proxytest;
