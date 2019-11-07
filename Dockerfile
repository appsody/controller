FROM busybox

COPY ./package/appsody-controller /appsody-controller
RUN  chmod +x /appsody-controller
WORKDIR /