# This image contains everything needed to build and test online-upgrade
FROM debian:8.7

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        build-essential make wget curl \
        libmysqlclient-dev mysql-client vim sshpass \
        openssh-server sudo locales git ssh-client \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install Golang 1.7
RUN cd /tmp \
 && wget -q https://storage.googleapis.com/golang/go1.7.5.linux-amd64.tar.gz \
 && tar -C /usr/local -xzf /tmp/go1.7.5.linux-amd64.tar.gz \
 && rm /tmp/*
ENV PATH /usr/local/go/bin:$PATH

# configure locale
RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen \
 && /usr/sbin/locale-gen en_US.UTF-8 \
 && DEBIAN_FRONTEND=noninteractive dpkg-reconfigure locales \
 && update-locale LANG=en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8
ENV LANGUAGE en_US:en

# configure SSH
ENV SSHPASS p455w0rd
RUN mkdir -p /var/run/sshd
RUN echo "root:$SSHPASS" | /usr/sbin/chpasswd
RUN sed -i "s/PermitRootLogin without-password/PermitRootLogin yes/" /etc/ssh/sshd_config

# add user account with sudo access
RUN useradd -s /bin/bash -m memsql-user
RUN echo "memsql-user:$SSHPASS" | /usr/sbin/chpasswd
RUN echo "memsql-user ALL=NOPASSWD: ALL" >> /etc/sudoers

# Timezone
RUN echo America/Los_Angeles | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata

# Define versions
ENV OPS_VERSION 5.7.1
ENV MEMSQL_LICENSE <enterprise license-key>
ENV MEMSQL_VERSION_5_5_12 28865cc3ab58b8ea8f4b0a6710fc87be6c8f1fb5
ENV MEMSQL_VERSION_5_7_2 03e5e3581e96d65caa30756f191323437a3840f0

# Install MemSQL Ops
RUN cd /tmp \
 && wget -q http://download.memsql.com/memsql-ops-$OPS_VERSION/memsql-ops-$OPS_VERSION.tar.gz \
 && tar -xzf memsql-ops-$OPS_VERSION.tar.gz \
 && ./memsql-ops-$OPS_VERSION/install.sh \
    --no-cluster --no-start --management-user memsql-user \
 && rm -rf /tmp/*

# The following blob does the following:
#  - installs an enterprise license
#  - deploys/deletes a MemSQL node for each version
RUN memsql-ops start \
 && memsql-ops license-add --license-key $MEMSQL_LICENSE \
 && memsql-ops memsql-deploy --version-hash $MEMSQL_VERSION_5_5_12 \
 && memsql-ops memsql-deploy --port 3307 --version-hash $MEMSQL_VERSION_5_7_2 \
 && memsql-ops memsql-delete --all --delete-without-prompting \
 && memsql-ops stop

# Make sure the next time Ops starts it has a unique agent id
RUN memsql-ops sqlite /var/lib/memsql-ops/data/topology.db -e "delete from agents" \
 && memsql-ops sqlite /var/lib/memsql-ops/data/variables.db -e "delete from variables"

# Code environment
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p /go/src/github.com/memsql/online-upgrade
RUN go get github.com/Masterminds/glide

CMD /usr/sbin/sshd && bash