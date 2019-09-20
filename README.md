# dhtp
a simple dhcp+http+tftp server for pxe deployment and packaged in one binary file, powered by golang.

一个简单的由dhcp+http+tftp组合的pxe远程部署工具。

#### 下载二进制包

[点击下载](https://raw.githubusercontent.com/xxxmailk/dhtp/master/dhtp.tar.gz)

[Download](https://raw.githubusercontent.com/xxxmailk/dhtp/master/dhtp.tar.gz)

#### 安装部署
1. 下载二进制包并解压到某个目录

2. 进入该目录并执行安装
    ```bash
    make install
    ```
    
3. 编辑配置文件，配置文件说明如下：
    ```bash
    # http 服务配置
    http:
    
      # http服务监听地址 偷懒可写0.0.0.0
      listen_ip: 0.0.0.0
    
      # http服务监听端口 默认80
      listemn_port: 80
    
      # http服务根目录地址，随便配置，注意权限, 至少644
      # 不行就chmod -R 777 /mnt/dhtp/http
      # 如果还是不行，那就检查一下防火墙和selinux是不是关闭了
      mount_path: /mnt/dhtp/http
    
    # tftp_files 服务配置
    tftp:
    
      # tftp_files 服务器根目录，配置原理同http服务
      mount_path: /mnt/dhtp/tftp
    
      # tftp服务监听地址 偷懒可写0.0.0.0
      listen_ip: 0.0.0.0
    
    # dhcp 服务配置
    dhcp:
    
      # 注意：
      # dhcp 服务器监听地址，必须写为你想用来提供pxe安装服务的那个网卡的地址
      # 不可以偷懒写成0.0.0.0
      listen_ip: 192.168.181.134
    
      # dhcp规定端口，不要改，详情参见RFC2132
      # https://tools.ietf.org/html/rfc2132
      listen_port: 67
    
      # 起始IP地址
      # 你想从哪个地址开始分配IP
      # 必须和你用来提供dhcp服务的那个网卡同网段
      # 否则pxe client无法访问tftp服务器
      start_ip: 192.168.181.135
    
      # 你想给多少个IP地址来给客户端使用，一般不超过253个
      # 如果你把子网掩码配的非常大，当我没说
      lease_range: 10
    
      # dhcp分配给客户端的子网掩码
      # 是的，你可以整个大段给客户端用
      netmask: 255.255.255.0
    
      # 这个地址填上面的Listen_ip就行
      # next server address -> siaddr: 参见RFC2132
      tftp_server: 192.168.181.134
    
      # pxe boot的文件名，一般不改，非要个性化也随你
      pxe_file: pxelinux.0
    ```
    
4. 启动dhtp服务
    ```bash
    systemctl start dhtp
    ```
    
5. 到此，你就可以开始进行pxe配置了。
    > tftp目录： /mnt/dhtp/tftp

    > http目录： /mnt/dhtp/http

    > 只需要按照pxe规则将linux iso中的相关内容放置到http目录下，并修改/mnt/tftp/pxelinux.cfg/default中的grub启动条目，你就能尝试正确地从pxe开始现状操作系统了。
6. 将iso挂载到某个目录
	```bash
	mount xxx.iso /mnt/centos_iso
	```
7. 在/mnt/dhtp/http中创建对应的iso文件目录并拷贝所有iso中的文件至该目录，比如安装suselinux，就在http目录下创建一个suse文件夹并将所有iso文件拷贝到该目录即可。
	```bash
	mkdir /mnt/dhtp/http/centos7.4
	cp -r /mnt/centos_iso/* /mnt/dhtp/http/centos7.4
	```
8. 在tftp目录中新建centos7.4目录，并将iso中的images/pxeboot/ 目录下的: initrd.img, vmlinuz 这两个文件拷贝到目录中
	```bash
	mkdir /mnt/dhtp/tftp/centos7.4
	cp /mnt/centos_iso/images/pxeboot/* /mnt/dhtp/tftp/centos7.4
	```
	
9. 创建kickstart文件
	```bash
	# cat /mnt/dhtp/http/centos7.4/kickstart
	#version=DEVEL
	auth --enableshadow --passalgo=sha512 #密码加密方式
url --url=http://192.168.181.134/centos7.4 #使用什么方式去引导启动
	install #安装
	text #命令模式安装  可以选择
	reboot #安装完，自动重启
	selinux --disabled #关闭SElinux
	firewall --disabled #关闭防火墙
	
	firstboot --enable #初始化开启
	ignoredisk --only-use=sda #选择磁盘  sda
	
	keyboard --vckeymap=us --xlayouts='us' #语言和键盘选择
	
	lang en_US.UTF-8 #文字选择
	
	network  --bootproto=dhcp --device=ens33 --onboot=off --ipv6=auto --activ
	ate #网卡设置
	network  --hostname=localhost.localdomain #主机名设置
	
	rootpw --iscrypted  xxxxxxx #设置密码  为加密文本
	
	services --disabled="chronyd"
	
	timezone Asia/Shanghai --isUtc --nontp  #时区选择
	
	bootloader --append=" crashkernel=auto" --location=mbr --boot-drive=sda #分区引导
	zerombr  #清除分区
	clearpart --all --initlabel #清空磁盘
	
	#磁盘分区
	part pv.198 --fstype="lvmpv" --ondisk=sda --size=10240
	part /boot --fstype="xfs" --ondisk=sda --size=1024
	volgroup centos --pesize=4096 pv.198
	logvol /  --fstype="xfs" --size=5120 --name=root --vgname=centos
	logvol swap  --fstype="swap" --size=2048 --name=swap --vgname=centos
	logvol /app  --fstype="xfs" --size=2048 --name=app --vgname=centos
	eula --agreed #同意选项  centos7中必备
	
	%packages #安装的包
	@^minimal
	@core
	kexec-tools
	vim-enhanced
	%end
	
	%post #安装后运行脚本
	rm -f /etc/yum.repos.d/*
	cat > /etc/yum.repos.d/base.repo  <<EOF
	[base]
	name=basemage
	baseurl=http://192.168.181.134/centos7.4
	gpgcheck=0
	EOF
	useradd ylx
	echo 123456 |passwd --stdin ylx &> /dev/null
	
	%end
	```
	> 当然，该文件也可以通过工具生成： system-config-kickstart
9. 编辑/mnt/tftp/pxelinux.cfg/default配置启动条目
	```bash
	#cat /mnt/tftp/pxelinux.cfg/default
	default menu.c32 
    prompt 
    timeout 30
    
    menu title Auto Install CentOS 7 PXE Menu
    
    label dhtp pxe
    kernel centos7.4/vmlinuz
 append initrd=centos7.4/initrd.img ks=http://192.168.163.20/ksdir/ks7-mini.cfg
	
	 label local  
	 #磁盘启动 ；不安装系统时，该项设置成默认选项，不然重复安装系统
	 menu label Boot from ^local drive
	 ocalboot 0xffff
	
	 menu end
	
	```
> 上面的grub配置文件中，initrd指定启动镜像位置，该镜像放在tftp中， kernel则是指定启动内核位置，也放置在tftp中。

接下来，就可以启动服务器，等待自动从pxe引导，或者按F12进入pxe安装操作系统了。
