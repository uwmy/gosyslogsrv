# gosyslogsrv
2023-08-11  0.0.1 beta uwmy

systemctl daemon-reload
setcat 'cap_net_bind_service,cap_net_raw=ep' /usr/local/sbin/gosyslogsrv

/etc/passwd
gosyslogsrv:x:1000:1000:,,,:/home/gosyslogsrv:/usr/sbin/nologin

/etc/shadow
gosyslogsrv:!:19585:0:99999:7:::

/etc/systemd/system/gosyslogsrv.service 
[Unit]
Description=GO SysLog Server

[Service]
Type=simple

ExecStart=/usr/local/sbin/gosyslogsrv -d /home/gosyslogsrv
Restart=always

WorkingDirectory=/home/gosyslogsrv

User=gosyslogsrv

User=gosyslogsrv

[Install]
WantedBy=multi-user.target
