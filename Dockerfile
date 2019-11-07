FROM busybox

COPY ./package/appsody-controller /appsody-controller
RUN  chmod +x /appsody-controller
WORKDIR /
CMD ["cp","/appsody-controller","/.appsody/appsody-controller"]