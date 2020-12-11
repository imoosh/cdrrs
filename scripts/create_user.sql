
create user 'centnet_cdrrs'@'%' identified by '123456';

create user 'centnet_cdrrs'@'localhost' identified by '123456';

grant all on *.* to 'centnet_cdrrs'@'localhost';

grant all on *.* to 'centnet_cdrrs'@'%';

flush privileges;