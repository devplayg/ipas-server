use ipasm;

select * from log_ipas_event;
select * from log_ipas_status;
select * from ast_ipas;

truncate table log_ipas_event;
truncate table log_ipas_status;
truncate table ast_ipas;