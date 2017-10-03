FROM concourse/buildroot:git

ADD assets/ /opt/resource/
RUN chmod 755 /opt/resource/*
