FROM centos:7

RUN yum install -y git && yum clean all && rm -rf /var/cache/yum

COPY stproxy /opt/stproxy

EXPOSE 80

ENTRYPOINT stproxy