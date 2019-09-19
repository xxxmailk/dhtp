build:
	@go build .

pack:
	@tar czvf dhtp.tar.gz Makefile dhtp dhtp.yml dhtp.service tftp_files


install:
	@install -m 755 dhtp /usr/bin
	@install -D -m 644 dhtp.yml /etc/dhtp/dhtp.yml
	@install -d /mnt/dhtp/http /mnt/dhtp/http
	@install -m 644 dhtp.service /lib/systemd/system/
	@install -m 644 tftp_files/menu.c32 /mnt/dhtp/tftp
	@install -m 644 tftp_files/pxelinux.0 /mnt/dhtp/tftp
	@install -d tftp_files/pxelinux.cfg /mnt/dhtp/tftp/pxelinux.cfg
	@install -m 644 tftp_files/pxelinux.cfg/default /mnt/dhtp/tftp/pxelinux.cfg/default
	@echo "install successfully."
