FROM centos:7

RUN rm /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN yum install -y git && yum clean all && rm -rf /var/cache/yum

COPY stproxy /opt/stproxy

EXPOSE 80

ENTRYPOINT /opt/stproxy
