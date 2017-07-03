FROM centos:7

LABEL Release="1.0.0-beta" Vendor="Crunchy Data Solutions"

RUN yum -y install openssh-clients  hostname && yum clean all -y

ADD build/crunchy-proxy /usr/bin

VOLUME /config

EXPOSE 5432

USER daemon

CMD ["crunchy-proxy","start", "--config=/config/config.yaml" ]

