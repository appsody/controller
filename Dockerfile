FROM busybox
ARG TARGETPLATFORM
COPY ./package/appsody-controller-$TARGETPLATFORM /appsody-controller
RUN  chmod +x /appsody-controller
WORKDIR /
CMD ["cp","/appsody-controller","/.appsody/appsody-controller"]