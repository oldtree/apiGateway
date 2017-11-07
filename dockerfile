FROM centos:latest
COPY apiGateway /root/apiGateway/apiGateway
COPY config.json /root/apiGateway/config.json
COPY run.sh /root/apiGateway/run.sh
EXPOSE 80
EXPOSE 8080
WORKDIR /root/apiGateway
CMD ["chmod","-x","*"]
ENTRYPOINT [ "run.sh" ]
