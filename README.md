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
5. 在配置文件中，默认tftp根目录在/mnt/tftp目录下，将boot image从iso中提取出来，放置到tftp目录即可。
6. tftp目录中我提前放置了一个grub样例，/mnt/tftp/pxelinux.cfg
具体如何配置grub请自行上网搜索教程,如何从iso中提取boot image也是如此。
7. 如何从pxe启动也请自行上网搜索教程,如何制作应答文件也是如此。
8. pxelinux.0文件版本为4.5版本，无需单独从iso中提取也可直接使用。
9. 请注意:本工具自带一个dhcp服务，在同一个局域网中请勿配置两个同样的dhcp服务器，会引起冲突。

如有疑问，请联系我： xxxmailk@163.com
