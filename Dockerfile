FROM busybox
COPY ./setController.sh /setController.sh
COPY ./package/appsody-controller /appsody-controller
RUN chmod +x /setController.sh /appsody-controller
WORKDIR /